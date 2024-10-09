package main

import (
	"context"
	"log/slog"
	"log/syslog"
	"net/netip"
	"time"

	"github.com/arista-northwest/go-passpersist/passpersist"
	"github.com/arista-northwest/go-passpersist/utils"
	"github.com/arista-northwest/go-passpersist/utils/arista"
	"github.com/arista-northwest/go-passpersist/utils/logger"
)

var (
	date    string
	tag     string
	version string
)

type IPAddress struct {
	netip.Addr
}

func (a *IPAddress) UnmarshalJSON(b []byte) error {
	b = b[1 : len(b)-1]

	if len(b) == 0 {
		return nil
	}

	addr, err := netip.ParseAddr(string(b))
	if err != nil {
		return err
	}
	*a = IPAddress{addr}

	return nil
}

func (a IPAddress) split() []int {
	l := 4
	if a.Is6() {
		l = 16
	}
	s := make([]int, l)
	for i, b := range a.AsSlice() {
		s[i] = int(b)
	}

	return s
}

type ShowBGPSummaryVRFAll struct {
	VRFs map[string]VRF
}

type VRF struct {
	Name     string `json:"vrf"`
	RouterId IPAddress
	Asn      int32 `json:",string"`
	Peers    map[string]Peer
}

type Peer struct {
	PrefixReceived int
	PrefixAccepted int
	UpDownTime     float64
	State          string
}

func init() {
	logger.EnableSyslogger(syslog.LOG_LOCAL4, slog.LevelInfo)
	// redirectStderr(filepath.Base(os.Args[0]) + ".err")
}

// func redirectStderr(path string) {
// 	f, err := os.Create(path)
// 	if err != nil {
// 		slog.Error("failed to create file", "path", path, slog.Any("err", err))
// 		os.Exit(1)
// 	}
// 	err = syscall.Dup2(int(f.Fd()), int(os.Stderr.Fd()))
// 	if err != nil {
// 		slog.Error("Failed to redirect stderr to file: %v", slog.Any("err", err))
// 		os.Exit(1)
// 	}
// }

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// mock := flag.Bool("mock", false, "use mock data")
	utils.CommonCLI(version, tag, date)

	var opts []passpersist.Option

	b, _ := utils.GetBaseOIDFromSNMPdConfig()
	if b != nil {
		opts = append(opts, passpersist.WithBaseOID(*b))
	}

	opts = append(opts, passpersist.WithRefresh(time.Second*300))
	pp := passpersist.NewPassPersist(opts...)

	pp.Run(ctx, func(pp *passpersist.PassPersist) {
		var v4Data ShowBGPSummaryVRFAll
		// utils.MustLoadMockDataFile(&v4Data, "v4data.json")
		if err := arista.EosCommandJson("show ip bgp summary vrf all", &v4Data); err != nil {
			slog.Error("failed to read data", slog.Any("error", err))
			return
		}
		for _, vrf := range v4Data.VRFs {
			for s, peer := range vrf.Peers {
				addr := IPAddress{netip.MustParseAddr(s)}
				idx := addr.split()
				idx = append(idx, utils.EncodeString(vrf.Name)...)

				dur := time.Since(time.Unix(int64(peer.UpDownTime), 0))
				pp.AddString(append([]int{1, 1, 1, 1}, idx...), vrf.Name)
				pp.AddString(append([]int{1, 1, 1, 2}, idx...), addr.Addr.String())
				pp.AddInt(append([]int{1, 1, 1, 3}, idx...), vrf.Asn)
				pp.AddTimeTicks(append([]int{1, 1, 1, 4}, idx...), dur)
				pp.AddInt(append([]int{1, 1, 1, 5}, idx...), int32(peer.PrefixReceived))
				pp.AddInt(append([]int{1, 1, 1, 6}, idx...), int32(peer.PrefixAccepted))
			}
		}

		var v6Data ShowBGPSummaryVRFAll
		// utils.MustLoadMockDataFile(&v6Data, "v6data.json")
		if err := arista.EosCommandJson("show ipv6 bgp summary vrf all", &v6Data); err != nil {
			slog.Error("failed to read data", slog.Any("error", err))
			return
		}
		for _, vrf := range v6Data.VRFs {
			for s, peer := range vrf.Peers {
				addr := IPAddress{netip.MustParseAddr(s)}
				idx := addr.split()
				idx = append(idx, utils.EncodeString(vrf.Name)...)

				dur := time.Since(time.Unix(int64(peer.UpDownTime), 0))

				pp.AddString(append([]int{1, 2, 1, 1}, idx...), vrf.Name)
				pp.AddString(append([]int{1, 2, 1, 2}, idx...), addr.Addr.String())
				pp.AddInt(append([]int{1, 2, 1, 3}, idx...), vrf.Asn)
				pp.AddTimeTicks(append([]int{1, 2, 1, 4}, idx...), dur)
				pp.AddInt(append([]int{1, 1, 1, 5}, idx...), int32(peer.PrefixReceived))
				pp.AddInt(append([]int{1, 1, 1, 6}, idx...), int32(peer.PrefixAccepted))
			}
		}
	})
}
