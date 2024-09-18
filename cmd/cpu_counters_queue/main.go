package main

/*
+--- cpuCountersQueue (11)
    |
    +--- egressQueuesSummaryTable(1)
    |   |
    |   +--- egressQueuesSummaryEntry(1)
    |       | Index: portId, destTypeId, queueId
    |       |
    |       +-- String port(1)
    |       |
    |       +-- String destType(2)
    |       |
    |       +-- Integer queue(3)
    |       |
    |       +-- Counter64 enqueuedPackets(4)
    |       |
    |       +-- Counter64 enqueuedBytes(5)
    |       |
    |       +-- Counter64 droppedPackets(6)
    |       |
    |       +-- Counter64 droppedBytes(7)
    |
    +--- ingressVoqsSummaryTable(2)
    |   |
    |   +--- ingressVoqsSummaryEntry(1)
    |       | Index: CoppClassId
    |       |
    |       +-- String coppClass(1)
    |       |
    |       +-- Counter64 enqueuedPackets(2)
    |       |
    |       +-- Counter64 enqueuedBytes(3)
    |       |
    |       +-- Counter64 droppedPackets(4)
    |       |
    |       +-- Counter64 droppedBytes(5)
*/

import (
	"context"
	"log/slog"
	"log/syslog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/arista-northwest/go-passpersist/passpersist"
	"github.com/arista-northwest/go-passpersist/utils/arista"
)

// var mockData []byte = []byte(`{
//     "ingressVoqs": {
//         "sources": {
//             "all": {
//                 "cpuClasses": {
//                     "CoppSystemAclLog": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 0,
//                                 "enqueuedBytes": 0,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemArpInspect": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 0,
//                                 "enqueuedBytes": 0,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemBfd": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 5148989859,
//                                 "enqueuedBytes": 411917216190,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemBgp": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 55720589,
//                                 "enqueuedBytes": 4945362015,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemBpdu": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 568901,
//                                 "enqueuedBytes": 69974823,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemCfm": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 0,
//                                 "enqueuedBytes": 0,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemCvx": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 0,
//                                 "enqueuedBytes": 0,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemCvxHeartbeat": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 0,
//                                 "enqueuedBytes": 0,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemDot1xMba": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 0,
//                                 "enqueuedBytes": 0,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemEgressTrap": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 46,
//                                 "enqueuedBytes": 421084,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemIgmp": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 0,
//                                 "enqueuedBytes": 0,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemIpBcast": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 0,
//                                 "enqueuedBytes": 0,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemIpUcast": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 5806,
//                                 "enqueuedBytes": 763382,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemL2Ucast": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 433,
//                                 "enqueuedBytes": 28272,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemL3DstMiss": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 35069,
//                                 "enqueuedBytes": 2455974,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemL3LpmOverflow": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 21891682,
//                                 "enqueuedBytes": 7221330550,
//                                 "droppedPackets": 666721257115,
//                                 "droppedBytes": 289471421816285
//                             }
//                         }
//                     },
//                     "CoppSystemL3Ttl1IpOptUcast": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 21,
//                                 "enqueuedBytes": 15855,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemLacp": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 13292468,
//                                 "enqueuedBytes": 2233190162,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemLdp": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 0,
//                                 "enqueuedBytes": 0,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemLldp": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 465123,
//                                 "enqueuedBytes": 107461750,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemMirroring": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 0,
//                                 "enqueuedBytes": 0,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemMlag": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 0,
//                                 "enqueuedBytes": 0,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemMplsArpSuppress": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 0,
//                                 "enqueuedBytes": 0,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemMplsLabelMiss": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 287967773,
//                                 "enqueuedBytes": 24189250782,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemMplsTtl": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 52,
//                                 "enqueuedBytes": 6084,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemMvrp": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 0,
//                                 "enqueuedBytes": 0,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemOspfIsisUcast": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 0,
//                                 "enqueuedBytes": 0,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemPtp": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 0,
//                                 "enqueuedBytes": 0,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemRsvp": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 0,
//                                 "enqueuedBytes": 0,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemSflow": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 377781439,
//                                 "enqueuedBytes": 147345806718,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemDefault": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 0,
//                                 "enqueuedBytes": 0,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemCfmSnoop": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 0,
//                                 "enqueuedBytes": 0,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemIpMcast": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 0,
//                                 "enqueuedBytes": 0,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemIpMcastMiss": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 0,
//                                 "enqueuedBytes": 0,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemL2Bcast": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 635,
//                                 "enqueuedBytes": 50286,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemL3Ttl1IpOptions": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 94,
//                                 "enqueuedBytes": 17136,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemLinkLocal": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 23042,
//                                 "enqueuedBytes": 2731208,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemMulticastSnoop": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 0,
//                                 "enqueuedBytes": 0,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemOspfIsis": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 27373738,
//                                 "enqueuedBytes": 32012217128,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemPtpSnoop": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 0,
//                                 "enqueuedBytes": 0,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemVxlanEncap": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 0,
//                                 "enqueuedBytes": 0,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     },
//                     "CoppSystemVxlanVtepLearn": {
//                         "ports": {
//                             "": {
//                                 "enqueuedPackets": 0,
//                                 "enqueuedBytes": 0,
//                                 "droppedPackets": 0,
//                                 "droppedBytes": 0
//                             }
//                         }
//                     }
//                 }
//             }
//         }
//     },
//     "egressQueues": {
//         "sources": {
//             "all": {
//                 "cpuPorts": {
//                     "CpuTm0": {
//                         "ucastQueues": {
//                             "queues": {
//                                 "0": {
//                                     "enqueuedPackets": 5934124280,
//                                     "enqueuedBytes": 630109626676,
//                                     "droppedPackets": 0,
//                                     "droppedBytes": 0
//                                 }
//                             }
//                         },
//                         "mcastQueues": {
//                             "queues": {
//                                 "0": {
//                                     "enqueuedPackets": 0,
//                                     "enqueuedBytes": 0,
//                                     "droppedPackets": 0,
//                                     "droppedBytes": 0
//                                 }
//                             }
//                         }
//                     },
//                     "CpuTm1": {
//                         "ucastQueues": {
//                             "queues": {
//                                 "0": {
//                                     "enqueuedPackets": 0,
//                                     "enqueuedBytes": 0,
//                                     "droppedPackets": 0,
//                                     "droppedBytes": 0
//                                 }
//                             }
//                         },
//                         "mcastQueues": {
//                             "queues": {
//                                 "0": {
//                                     "enqueuedPackets": 0,
//                                     "enqueuedBytes": 0,
//                                     "droppedPackets": 0,
//                                     "droppedBytes": 0
//                                 }
//                             }
//                         }
//                     },
//                     "CpuTm2": {
//                         "ucastQueues": {
//                             "queues": {
//                                 "0": {
//                                     "enqueuedPackets": 0,
//                                     "enqueuedBytes": 0,
//                                     "droppedPackets": 0,
//                                     "droppedBytes": 0
//                                 }
//                             }
//                         },
//                         "mcastQueues": {
//                             "queues": {
//                                 "0": {
//                                     "enqueuedPackets": 0,
//                                     "enqueuedBytes": 0,
//                                     "droppedPackets": 0,
//                                     "droppedBytes": 0
//                                 }
//                             }
//                         }
//                     },
//                     "CpuTm3": {
//                         "ucastQueues": {
//                             "queues": {
//                                 "0": {
//                                     "enqueuedPackets": 0,
//                                     "enqueuedBytes": 0,
//                                     "droppedPackets": 0,
//                                     "droppedBytes": 0
//                                 }
//                             }
//                         },
//                         "mcastQueues": {
//                             "queues": {
//                                 "0": {
//                                     "enqueuedPackets": 0,
//                                     "enqueuedBytes": 0,
//                                     "droppedPackets": 0,
//                                     "droppedBytes": 0
//                                 }
//                             }
//                         }
//                     },
//                     "CpuTm4": {
//                         "ucastQueues": {
//                             "queues": {
//                                 "0": {
//                                     "enqueuedPackets": 0,
//                                     "enqueuedBytes": 0,
//                                     "droppedPackets": 0,
//                                     "droppedBytes": 0
//                                 }
//                             }
//                         },
//                         "mcastQueues": {
//                             "queues": {
//                                 "0": {
//                                     "enqueuedPackets": 0,
//                                     "enqueuedBytes": 0,
//                                     "droppedPackets": 0,
//                                     "droppedBytes": 0
//                                 }
//                             }
//                         }
//                     },
//                     "CpuTm5": {
//                         "ucastQueues": {
//                             "queues": {
//                                 "0": {
//                                     "enqueuedPackets": 0,
//                                     "enqueuedBytes": 0,
//                                     "droppedPackets": 0,
//                                     "droppedBytes": 0
//                                 }
//                             }
//                         },
//                         "mcastQueues": {
//                             "queues": {
//                                 "0": {
//                                     "enqueuedPackets": 0,
//                                     "enqueuedBytes": 0,
//                                     "droppedPackets": 0,
//                                     "droppedBytes": 0
//                                 }
//                             }
//                         }
//                     }
//                 }
//             }
//         }
//     }
// }`)

var coppClassMap map[string]int = map[string]int{
	"CoppSystemAclLog":           1,
	"CoppSystemArpInspect":       2,
	"CoppSystemBfd":              3,
	"CoppSystemBgp":              4,
	"CoppSystemBpdu":             5,
	"CoppSystemCfm":              6,
	"CoppSystemCvx":              7,
	"CoppSystemCvxHeartbeat":     8,
	"CoppSystemDot1xMba":         9,
	"CoppSystemEgressTrap":       10,
	"CoppSystemIgmp":             11,
	"CoppSystemIpBcast":          12,
	"CoppSystemIpUcast":          13,
	"CoppSystemL2Ucast":          14,
	"CoppSystemL3DstMiss":        15,
	"CoppSystemL3LpmOverflow":    16,
	"CoppSystemL3Ttl1IpOptUcast": 17,
	"CoppSystemLacp":             18,
	"CoppSystemLdp":              19,
	"CoppSystemLldp":             20,
	"CoppSystemMirroring":        21,
	"CoppSystemMlag":             22,
	"CoppSystemMplsArpSuppress":  23,
	"CoppSystemMplsLabelMiss":    24,
	"CoppSystemMplsTtl":          25,
	"CoppSystemMvrp":             26,
	"CoppSystemOspfIsisUcast":    27,
	"CoppSystemPtp":              28,
	"CoppSystemRsvp":             29,
	"CoppSystemSflow":            30,
	"CoppSystemDefault":          31,
	"CoppSystemCfmSnoop":         32,
	"CoppSystemIpMcast":          33,
	"CoppSystemIpMcastMiss":      34,
	"CoppSystemL2Bcast":          35,
	"CoppSystemL3Ttl1IpOptions":  36,
	"CoppSystemLinkLocal":        37,
	"CoppSystemMulticastSnoop":   38,
	"CoppSystemOspfIsis":         39,
	"CoppSystemPtpSnoop":         40,
	"CoppSystemVxlanEncap":       41,
	"CoppSystemVxlanVtepLearn":   42,
}

var destTypeMap map[string]int = map[string]int{
	"ucastQueues": 0,
	"mcastQueues": 1,
}

type CpuCountersQueueSummary struct {
	IngressVoqs  IngressVoqs
	EgressQueues EgressQueues
}

type IngressVoqs struct {
	Sources map[string]IngressSource
}

type IngressSource struct {
	CpuClasses map[string]CpuClass
}

type CpuClass struct {
	Ports map[string]IngressCounters
}

type IngressCounters struct {
	EnqueuedPackets int
	EnqueuedBytes   int
	DroppedPackets  int
	DroppedBytes    int
}

type EgressQueues struct {
	Sources map[string]EgressSource
}

type EgressSource struct {
	CpuPorts map[string]map[string]CpuPortQueues
}

type CpuPortQueues struct {
	Queues map[string]CpuPortQueueCounters
}

type CpuPortQueueCounters struct {
	EnqueuedPackets int
	EnqueuedBytes   int
	DroppedPackets  int
	DroppedBytes    int
}

func init() {
	w, _ := syslog.New(syslog.LOG_LOCAL4, filepath.Base(os.Args[0]))
	l := slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(l)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var data CpuCountersQueueSummary

	var opts []passpersist.ConfigFunc

	b, _ := arista.GetBaseOIDFromSnmpConfig()
	if b != nil {
		opts = append(opts, passpersist.WithBaseOID(*b))
	}
	opts = append(opts, passpersist.WithRefreshInterval(time.Second*30))

	pp := passpersist.NewPassPersist(ctx, opts...)

	pp.Run(func(pp *passpersist.PassPersist) {
		// if err := json.Unmarshal(mockData, &data); err != nil {
		// 	panic(err)
		// }
		if err := arista.EosCommandJson("show cpu counters queue summary", &data); err != nil {
			slog.Warn("failed to run eos command", slog.Any("error", err)) //.Msgf("failed to read data: %s", err).Send()
			return
		}

		for port, destTypes := range data.EgressQueues.Sources["all"].CpuPorts {
			egressQueuesSummaryTable := []int{11, 1, 1}
			re := regexp.MustCompile(`CpuTm(\d+)`)

			portId, err := strconv.Atoi(re.FindStringSubmatch(port)[1])
			if err != nil {
				slog.Warn("failed to patse cpu port", slog.Any("error", err))
				continue
			}

			for destType, queues := range destTypes {
				if destTypeId, ok := destTypeMap[destType]; ok {
					for q, counters := range queues.Queues {
						queueId, _ := strconv.Atoi(q)
						//idx := []int{portId, destTypeId, queueId}

						pp.AddString(
							append(egressQueuesSummaryTable, 1, portId, destTypeId, queueId),
							port,
						)
						pp.AddString(
							append(egressQueuesSummaryTable, 2, portId, destTypeId, queueId),
							destType,
						)
						pp.AddInt(
							append(egressQueuesSummaryTable, 3, portId, destTypeId, queueId),
							int32(queueId),
						)
						pp.AddCounter64(
							append(egressQueuesSummaryTable, 4, portId, destTypeId, queueId),
							uint64(counters.EnqueuedPackets),
						)
						pp.AddCounter64(
							append(egressQueuesSummaryTable, 5, portId, destTypeId, queueId),
							uint64(counters.EnqueuedBytes),
						)
						pp.AddCounter64(
							append(egressQueuesSummaryTable, 6, portId, destTypeId, queueId),
							uint64(counters.DroppedPackets),
						)
						pp.AddCounter64(
							append(egressQueuesSummaryTable, 7, portId, destTypeId, queueId),
							uint64(counters.DroppedBytes),
						)
					}
				}
			}
		}
		for coppClass, cpuClass := range data.IngressVoqs.Sources["all"].CpuClasses {
			ingressVoqsSummaryTable := []int{11, 2, 1}
			counters := cpuClass.Ports[""]

			if coppClassIdx, ok := coppClassMap[coppClass]; ok {
				pp.AddString(append(ingressVoqsSummaryTable, 1, coppClassIdx), coppClass)
				pp.AddCounter64(append(ingressVoqsSummaryTable, 2, coppClassIdx), uint64(counters.EnqueuedPackets))
				pp.AddCounter64(append(ingressVoqsSummaryTable, 3, coppClassIdx), uint64(counters.EnqueuedBytes))
				pp.AddCounter64(append(ingressVoqsSummaryTable, 4, coppClassIdx), uint64(counters.DroppedPackets))
				pp.AddCounter64(append(ingressVoqsSummaryTable, 5, coppClassIdx), uint64(counters.DroppedBytes))
			}
		}
	})
}
