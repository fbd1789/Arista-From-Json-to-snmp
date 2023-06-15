package main

import (
	"context"
	"fmt"
	"log/syslog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/arista-northwest/go-passpersist/passpersist"
	"github.com/rs/zerolog/log"
)

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
	EnqueuedBytes   uint64 `json:"enqueuedBytes"`
	EnqueuedPackets uint64 `json:"enqueuedPackets"`
	DroppedBytes    uint64 `json:"droppedBytes"`
	DroppedPackets  uint64 `json:"droppedPackets"`
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
	passpersist.BaseOid, _ = passpersist.MustNewOid(passpersist.AristaExperimentalMib).Append([]int{224})
	passpersist.EnableSyslogLogger("debug", syslog.LOG_LOCAL4, "intf_tc_queue_counters")
	passpersist.RefreshInterval = 60 * time.Second

	pp := passpersist.NewPassPersist()
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
					//pp.AddOctetString([]int{1, idx, tci}, []byte{idx, tci})
					pp.AddString([]int{1, idx, tci}, fmt.Sprintf("%d.%d", idx, tci))
					pp.AddString([]int{2, idx, tci}, strings.Join([]string{intf, tc}, ":"))
					pp.AddCounter64([]int{3, idx, tci}, counters.EnqueuedBytes)
					pp.AddCounter64([]int{4, idx, tci}, counters.EnqueuedPackets)
					pp.AddCounter64([]int{5, idx, tci}, counters.DroppedBytes)
					pp.AddCounter64([]int{6, idx, tci}, counters.DroppedPackets)
				}
			}

		}
	})
}
