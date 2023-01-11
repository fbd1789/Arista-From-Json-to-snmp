package main

import (
	"encoding/json"
	"fmt"

	//"os/exec"

	"github.com/go-cmd/cmd"
)

/*
RSVP-TE

Need : LSPResBWMegB, CSPFMetric, Metric

LSPResBWMegB = Bandwidth on the RSVP tunnel.

RWA01#show traffic-engineering rsvp tunnel | json

RWA01#show ip route 1.1.1.2 | json

MPLS-TE-STD-MIB::mplsTunnelName[0][2][16843009][16843010] = STRING: TU.rwa01.icr01

mplsTunnelIndex, mplsTunnelInstance, mplsTunnelIngressLSRId, mplsTunnelEgressLSRId
*/

var tunnels = []byte(`{
    "tunnels": {
        "TU.rwa01.icr01": {
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
                            "10.12.12.3"
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

var routes = []byte(`{
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
                                "tunnelName": "TU.rwa01.icr01",
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
	RequestedBandwidth int `json:"requestedBandwidth"`
	CurrentBandwidth   int `json:"currentBandwidth"`
	Metric             int `json:"metric"`
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
|	|	| Index: destination
|   |	|
|   |   +-- Integer currentBandwidth(1)
|   |   |
|   |   +-- Integer requestedBandwidth(2)
|   |   |
|   |   +-- Integer cspfMetric(2)
|   |   |
|   |   +-- Integer metric(3)
*/

func eosCommand(command string) []string {
	// TERM=dumb Cli -p 15 -c ''
	// name := "Cli"
	// args := append([]string{"-p", "15", "-c"}, tokens...)
	// cmd := exec.Command(name, args...)
	// cmd.Env = append(cmd.Env, "TERM=dumb")
	// err := cmd.Run()
	// if err != nil {
	// 	log.Fatalf("command failed: %s", err)
	// }

	c := cmd.NewCmd("Cli", "-p", "15", "-c", command)
	c.Env = append(c.Env, "TERM=dumb")
	<-c.Start()

	return c.Status().Stdout
}

func main() {
	var tun Tunnels
	var rtes Routes

	json.Unmarshal(tunnels, &tun)
	json.Unmarshal(routes, &rtes)
	// fmt.Printf("%v\n%v\n", tun, rtes)

	// fmt.Printf("%d", rtes.VRFs["default"].Routes["1.1.1.2/32"].Metric)

	tuns := eosCommand("show snmp mib walk MPLS-TE-STD-MIB::mplsTunnelName")

	fmt.Printf("TUNS: %s", tuns)
	// passpersist.Config.Refresh = time.Second * 5
	// pp := passpersist.NewPassPersist(&passpersist.Config)
	// ctx := context.Background()
	// pp.Run(ctx, func(pp *passpersist.PassPersist) {
	// 	pp.AddString([]int{255, 0}, "Hello")
	// 	pp.AddInt([]int{255, 1}, 42)
	// 	pp.AddString([]int{255, 2}, "!")
	// })
}
