package main

import (
	"context"
	"flag"
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

type BgpRoutePath struct {
	// "nextHop": "97.1.1.25"
	NextHop         *IPAddress `json:"nextHop"`
	LocalPreference int        `json:"localPreference"`
}

type BgpRouteEntry struct {
	Address       IPAddress
	MaskLength    int
	BgpRoutePaths []BgpRoutePath `json:"bgpRoutePaths"`
}

type Vrf struct {
	Name            string
	RouterId        IPAddress
	Asn             int32 `json:",string"`
	BgpRouteEntries map[string]BgpRouteEntry
}

type ShowBgpVrfAll struct {
	Vrfs map[string]Vrf
}

// func encodeString(s string) []int {
// 	b, _ := asn1.Marshal(s)
// 	oid := make([]int, len(b))
// 	for i, b := range b {
// 		oid[i] = int(b)
// 	}
// 	return oid
// }

func init() {
	logger.EnableSyslogger(syslog.LOG_LOCAL4, slog.LevelInfo)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mock := flag.Bool("mock", false, "use mock data")
	utils.CommonCLI(version, tag, date)

	var opts []passpersist.Option

	b, _ := utils.GetBaseOIDFromSNMPdConfig()
	if b != nil {
		opts = append(opts, passpersist.WithBaseOID(*b))
	}

	opts = append(opts, passpersist.WithRefresh(time.Second*30))
	pp := passpersist.NewPassPersist(opts...)

	pp.Run(ctx, func(pp *passpersist.PassPersist) {
		var v4Data ShowBgpVrfAll
		if *mock {
			utils.MustLoadMockDataFile(&v4Data, "v4_data.json")
		} else {
			if err := arista.EosCommandJson("show ip bgp vrf all", &v4Data); err != nil {
				slog.Error("failed to read data", slog.Any("error", err))
				return
			}
		}
		for name, vrf := range v4Data.Vrfs {
			for p, entry := range vrf.BgpRouteEntries {
				prefix := netip.MustParsePrefix(p)
				for _, path := range entry.BgpRoutePaths {
					if !path.NextHop.IsValid() {
						continue
					}

					idx := entry.Address.split()
					idx = append(idx, entry.MaskLength)
					idx = append(idx, path.NextHop.split()...)
					idx = append(idx, utils.EncodeString(name)...)

					pp.AddString(append([]int{1, 1, 1, 1}, idx...), name)
					pp.AddString(append([]int{1, 1, 1, 2}, idx...), prefix.Addr().String())
					pp.AddInt(append([]int{1, 1, 1, 3}, idx...), int32(prefix.Bits()))
					pp.AddString(append([]int{1, 1, 1, 4}, idx...), path.NextHop.Addr.String())
				}
			}
		}

		var v6Data ShowBgpVrfAll
		if *mock {
			utils.MustLoadMockDataFile(&v6Data, "v6_data.json")
		} else {
			if err := arista.EosCommandJson("show ipv6 bgp vrf all", &v6Data); err != nil {
				slog.Error("failed to read data", slog.Any("error", err))
			}
		}

		for name, vrf := range v6Data.Vrfs {
			for p, entry := range vrf.BgpRouteEntries {
				prefix := netip.MustParsePrefix(p)
				for _, path := range entry.BgpRoutePaths {
					if !path.NextHop.IsValid() {
						continue
					}

					idx := entry.Address.split()
					idx = append(idx, entry.MaskLength)
					idx = append(idx, path.NextHop.split()...)
					idx = append(idx, utils.EncodeString(name)...)

					pp.AddString(append([]int{1, 2, 1, 1}, idx...), name)
					pp.AddString(append([]int{1, 2, 1, 2}, idx...), prefix.Addr().String())
					pp.AddInt(append([]int{1, 2, 1, 3}, idx...), int32(prefix.Bits()))
					pp.AddString(append([]int{1, 2, 1, 4}, idx...), path.NextHop.Addr.String())
				}
			}
		}

	})
}
