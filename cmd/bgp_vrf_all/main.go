package main

import (
	"context"
	"encoding/asn1"
	"log/slog"
	"log/syslog"
	"net/netip"
	"os"
	"path/filepath"
	"time"

	"github.com/arista-northwest/go-passpersist/passpersist"
	"github.com/arista-northwest/go-passpersist/utils/arista"
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

func (a *IPAddress) split() []int {
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

func encodeString(s string) []int {
	b, _ := asn1.Marshal(s)
	oid := make([]int, len(b))
	for i, b := range b {
		oid[i] = int(b)
	}
	return oid
}

func init() {
	w, _ := syslog.New(syslog.LOG_LOCAL4, filepath.Base(os.Args[0]))
	l := slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(l)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var opts []passpersist.ConfigFunc
	b, _ := arista.GetBaseOidFromSnmpConfig()
	if b != nil {
		opts = append(opts, passpersist.WithBaseOid(*b))
	}

	opts = append(opts, passpersist.WithRefreshInterval(time.Second*30))
	pp := passpersist.NewPassPersist(ctx, opts...)

	pp.Run(func(pp *passpersist.PassPersist) {
		var v4Data ShowBgpVrfAll
		// utils.MustLoadMockDataFile(&v4Data, "v4_data.json")
		if err := arista.EosCommandJson("show ip bgp vrf all", &v4Data); err != nil {
			slog.Error("failed to read data", slog.Any("error", err))
			return
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
					idx = append(idx, encodeString(name)...)

					pp.AddString(append([]int{1, 1, 1, 1}, idx...), name)
					pp.AddIP(append([]int{1, 1, 1, 2}, idx...), prefix.Addr())
					pp.AddInt(append([]int{1, 1, 1, 3}, idx...), int32(prefix.Bits()))
					pp.AddIP(append([]int{1, 1, 1, 4}, idx...), path.NextHop.Addr)
				}
			}
		}

		var v6Data ShowBgpVrfAll
		// utils.MustLoadMockDataFile(&v6Data, "v6_data.json")
		if err := arista.EosCommandJson("show ipv6 bgp vrf all", &v6Data); err != nil {
			slog.Error("failed to read data", slog.Any("error", err))
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
					idx = append(idx, encodeString(name)...)

					pp.AddString(append([]int{1, 2, 1, 1}, idx...), name)
					pp.AddIPV6(append([]int{1, 2, 1, 2}, idx...), prefix.Addr())
					pp.AddInt(append([]int{1, 2, 1, 3}, idx...), int32(prefix.Bits()))
					pp.AddIPV6(append([]int{1, 2, 1, 4}, idx...), path.NextHop.Addr)
				}
			}
		}

	})
}

// func addData(pp *passpersist.PassPersist, table int, data ShowBgpVrfAll) error {

// 	for name, vrf := range data.Vrfs {
// 		for p, entry := range vrf.BgpRouteEntries {
// 			prefix := netip.MustParsePrefix(p)
// 			for _, path := range entry.BgpRoutePaths {
// 				if !path.NextHop.IsValid() {
// 					continue
// 				}

// 				idx := entry.Address.split()
// 				idx = append(idx, entry.MaskLength)
// 				idx = append(idx, path.NextHop.split()...)
// 				idx = append(idx, encodeString(name)...)

// 				// fmt.Printf("%s %s %s\n", name, prefix.String(), path.NextHop.String())
// 				//pp.AddString(append([]int{1, 1, 1}, idx...), intsToString(idx, "."))
// 				pp.AddString(append([]int{1, table, 1, 1}, idx...), name)
// 				pp.AddIPV6(append([]int{1, table, 1, 2}, idx...), prefix.Addr())
// 				pp.AddInt(append([]int{1, table, 1, 3}, idx...), int32(prefix.Bits()))
// 				pp.AddIPV6(append([]int{1, table, 1, 4}, idx...), path.NextHop.Addr)
// 			}
// 		}
// 	}
// 	return nil
// }
