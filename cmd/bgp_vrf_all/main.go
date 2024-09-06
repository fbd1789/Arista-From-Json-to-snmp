package main

/*
BaseOID: 1.3.6.1.4.1.30065.4.226

+-- bgpVrfAll(1)
|   |
|   +-- bgpVrfAllTable(1)
|   |   |
|   |   +-- bgpVrfAllEntry(1)
|	|	|   | Index: [address][maskLen][nextHop][vrfName]
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
	"net/netip"
	"time"

	"github.com/rs/zerolog/log"

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

// func intsToString(a []int, sep string) string {
// 	sliced := make([]string, len(a))
// 	for i, val := range a {
// 		sliced[i] = strconv.Itoa(val)
// 	}
// 	return strings.Join(sliced, sep)
// }

// func loadMockFile(d any, path string) error {
// 	file, err := os.Open(path)
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()
// 	raw, _ := io.ReadAll(file)
// 	return json.Unmarshal(raw, &d)
// }

func main() {

	//passpersist.BaseOid, _ = passpersist.MustNewOid(passpersist.AristaExperimentalMib).Append([]int{226})
	// passpersist.EnableSyslogLogger("info", syslog.LOG_LOCAL4, "bgp_vrfs_all")
	// uncomment for debugging
	//passpersist.EnableConsoleLogger("debug")
	// passpersist.RefreshInterval = 180 * time.Second
	oid := passpersist.MustNewOid(passpersist.AristaExperimentalMib).MustAppend([]int{226})
	cfg := passpersist.MustNewConfig(
		passpersist.WithBaseOid(oid),
		passpersist.WithRefreshInterval(time.Second),
	)
	pp := passpersist.NewPassPersist(cfg)
	ctx := context.Background()
	pp.Run(ctx, func(pp *passpersist.PassPersist) {
		var data ShowBgpVrfAll
		//err := loadMockFile(&data, "bgp_vrfs_all.json")
		err := arista.EosCommandJson("show ip bgp vrf all", &data)
		if err != nil {
			log.Error().Msgf("failed to read data: %s", err)
			return
		}

		for name, vrf := range data.Vrfs {
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

					// fmt.Printf("%s %s %s\n", name, prefix.String(), path.NextHop.String())
					//pp.AddString(append([]int{1, 1, 1}, idx...), intsToString(idx, "."))
					pp.AddString(append([]int{1, 1, 1}, idx...), name)
					pp.AddIP(append([]int{1, 1, 2}, idx...), prefix.Addr())
					pp.AddInt(append([]int{1, 1, 3}, idx...), int32(prefix.Bits()))
					pp.AddIP(append([]int{1, 1, 4}, idx...), path.NextHop.Addr)
				}
			}
		}
	})
}
