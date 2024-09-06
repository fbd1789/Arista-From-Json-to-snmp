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
	"regexp"
	"strconv"
	"time"

	"github.com/arista-northwest/go-passpersist/passpersist"
	"github.com/arista-northwest/go-passpersist/utils/arista"
	"github.com/rs/zerolog/log"
)

var mockData []byte = []byte(`{
    "ingressVoqs": {
        "sources": {
            "all": {
                "cpuClasses": {
                    "CoppSystemAclLog": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 0,
                                "enqueuedBytes": 0,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemArpInspect": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 0,
                                "enqueuedBytes": 0,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemBfd": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 5148989859,
                                "enqueuedBytes": 411917216190,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemBgp": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 55720589,
                                "enqueuedBytes": 4945362015,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemBpdu": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 568901,
                                "enqueuedBytes": 69974823,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemCfm": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 0,
                                "enqueuedBytes": 0,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemCvx": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 0,
                                "enqueuedBytes": 0,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemCvxHeartbeat": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 0,
                                "enqueuedBytes": 0,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemDot1xMba": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 0,
                                "enqueuedBytes": 0,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemEgressTrap": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 46,
                                "enqueuedBytes": 421084,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemIgmp": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 0,
                                "enqueuedBytes": 0,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemIpBcast": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 0,
                                "enqueuedBytes": 0,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemIpUcast": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 5806,
                                "enqueuedBytes": 763382,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemL2Ucast": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 433,
                                "enqueuedBytes": 28272,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemL3DstMiss": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 35069,
                                "enqueuedBytes": 2455974,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemL3LpmOverflow": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 21891682,
                                "enqueuedBytes": 7221330550,
                                "droppedPackets": 666721257115,
                                "droppedBytes": 289471421816285
                            }
                        }
                    },
                    "CoppSystemL3Ttl1IpOptUcast": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 21,
                                "enqueuedBytes": 15855,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemLacp": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 13292468,
                                "enqueuedBytes": 2233190162,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemLdp": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 0,
                                "enqueuedBytes": 0,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemLldp": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 465123,
                                "enqueuedBytes": 107461750,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemMirroring": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 0,
                                "enqueuedBytes": 0,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemMlag": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 0,
                                "enqueuedBytes": 0,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemMplsArpSuppress": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 0,
                                "enqueuedBytes": 0,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemMplsLabelMiss": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 287967773,
                                "enqueuedBytes": 24189250782,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemMplsTtl": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 52,
                                "enqueuedBytes": 6084,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemMvrp": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 0,
                                "enqueuedBytes": 0,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemOspfIsisUcast": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 0,
                                "enqueuedBytes": 0,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemPtp": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 0,
                                "enqueuedBytes": 0,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemRsvp": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 0,
                                "enqueuedBytes": 0,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemSflow": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 377781439,
                                "enqueuedBytes": 147345806718,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemDefault": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 0,
                                "enqueuedBytes": 0,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemCfmSnoop": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 0,
                                "enqueuedBytes": 0,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemIpMcast": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 0,
                                "enqueuedBytes": 0,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemIpMcastMiss": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 0,
                                "enqueuedBytes": 0,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemL2Bcast": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 635,
                                "enqueuedBytes": 50286,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemL3Ttl1IpOptions": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 94,
                                "enqueuedBytes": 17136,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemLinkLocal": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 23042,
                                "enqueuedBytes": 2731208,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemMulticastSnoop": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 0,
                                "enqueuedBytes": 0,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemOspfIsis": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 27373738,
                                "enqueuedBytes": 32012217128,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemPtpSnoop": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 0,
                                "enqueuedBytes": 0,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemVxlanEncap": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 0,
                                "enqueuedBytes": 0,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    },
                    "CoppSystemVxlanVtepLearn": {
                        "ports": {
                            "": {
                                "enqueuedPackets": 0,
                                "enqueuedBytes": 0,
                                "droppedPackets": 0,
                                "droppedBytes": 0
                            }
                        }
                    }
                }
            }
        }
    },
    "egressQueues": {
        "sources": {
            "all": {
                "cpuPorts": {
                    "CpuTm0": {
                        "ucastQueues": {
                            "queues": {
                                "0": {
                                    "enqueuedPackets": 5934124280,
                                    "enqueuedBytes": 630109626676,
                                    "droppedPackets": 0,
                                    "droppedBytes": 0
                                }
                            }
                        },
                        "mcastQueues": {
                            "queues": {
                                "0": {
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0,
                                    "droppedBytes": 0
                                }
                            }
                        }
                    },
                    "CpuTm1": {
                        "ucastQueues": {
                            "queues": {
                                "0": {
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0,
                                    "droppedBytes": 0
                                }
                            }
                        },
                        "mcastQueues": {
                            "queues": {
                                "0": {
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0,
                                    "droppedBytes": 0
                                }
                            }
                        }
                    },
                    "CpuTm2": {
                        "ucastQueues": {
                            "queues": {
                                "0": {
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0,
                                    "droppedBytes": 0
                                }
                            }
                        },
                        "mcastQueues": {
                            "queues": {
                                "0": {
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0,
                                    "droppedBytes": 0
                                }
                            }
                        }
                    },
                    "CpuTm3": {
                        "ucastQueues": {
                            "queues": {
                                "0": {
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0,
                                    "droppedBytes": 0
                                }
                            }
                        },
                        "mcastQueues": {
                            "queues": {
                                "0": {
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0,
                                    "droppedBytes": 0
                                }
                            }
                        }
                    },
                    "CpuTm4": {
                        "ucastQueues": {
                            "queues": {
                                "0": {
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0,
                                    "droppedBytes": 0
                                }
                            }
                        },
                        "mcastQueues": {
                            "queues": {
                                "0": {
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0,
                                    "droppedBytes": 0
                                }
                            }
                        }
                    },
                    "CpuTm5": {
                        "ucastQueues": {
                            "queues": {
                                "0": {
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0,
                                    "droppedBytes": 0
                                }
                            }
                        },
                        "mcastQueues": {
                            "queues": {
                                "0": {
                                    "enqueuedPackets": 0,
                                    "enqueuedBytes": 0,
                                    "droppedPackets": 0,
                                    "droppedBytes": 0
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}`)

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

var queueTypeMap map[string]int = map[string]int{
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

func main() {
	var data CpuCountersQueueSummary
	// if err := json.Unmarshal(mockData, &data); err != nil {
	// 	panic(err)
	// }

	err := arista.EosCommandJson("show cpu counters queue summary", &data)
	if err != nil {
		log.Error().Msgf("failed to read data: %s", err)
		return
	}

	baseOid := passpersist.MustNewOid(passpersist.NetSnmpExtendMib).MustAppend([]int{5})

	cfg := passpersist.MustNewConfig(
		passpersist.WithBaseOid(baseOid),
		passpersist.WithRefreshInterval(time.Second),
	)
	pp := passpersist.NewPassPersist(cfg)
	ctx := context.Background()

	pp.Run(ctx, func(pp *passpersist.PassPersist) {

		for port, queueTypes := range data.EgressQueues.Sources["all"].CpuPorts {
			egressQueuesSummaryTable := passpersist.MustNewOid("1.1")
			re := regexp.MustCompile(`CpuTm(\d+)`)

			portId, err := strconv.Atoi(re.FindStringSubmatch(port)[1])
			if err != nil {
				// log error
				continue
			}

			for queueType, queues := range queueTypes {
				if queueTypeId, ok := queueTypeMap[queueType]; ok {
					for q, counters := range queues.Queues {
						queueId, _ := strconv.Atoi(q)
						idx := []int{portId, queueTypeId, queueId}

						pp.AddString(
							egressQueuesSummaryTable.MustAppend([]int{1}).MustAppend(idx).Value,
							port,
						)
						pp.AddString(
							egressQueuesSummaryTable.MustAppend([]int{2}).MustAppend(idx).Value,
							queueType,
						)
						pp.AddInt(
							egressQueuesSummaryTable.MustAppend([]int{3}).MustAppend(idx).Value,
							int32(queueId),
						)
						pp.AddCounter64(
							egressQueuesSummaryTable.MustAppend([]int{4}).MustAppend(idx).Value,
							uint64(counters.EnqueuedPackets),
						)
						pp.AddCounter64(
							egressQueuesSummaryTable.MustAppend([]int{5}).MustAppend(idx).Value,
							uint64(counters.EnqueuedBytes),
						)
						pp.AddCounter64(
							egressQueuesSummaryTable.MustAppend([]int{6}).MustAppend(idx).Value,
							uint64(counters.DroppedPackets),
						)
						pp.AddCounter64(
							egressQueuesSummaryTable.MustAppend([]int{7}).MustAppend(idx).Value,
							uint64(counters.DroppedBytes),
						)
					}
				}
			}
		}
		for coppClass, cpuClass := range data.IngressVoqs.Sources["all"].CpuClasses {
			ingressVoqsSummaryTable := passpersist.MustNewOid("2.1")
			counters := cpuClass.Ports[""]

			if coppClassId, ok := coppClassMap[coppClass]; ok {
				pp.AddString(ingressVoqsSummaryTable.MustAppend([]int{1, coppClassId}).Value, coppClass)
				pp.AddCounter64(ingressVoqsSummaryTable.MustAppend([]int{2, coppClassId}).Value, uint64(counters.EnqueuedPackets))
				pp.AddCounter64(ingressVoqsSummaryTable.MustAppend([]int{3, coppClassId}).Value, uint64(counters.EnqueuedBytes))
				pp.AddCounter64(ingressVoqsSummaryTable.MustAppend([]int{4, coppClassId}).Value, uint64(counters.DroppedPackets))
				pp.AddCounter64(ingressVoqsSummaryTable.MustAppend([]int{5, coppClassId}).Value, uint64(counters.DroppedBytes))
			}
		}
	})
}