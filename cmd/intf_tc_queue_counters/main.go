package main

import (
	"context"
	"fmt"
	"log/slog"
	"log/syslog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/arista-northwest/go-passpersist/passpersist"
	"github.com/arista-northwest/go-passpersist/utils/arista"
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

	// global settings
	//passpersist.BaseOID, _ = passpersist.MustNewOID(passpersist.AristaExperimentalMib).Append([]int{224})
	w, _ := syslog.New(syslog.LOG_LOCAL4, filepath.Base(os.Args[0]))
	l := slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(l)

	var opts []passpersist.ConfigFunc

	b, _ := arista.GetBaseOIDFromSnmpConfig()
	if b != nil {
		opts = append(opts, passpersist.WithBaseOID(*b))
	}
	opts = append(opts, passpersist.WithRefreshInterval(time.Second*60))

	pp := passpersist.NewPassPersist(ctx, opts...)

	pp.Run(func(pp *passpersist.PassPersist) {
		var data InterfaceQueueCounters
		idxs, err := arista.GetIfIndexeMap()
		if err != nil {
			slog.Warn("failed to get ifIndex map. data not refreshed")
			return
		}

		arista.EosCommandJson("show interfaces counters queue", &data)

		for intf, idx := range idxs {
			if tcs, ok := data.IngressVoqCounters.Interface[intf]; ok {
				for tc, counters := range tcs.TrafficClasses {
					slog.Debug("updating interface", "interfaces", intf, "traffic-class", tc)
					tci := getTrafficClassIndex(tc)
					pp.AddString([]int{1, 1, 1, idx, tci}, fmt.Sprintf("%d.%d", idx, tci))
					pp.AddString([]int{1, 1, 2, idx, tci}, strings.Join([]string{intf, tc}, ":"))
					pp.AddCounter64([]int{1, 1, 3, idx, tci}, counters.EnqueuedBytes)
					pp.AddCounter64([]int{1, 1, 4, idx, tci}, counters.EnqueuedPackets)
					pp.AddCounter64([]int{1, 1, 5, idx, tci}, counters.DroppedBytes)
					pp.AddCounter64([]int{1, 1, 6, idx, tci}, counters.DroppedPackets)
				}
			}

		}
	})
}
