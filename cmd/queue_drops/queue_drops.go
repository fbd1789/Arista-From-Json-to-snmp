package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/arista-northwest/go-passpersist/passpersist"
	"github.com/go-cmd/cmd"
	"github.com/rs/zerolog/log"
)

// "ingressVoqCounters": {
// 	"interfaces": {
// 		"Ethernet23/1": {
// 			"trafficClasses": {
// 				"TC6": {
// 					"droppedBytes": 0,
// 					"enqueuedPackets": 0,
// 					"enqueuedBytes": 0,
// 					"droppedPackets": 0
// 				}

// func init() {
// 	zerolog.SetGlobalLevel(zerolog.DebugLevel)
// }

type InterfaceQueueCounters struct {
	IngressVoqCounters Interfaces `json:"ingressVoqCounters"`
}

type Interfaces struct {
	Interface map[string]Interface `json:"interfaces"`
}

type Interface struct {
	TrafficClasses map[string]Counters `json:"trafficClasses"`
}

type Counters struct {
	EnqueuedBytes   int64 `json:"enqueuedBytes"`
	EnqueuedPackets int64 `json:"enqueuedPackets"`
	DroppedBytes    int64 `json:"droppedBytes"`
	DroppedPackets  int64 `json:"droppedPackets"`
}

func eosCommand(command string) []string {
	c := cmd.NewCmd("Cli", "-p", "15", "-c", command)
	c.Env = append(c.Env, "TERM=dumb")
	<-c.Start()
	return c.Status().Stdout
}

func eosCommandJson(cmd string, v any) any {
	out := eosCommand(fmt.Sprintf("%s | json", cmd))
	var buf bytes.Buffer
	for _, l := range out {
		buf.WriteString(l)
	}
	return json.Unmarshal(buf.Bytes(), v)
}

func getIfIndexeMap() map[string]int {
	indexes := make(map[string]int)
	out := eosCommand("show snmp mib walk IF-MIB::ifDescr")
	for _, l := range out {
		re := regexp.MustCompile(`IF-MIB::ifDescr\[(\d+)\] = STRING: ([^$]+)`)
		t := re.FindStringSubmatch(string(l))
		if t == nil {
			continue
		}
		idx, _ := strconv.Atoi(t[1])
		name := t[2]
		indexes[name] = idx
	}

	return indexes
}

func getTrafficClassIndex(s string) int {
	re := regexp.MustCompile(`TC(\d+)`)
	m := re.FindStringSubmatch(s)
	idx, _ := strconv.Atoi(m[1])
	return idx
}
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf := &passpersist.ConfigT{
		BaseOid:  passpersist.MustNewOid(passpersist.DEFAULT_BASE_OID),
		Refresh:  60 * time.Second,
		LogLevel: 5,
	}
	pp := passpersist.NewPassPersist(conf)
	pp.Run(ctx, func(pp *passpersist.PassPersist) {
		var data InterfaceQueueCounters

		idxs := getIfIndexeMap()

		eosCommandJson("show interfaces counters queue", &data)
		for intf, idx := range idxs {
			log.Debug().Msgf("updating interface %s@%d", intf, idx)
			if tcs, ok := data.IngressVoqCounters.Interface[intf]; ok {
				for tc, counters := range tcs.TrafficClasses {
					log.Debug().Msgf("updating interface '%s:%s'", intf, tc)
					tci := getTrafficClassIndex(tc)
					pp.AddString([]int{1, idx, tci}, strings.Join([]string{intf, tc}, ":"))
					pp.AddCounter64([]int{3, idx, tci}, counters.EnqueuedBytes)
					pp.AddCounter64([]int{4, idx, tci}, counters.EnqueuedPackets)
					pp.AddCounter64([]int{5, idx, tci}, counters.DroppedBytes)
					pp.AddCounter64([]int{6, idx, tci}, counters.DroppedPackets)
				}
			}

		}
	})
}

//DATA

var ifi = `IF-MIB::ifDescr[1001] = STRING: Ethernet1/1
IF-MIB::ifDescr[2001] = STRING: Ethernet2/1
IF-MIB::ifDescr[3001] = STRING: Ethernet3/1
IF-MIB::ifDescr[4001] = STRING: Ethernet4/1
IF-MIB::ifDescr[5001] = STRING: Ethernet5/1
IF-MIB::ifDescr[6001] = STRING: Ethernet6/1
IF-MIB::ifDescr[7001] = STRING: Ethernet7/1
IF-MIB::ifDescr[8001] = STRING: Ethernet8/1
IF-MIB::ifDescr[9001] = STRING: Ethernet9/1
IF-MIB::ifDescr[10001] = STRING: Ethernet10/1
IF-MIB::ifDescr[11001] = STRING: Ethernet11/1
IF-MIB::ifDescr[12001] = STRING: Ethernet12/1
IF-MIB::ifDescr[13001] = STRING: Ethernet13/1
IF-MIB::ifDescr[14001] = STRING: Ethernet14/1
IF-MIB::ifDescr[15001] = STRING: Ethernet15/1
IF-MIB::ifDescr[16001] = STRING: Ethernet16/1
IF-MIB::ifDescr[17001] = STRING: Ethernet17/1
IF-MIB::ifDescr[18001] = STRING: Ethernet18/1
IF-MIB::ifDescr[19001] = STRING: Ethernet19/1
IF-MIB::ifDescr[20001] = STRING: Ethernet20/1
IF-MIB::ifDescr[21001] = STRING: Ethernet21/1
IF-MIB::ifDescr[22001] = STRING: Ethernet22/1
IF-MIB::ifDescr[23001] = STRING: Ethernet23/1
IF-MIB::ifDescr[24001] = STRING: Ethernet24/1
IF-MIB::ifDescr[25001] = STRING: Ethernet25/1
IF-MIB::ifDescr[26001] = STRING: Ethernet26/1
IF-MIB::ifDescr[27001] = STRING: Ethernet27/1
IF-MIB::ifDescr[28001] = STRING: Ethernet28/1
IF-MIB::ifDescr[29001] = STRING: Ethernet29/1
IF-MIB::ifDescr[30001] = STRING: Ethernet30/1
IF-MIB::ifDescr[31001] = STRING: Ethernet31/1
IF-MIB::ifDescr[32001] = STRING: Ethernet32/1
IF-MIB::ifDescr[33001] = STRING: Ethernet33/1
IF-MIB::ifDescr[33003] = STRING: Ethernet33/3
IF-MIB::ifDescr[33005] = STRING: Ethernet33/5
IF-MIB::ifDescr[33006] = STRING: Ethernet33/6
IF-MIB::ifDescr[33007] = STRING: Ethernet33/7
IF-MIB::ifDescr[33008] = STRING: Ethernet33/8
IF-MIB::ifDescr[34001] = STRING: Ethernet34/1
IF-MIB::ifDescr[34003] = STRING: Ethernet34/3
IF-MIB::ifDescr[34004] = STRING: Ethernet34/4
IF-MIB::ifDescr[34005] = STRING: Ethernet34/5
IF-MIB::ifDescr[34006] = STRING: Ethernet34/6
IF-MIB::ifDescr[34007] = STRING: Ethernet34/7
IF-MIB::ifDescr[34008] = STRING: Ethernet34/8
IF-MIB::ifDescr[35001] = STRING: Ethernet35/1
IF-MIB::ifDescr[35003] = STRING: Ethernet35/3
IF-MIB::ifDescr[35005] = STRING: Ethernet35/5
IF-MIB::ifDescr[35006] = STRING: Ethernet35/6
IF-MIB::ifDescr[35007] = STRING: Ethernet35/7
IF-MIB::ifDescr[35008] = STRING: Ethernet35/8
IF-MIB::ifDescr[36001] = STRING: Ethernet36/1
IF-MIB::ifDescr[36003] = STRING: Ethernet36/3
IF-MIB::ifDescr[36004] = STRING: Ethernet36/4
IF-MIB::ifDescr[36005] = STRING: Ethernet36/5
IF-MIB::ifDescr[36006] = STRING: Ethernet36/6
IF-MIB::ifDescr[36007] = STRING: Ethernet36/7
IF-MIB::ifDescr[36008] = STRING: Ethernet36/8
IF-MIB::ifDescr[999001] = STRING: Management1
IF-MIB::ifDescr[1000100] = STRING: Port-Channel100
IF-MIB::ifDescr[1000101] = STRING: Port-Channel101
IF-MIB::ifDescr[1000103] = STRING: Port-Channel103
IF-MIB::ifDescr[1000104] = STRING: Port-Channel104
IF-MIB::ifDescr[1000401] = STRING: Port-Channel401
IF-MIB::ifDescr[1000500] = STRING: Port-Channel500
IF-MIB::ifDescr[2000001] = STRING: Vlan1
IF-MIB::ifDescr[5000000] = STRING: Loopback0
IF-MIB::ifDescr[5000001] = STRING: Loopback1
`

var iqd = []byte(`{
    "ingressVoqCounters": {
        "interfaces": {
            "Ethernet23/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet1/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 165,
                        "enqueuedBytes": 33000,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 6,
                        "enqueuedBytes": 564,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet35/5": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet33/8": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Port-Channel103": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet22/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Port-Channel101": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Port-Channel100": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 5075,
                        "enqueuedBytes": 696387,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 1,
                        "enqueuedBytes": 64,
                        "droppedPackets": 0
                    }
                }
            },
            "InternalRecirc0/2": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet13/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet32/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet36/3": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "InternalRecirc0/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet21/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet25/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 167,
                        "enqueuedBytes": 33567,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Port-Channel500": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 99610,
                        "enqueuedBytes": 10255786,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 343,
                        "enqueuedBytes": 41822,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet36/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet20/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet24/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet27/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 2125,
                        "enqueuedBytes": 464042,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Port-Channel401": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet26/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 7030,
                        "enqueuedBytes": 1126402,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 5832,
                        "enqueuedBytes": 393328,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet36/6": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "EventorFap0": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet35/8": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet17/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "L3FloodFap0.0": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet33/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Port-Channel104": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet29/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet16/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet18/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet35/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet31/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 87118,
                        "enqueuedBytes": 8752328,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 138637604,
                        "enqueuedBytes": 189240326914,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet36/4": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet19/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "FabricMcast3": {
                "trafficClasses": {
                    "TC6-7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "FabricMcast2": {
                "trafficClasses": {
                    "TC4-5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "FabricMcast1": {
                "trafficClasses": {
                    "TC2-3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "FabricMcast0": {
                "trafficClasses": {
                    "TC0-1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet8/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet34/6": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet4/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 95884,
                        "enqueuedBytes": 9599954,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 1,
                        "enqueuedBytes": 64,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 336,
                        "enqueuedBytes": 40928,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet14/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet33/6": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet33/7": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet10/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet33/5": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet36/7": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet33/3": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet36/5": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet15/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet34/5": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet34/4": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet34/7": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet11/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet34/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet34/3": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet28/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 2125,
                        "enqueuedBytes": 464042,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "L3FloodFap0.1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet9/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 2122,
                        "enqueuedBytes": 461530,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet35/6": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet35/7": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet34/8": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "InternalRecirc0/3": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet35/3": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet36/8": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet3/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 4295,
                        "enqueuedBytes": 744475,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 9,
                        "enqueuedBytes": 576,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 7,
                        "enqueuedBytes": 894,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet12/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "OlpFap0": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet30/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 2465,
                        "enqueuedBytes": 507958,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    }
                }
            },
            "Ethernet2/1": {
                "trafficClasses": {
                    "TC6": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC7": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 395,
                        "enqueuedBytes": 54077,
                        "droppedPackets": 0
                    },
                    "TC4": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC5": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC2": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC3": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 0,
                        "enqueuedBytes": 0,
                        "droppedPackets": 0
                    },
                    "TC0": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 5,
                        "enqueuedBytes": 320,
                        "droppedPackets": 0
                    },
                    "TC1": {
                        "droppedBytes": 0,
                        "enqueuedPackets": 3,
                        "enqueuedBytes": 192,
                        "droppedPackets": 0
                    }
                }
            }
        }
    },
    "egressQueueCounters": {
        "interfaces": {
            "Ethernet23/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet1/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 165,
                                    "enqueuedBytes": 31845,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 6,
                                    "enqueuedBytes": 564,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet35/5": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet33/8": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Port-Channel103": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC3,7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2,6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1,5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0,4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC3,7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2,6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1,5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0,4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet22/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Port-Channel101": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC3,7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2,6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1,5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0,4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC3,7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2,6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1,5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0,4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Port-Channel100": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC3,7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2,6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1,5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0,4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC3,7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2,6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1,5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 5075,
                                    "enqueuedBytes": 660862,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 1,
                                    "enqueuedBytes": 64,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0,4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "InternalRecirc0/2": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC0,4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2,6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1,5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3,7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC0,4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2,6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1,5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3,7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet13/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet32/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet36/3": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "InternalRecirc0/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC0,4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2,6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1,5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3,7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC0,4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2,6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1,5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3,7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet21/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet25/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 167,
                                    "enqueuedBytes": 32398,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Port-Channel500": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC3,7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2,6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1,5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0,4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC3,7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2,6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1,5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 99613,
                                    "enqueuedBytes": 9559068,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 343,
                                    "enqueuedBytes": 41822,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0,4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet36/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet20/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet24/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet27/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 2125,
                                    "enqueuedBytes": 449167,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Port-Channel401": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC3,7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2,6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1,5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0,4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC3,7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2,6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1,5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0,4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet26/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 7030,
                                    "enqueuedBytes": 1077192,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 5832,
                                    "enqueuedBytes": 393328,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet36/6": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "EventorFap0": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet35/8": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet17/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "L3FloodFap0.0": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC0,4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2,6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1,5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3,7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC0,4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2,6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1,5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3,7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet33/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Port-Channel104": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC3,7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2,6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1,5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0,4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC3,7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2,6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1,5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0,4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet29/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet16/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet18/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet35/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet31/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 87118,
                                    "enqueuedBytes": 8142502,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 138637604,
                                    "enqueuedBytes": 189240326914,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet36/4": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet19/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet8/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet34/6": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet4/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 95884,
                                    "enqueuedBytes": 8928766,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 1,
                                    "enqueuedBytes": 64,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 336,
                                    "enqueuedBytes": 40928,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet14/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet33/6": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet33/7": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet10/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet33/5": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet36/7": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet33/3": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet36/5": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet15/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet34/5": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet34/4": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet34/7": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet11/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet34/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet34/3": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet28/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 2125,
                                    "enqueuedBytes": 449167,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "L3FloodFap0.1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC0,4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2,6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1,5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3,7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC0,4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2,6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1,5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3,7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet9/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 2122,
                                    "enqueuedBytes": 446676,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet35/6": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet35/7": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet34/8": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "InternalRecirc0/3": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC0,4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2,6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1,5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3,7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC0,4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2,6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1,5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3,7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet35/3": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet36/8": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet3/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 4296,
                                    "enqueuedBytes": 714480,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 9,
                                    "enqueuedBytes": 576,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 7,
                                    "enqueuedBytes": 894,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet12/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "PeerLinkRecircFap0": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC0,4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2,6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1,5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3,7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC0,4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2,6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1,5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3,7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "OlpFap0": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC0,4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2,6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1,5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3,7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC0,4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2,6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1,5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3,7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet30/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 2465,
                                    "enqueuedBytes": 490703,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            },
            "Ethernet2/1": {
                "mcastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                },
                "ucastQueues": {
                    "trafficClasses": {
                        "TC6": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC7": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 395,
                                    "enqueuedBytes": 51337,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC4": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC5": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC2": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC3": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC0": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 5,
                                    "enqueuedBytes": 320,
                                    "droppedPackets": 0
                                }
                            }
                        },
                        "TC1": {
                            "dropPrecedences": {
                                "DP0-3": {
                                    "droppedBytes": 0,
                                    "enqueuedPackets": 3,
                                    "enqueuedBytes": 192,
                                    "droppedPackets": 0
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}`)
