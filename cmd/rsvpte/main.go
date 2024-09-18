package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strconv"
	"strings"

	//"os/exec"

	"github.com/arista-northwest/go-passpersist/passpersist"
)

/*
** INCOMPLETE **

RSVP-TE

Need : LSPResBWMegB, CSPFMetric, Metric

LSPResBWMegB = Bandwidth on the RSVP tunnel.

RWA01#show traffic-engineering rsvp tunnel | json

RWA01#show ip route 1.1.1.2 | json

MPLS-TE-STD-MIB::mplsTunnelName[0][2][16843009][16843010] = STRING: TU.rwa01.icr01

mplsTunnelIndex, mplsTunnelInstance, mplsTunnelIngressLSRId, mplsTunnelEgressLSRId
*/

var mockTunnelIndexNames = []string{"MPLS-TE-STD-MIB::mplsTunnelName[0][2][16843009][16843010] = STRING: TUN-CEOS1-CEOS2"}

var mockTunnels = []byte(`{
    "tunnels": {
        "TUN-CEOS1-CEOS2": {
            "autoBandwidth": true,
            "requestedBandwidth": 5000000,
            "signalBandwidth": true,
            "metric": 11008,
            "destination": "1.1.1.2",
            "mbb": false,
            "subTunnels": {
                "1": {
                    "sessionId": 3,
                    "activePathType": "primary",
                    "currentBandwidth": 5000000,
                    "secondaryPaths": [],
                    "lspCount": 1,
                    "primaryPath": {
                        "hops": [
                            "169.254.0.1"
                        ],
                        "state": "up",
                        "pathErrors": [],
                        "inUse": true,
                        "specName": "to-icr01",
                        "specType": "dynamic",
                        "mbb": false
                    },
                    "state": "up",
                    "mbb": false
                }
            },
            "source": "1.1.1.1",
            "state": "up",
            "splitTunnel": true,
            "currentBandwidth": 5000000,
            "splitTunnelParams": {
                "reductionDelay": 86400,
                "minBandwidth": 5000000,
                "quantum": 0,
                "maxBandwidth": 5000000000,
                "subTunnelLimit": 20
            },
            "lspCount": 1,
            "activePathType": "split"
        }
    }
}`)

var mockRoutes = []byte(`{
    "vrfs": {
        "default": {
            "routes": {
                "1.1.1.2/32": {
                    "kernelProgrammed": true,
                    "directlyConnected": false,
                    "routeAction": "forward",
                    "routeLeaked": false,
                    "vias": [
                        {
                            "tunnelDescriptor": {
                                "tunnelIndex": 1,
                                "tunnelType": "RSVP LER",
                                "tunnelName": "TUN-CEOS1-CEOS2",
                                "tunnelAddressFamily": "IPv4",
                                "tunnelEndPoint": "1.1.1.2/32"
                            },
                            "tunViaBackupVias": [],
                            "vias": [
                                {
                                    "resolvingTunnel": {
                                        "type": "RSVP LER SUB",
                                        "index": 1
                                    }
                                }
                            ]
                        }
                    ],
                    "metric": 11010,
                    "hardwareProgrammed": true,
                    "routeType": "ISISLevel1",
                    "preference": 115
                }
            },
            "allRoutesProgrammedKernel": true,
            "routingDisabled": false,
            "allRoutesProgrammedHardware": true,
            "defaultRouteState": "notSet"
        }
    }
}`)

type Tunnel struct {
	Destination        string `json:"destination"`
	RequestedBandwidth int    `json:"requestedBandwidth"`
	CurrentBandwidth   int    `json:"currentBandwidth"`
	Metric             int    `json:"metric"`
}

type Tunnels struct {
	Tunnels map[string]Tunnel `json:"tunnels"`
}

type Route struct {
	Metric int `json:"metric"`
}

type VRF struct {
	Routes map[string]Route `json:"routes"`
}

type Routes struct {
	VRFs map[string]VRF `json:"vrfs"`
}

/*

+-- rsvpTeTunnels(1)
|   |
|   +-- rsvpTeTunnelsTable(1)
|   |   |
|   |   +-- rsvpTeTunnelsEntry(1)
|	|	| Index: [][][][]
|   |	|
|   |   +-- Integer currentBandwidth(1)
|   |   |
|   |   +-- Integer requestedBandwidth(2)
|   |   |
|   |   +-- Integer cspfMetric(2)
|   |   |
|   |   +-- Integer metric(3)
*/

func parseTunnelIndexNames(out []string) (map[string]passpersist.OID, error) {
	// convert this:
	// "MPLS-TE-STD-MIB::mplsTunnelName[0][2][16843009][16843010] = STRING: TUN-CEOS1-CEOS2"
	// to:
	// map[TUN-CEOS1-CEOS2:0.0.2.16843009.16843010]

	tuns := make(map[string]passpersist.OID)

	for _, line := range out {
		idx := passpersist.MustNewOID(".")
		line = strings.Trim(line, "\n")

		re := regexp.MustCompile(`(?:\[(\d+)\])(?:\[(\d+)\])(?:\[(\d+)\])(?:\[(\d+)\])\s+=\s+STRING: ([^$]+)`)
		groups := re.FindStringSubmatch(line)

		if len(groups) == 0 {
			return nil, fmt.Errorf("failed to parse tunnel indexes from: %s", line)
		}

		name := groups[len(groups)-1]

		for _, grp := range groups[:len(groups)-1] {
			t, _ := strconv.Atoi(grp)
			idx, _ = idx.Append([]int{t})
		}

		tuns[name] = idx
	}
	return tuns, nil
}

func main() {
	// var tun Tunnels
	// var rtes Routes

	// tun = json.Unmarshal(tunnels, &tun)
	// json.Unmarshal(routes, &rtes)

	// fmt.Printf("%d", rtes.VRFs["default"].Routes["1.1.1.2/32"].Metric)
	// eosCommand("show snmp mib walk MPLS-TE-STD-MIB::mplsTunnelName")

	tunnels, err := parseTunnelIndexNames(mockTunnelIndexNames)
	if err != nil {
		slog.Error("failed to parse", slog.Any("error", err))
		os.Exit(1)
	}

	// //fmt.Printf("%+v\n", tunnels)
	// var tuns Tunnels
	// json.Unmarshal(mockTunnels, tuns)

	// var rtes Routes
	// json.Unmarshal(mockTunnels, tuns)

	for name, idx := range tunnels {
		tuns := &Tunnels{}
		rtes := &Routes{}

		json.Unmarshal(mockTunnels, tuns)
		json.Unmarshal(mockRoutes, rtes)

		t := tuns.Tunnels[name]

		currBw := t.CurrentBandwidth
		reqBw := t.RequestedBandwidth
		cspfMet := t.Metric
		metric := rtes.VRFs["default"].Routes[t.Destination].Metric
		fmt.Printf("%+v\n\n", t)
		fmt.Printf("%+v\n\n", rtes)
		fmt.Printf("%s %s %d %d %d %d\n", idx, name, currBw, reqBw, cspfMet, metric)
	}
	// passpersist.Config.Refresh = time.Second * 5
	// pp := passpersist.NewPassPersist(&passpersist.Config)
	// ctx := context.Background()
	// pp.Run(ctx, func(pp *passpersist.PassPersist) {

	// })
}
