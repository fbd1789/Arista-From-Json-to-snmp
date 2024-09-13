package main

/*
BaseOID: <derived>

+-- bgpVrfAll(1)
|   |
|   +-- ipbgpVrfAllTable(1)
|   |   |
|   |   +-- bgpVrfAllEntry(1)
|	|	|   | Index: [prefix][prefixLen][nextHop][vrfName]
|   |	|   |
|   |   |   +-- String vrfName(1)
|   |   |   |
|   |   |   +-- IpAddress prefix(2)
|   |   |   |
|   |   |   +-- Integer prefixLen int(3)
|   |   |   |
|   |   |   +-- IpAddress nextHop metric(4)
|   |   |
|   +-- ipv6bgpVrfAllTable(2)
|   |   |
|   |   +-- bgpVrfAllEntry(1)
|	|	|   | Index: [prefix][prefixLen][nextHop][vrfName]
|   |	|   |
|   |   |   +-- String vrfName(1)
|   |   |   |
|   |   |   +-- IpAddress prefix(2)
|   |   |   |
|   |   |   +-- Integer prefixLen int(3)
|   |   |   |
|   |   |   +-- IpAddress nextHop metric(4)

*/

import (
	"context"
	"encoding/asn1"
	"log/syslog"
	"net/netip"
	"time"

	"github.com/arista-northwest/go-passpersist/passpersist"
	"github.com/arista-northwest/go-passpersist/utils/arista"
	"github.com/rs/zerolog/log"
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

// func intsToString(a []int, sep string) string {
// 	sliced := make([]string, len(a))
// 	for i, val := range a {
// 		sliced[i] = strconv.Itoa(val)
// 	}
// 	return strings.Join(sliced, sep)
// }

func main() {

	passpersist.EnableSyslogLogger("debug", syslog.LOG_LOCAL4, "ip_bgp_vrfs_all")

	// uncomment for debugging
	//passpersist.EnableConsoleLogger("debug")

	//basOid := passpersist.MustNewOid(passpersist.AristaExperimentalMib).MustAppend([]int{226})
	baseOid := arista.MustGetBaseOid()

	cfg := passpersist.MustNewConfig(
		passpersist.WithBaseOid(baseOid),
		passpersist.WithRefreshInterval(60*time.Second),
	)
	pp := passpersist.NewPassPersist(cfg)
	ctx := context.Background()
	pp.Run(ctx, func(pp *passpersist.PassPersist) {
		var v4Data ShowBgpVrfAll
		// utils.MustLoadMockDataFile(&v4Data, "v4_data.json")
		if err := arista.EosCommandJson("show ip bgp vrf all", &v4Data); err != nil {
			log.Error().Msgf("failed to read data: %s", err)
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
			log.Error().Msgf("failed to read data: %s", err)
			return
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
