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
|   |   |   +-- IpAddress address(2)
|   |   |   |
|   |   |   +-- Integer maskLen int(3)
|   |   |   |
|   |   |   +-- IpAddress nextHop metric(4)

*/

import (
	"bytes"
	"context"
	"encoding/asn1"
	"encoding/json"
	"fmt"
	"log/syslog"
	"net/netip"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/arista-northwest/go-passpersist/passpersist"
	"github.com/go-cmd/cmd"
)

func eosCommand(command string) ([]string, error) {
	c := cmd.NewCmd("Cli", "-p15", "-c", command)
	c.Env = append(c.Env, "TERM=dumb")
	<-c.Start()

	stderr := c.Status().Stderr
	if len(stderr) > 0 {
		return []string{}, fmt.Errorf("%s", strings.Join(stderr, "\n"))
	}

	return c.Status().Stdout, nil
}

func eosCommandJson(command string, v any) error {
	out, err := eosCommand(fmt.Sprintf("%s | json", command))
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	for _, l := range out {
		buf.WriteString(l)
	}

	return json.Unmarshal(buf.Bytes(), v)
}

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

	passpersist.BaseOid, _ = passpersist.MustNewOid(passpersist.AristaExperimentalMib).Append([]int{226})
	passpersist.EnableSyslogLogger("info", syslog.LOG_LOCAL4, "bgp_vrfs_all")
	// uncomment for debugging
	//passpersist.EnableConsoleLogger("debug")
	passpersist.RefreshInterval = 180 * time.Second

	pp := passpersist.NewPassPersist()
	ctx := context.Background()
	pp.Run(ctx, func(pp *passpersist.PassPersist) {
		var data ShowBgpVrfAll
		//err := loadMockFile(&data, "bgp_vrfs_all.json")
		err := eosCommandJson("show ip bgp vrf all", &data)
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
