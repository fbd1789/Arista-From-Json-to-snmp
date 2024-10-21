package main

import (
	"context"
	"log/slog"
	"log/syslog"
	"time"
	// "strconv"
	"github.com/arista-northwest/go-passpersist/passpersist"
	"github.com/arista-northwest/go-passpersist/utils"
	"github.com/arista-northwest/go-passpersist/utils/arista"
	"github.com/arista-northwest/go-passpersist/utils/logger"
)

var (
	date    string
	tag     string
	version string
)

type Counters struct {
    Received  int64 `json:"received"`
    Forwarded int64 `json:"forwarded"`
    Dropped   int64 `json:"dropped"`
}

type InterfaceStats struct {
    Requests      Counters `json:"requests"`
    Replies       Counters `json:"replies"`
    LastResetTime float64  `json:"lastResetTime"`
}

type GlobalStats struct {
    AllRequests   Counters `json:"allRequests"`
    AllResponses  Counters `json:"allResponses"`
    LastResetTime float64  `json:"lastResetTime"`
}

type Data struct {
    GlobalCounters   GlobalStats                `json:"globalCounters"`
    InterfaceCounters map[string]InterfaceStats `json:"interfaceCounters"`
}

func init() {
	logger.EnableSyslogger(syslog.LOG_LOCAL4, slog.LevelInfo)
}

func main() {
	defer utils.CapPanic()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	utils.CommonCLI(version, tag, date)

	data := &Data{}

	var opts []passpersist.Option

	b, _ := utils.GetBaseOIDFromSNMPdConfig()
	if b != nil {
		opts = append(opts, passpersist.WithBaseOID(*b))
	}
	opts = append(opts, passpersist.WithRefresh(time.Second*300))

	pp := passpersist.NewPassPersist(opts...)

	pp.Run(ctx, func(pp *passpersist.PassPersist) {
		slog.Debug("show vrf...")
		if err := arista.EosCommandJson("show ip dhcp relay counters", &data); err != nil {
			slog.Error("failed to run eos command", slog.Any("error", err))
			return
		}
		index :=1
		for iface, stats := range data.InterfaceCounters{
			pp.AddString([]int{index}, iface)
			pp.AddCounter64([]int{index, 1}, uint64(stats.Requests.Received))
			pp.AddCounter64([]int{index, 2}, uint64(stats.Requests.Forwarded))
			pp.AddCounter64([]int{index, 3}, uint64(stats.Requests.Dropped))
			pp.AddCounter64([]int{index, 4}, uint64(stats.Replies.Received))
			pp.AddCounter64([]int{index, 5}, uint64(stats.Replies.Forwarded))
			pp.AddCounter64([]int{index, 6}, uint64(stats.Replies.Dropped)) 
			index++
		} 
		// pp.AddCounter64([]int{1, 1}, 34)
	})
}
   