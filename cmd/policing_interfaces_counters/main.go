package main

import (
	"context"
	"log/slog"
	"log/syslog"
	"time"

	"github.com/arista-northwest/go-passpersist/passpersist"
	"github.com/arista-northwest/go-passpersist/utils"
	"github.com/arista-northwest/go-passpersist/utils/arista"
	"github.com/arista-northwest/go-passpersist/utils/logger"
)

// var mock []byte = []byte(`{
//     "interfaces": {
//         "Ethernet48/1.1792": {
//             "profileName": "data-plane-policer",
//             "mode": "committed",
//             "counters": {
//                 "conformedPackets": 126089999,
//                 "conformedBytes": 128939946940,
//                 "yellowPackets": 0,
//                 "yellowBytes": 0,
//                 "exceededPackets": 0,
//                 "exceededBytes": 0,
//                 "droppedBitsRate": 0.0,
//                 "conformedBitsRate": 0.0,
//                 "exceededBitsRate": 1.4
//             }
//         },
// 		"Ethernet48/1.1758": {
//             "profileName": "data-plane-policer",
//             "mode": "committed",
//             "counters": {
//                 "conformedPackets": 125823053,
//                 "conformedBytes": 128666995001,
//                 "yellowPackets": 0,
//                 "yellowBytes": 0,
//                 "exceededPackets": 0,
//                 "exceededBytes": 0,
//                 "droppedBitsRate": 0.0,
//                 "conformedBitsRate": 0.0,
//                 "exceededBitsRate": 1.1
//             }
//         },
//         "Ethernet48/1.1588": {
//             "profileName": "data-plane-policer",
//             "mode": "committed",
//             "counters": {
//                 "conformedPackets": 125795389,
//                 "conformedBytes": 128638777298,
//                 "yellowPackets": 0,
//                 "yellowBytes": 0,
//                 "exceededPackets": 0,
//                 "exceededBytes": 0,
//                 "droppedBitsRate": 0.0,
//                 "conformedBitsRate": 0.0,
//                 "exceededBitsRate": 1.3
//             }
//         }
// 	},
// 	"egrInterfaces": {
//         "Ethernet34/1.1054": {
//             "profileName": "data-plane-policer",
//             "mode": "committed",
//             "counters": {
//                 "conformedPackets": 13661252,
//                 "conformedBytes": 14112073316,
//                 "yellowPackets": 0,
//                 "yellowBytes": 0,
//                 "exceededPackets": 0,
//                 "exceededBytes": 0,
//                 "droppedBitsRate": 0.0,
//                 "conformedBitsRate": 0.0,
//                 "exceededBitsRate": 0.0
//             }
//         },
// 		"Ethernet48/1.1713": {
//             "profileName": "data-plane-policer",
//             "mode": "committed",
//             "counters": {
//                 "conformedPackets": 195605440,
//                 "conformedBytes": 196648674880,
//                 "yellowPackets": 0,
//                 "yellowBytes": 0,
//                 "exceededPackets": 0,
//                 "exceededBytes": 0,
//                 "droppedBitsRate": 0.0,
//                 "conformedBitsRate": 0.0,
//                 "exceededBitsRate": 0.0
//             }
//         },
//         "Ethernet48/1.1837": {
//             "profileName": "data-plane-policer",
//             "mode": "committed",
//             "counters": {
//                 "conformedPackets": 195604320,
//                 "conformedBytes": 196647547040,
//                 "yellowPackets": 0,
//                 "yellowBytes": 0,
//                 "exceededPackets": 0,
//                 "exceededBytes": 0,
//                 "droppedBitsRate": 0.0,
//                 "conformedBitsRate": 0.0,
//                 "exceededBitsRate": 0.0
//             }
//         }
// 	},
//     "cpuInterfaces": {}
// }`)

// var mockIfIndexMap map[string]int = map[string]int{
// 	"Ethernet48/1.1792": 4811792,
// 	"Ethernet48/1.1758": 4811758,
// 	"Ethernet48/1.1588": 4811588,
// 	"Ethernet34/1.1054": 4811643,
// 	"Ethernet48/1.1713": 4811678,
// 	"Ethernet48/1.1837": 4811323,
// }

var (
	date    string = ""
	tag     string = ""
	version string = ""
)

type PolicingInterfaceCounters struct {
	Interfaces    map[string]Interface
	EgrInterfaces map[string]Interface
}

type Interface struct {
	ProfileName string
	Mode        string
	Counters    Counters
}

type Counters struct {
	ConformedPackets  uint64
	ConformedBytes    uint64
	YellowPackets     uint64
	YellowBytes       uint64
	ExceededPackets   uint64
	ExceededBytes     uint64
	DroppedBitsRate   float64
	ConformedBitsRate float64
	ExceededBitsRate  float64
}

func init() {
	logger.EnableSyslogger(syslog.LOG_LOCAL4, slog.LevelInfo)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	utils.CommonCLI(version, tag, date)

	data := &PolicingInterfaceCounters{}

	var opts []passpersist.Option

	b, _ := utils.GetBaseOIDFromSNMPdConfig()
	if b != nil {
		opts = append(opts, passpersist.WithBaseOID(*b))
	}
	opts = append(opts, passpersist.WithRefresh(time.Second*60))
	pp := passpersist.NewPassPersist(opts...)

	pp.Run(ctx, func(pp *passpersist.PassPersist) {
		// if err := json.Unmarshal(mock, &data); err != nil {
		// 	panic(err)
		// }
		// ifIndexMap := mockIfIndexMap

		if err := arista.EosCommandJson("show policing interfaces counters", &data); err != nil {
			slog.Error("failed to run eos command", slog.Any("error", err)) //.Msgf("failed to read data: %s", err).Send()
			return
		}
		ifIndexMap, err := arista.GetIfIndexeMap()
		if err != nil {
			slog.Warn("failed to update ifIndex map", slog.Any("error", err))
			return
		}

		for intf, profile := range data.Interfaces {
			idx, ok := ifIndexMap[intf]
			if !ok {
				slog.Warn("no index found", "interface", intf)
				continue
			}

			pp.AddString([]int{1, 1, 1, idx}, intf)
			pp.AddString([]int{1, 1, 2, idx}, profile.ProfileName)
			pp.AddString([]int{1, 1, 3, idx}, profile.Mode)
			pp.AddCounter64([]int{1, 1, 4, idx}, profile.Counters.ConformedPackets)
			pp.AddCounter64([]int{1, 1, 5, idx}, profile.Counters.ConformedBytes)
			pp.AddCounter64([]int{1, 1, 6, idx}, profile.Counters.YellowPackets)
			pp.AddCounter64([]int{1, 1, 7, idx}, profile.Counters.YellowBytes)
			pp.AddCounter64([]int{1, 1, 8, idx}, profile.Counters.ExceededPackets)
			pp.AddCounter64([]int{1, 1, 9, idx}, profile.Counters.ExceededBytes)
		}

		for intf, profile := range data.EgrInterfaces {
			idx, ok := ifIndexMap[intf]
			if !ok {
				slog.Warn("no index found", "interface", intf)
				continue
			}

			pp.AddString([]int{2, 1, 1, idx}, intf)
			pp.AddString([]int{2, 1, 2, idx}, profile.ProfileName)
			pp.AddString([]int{2, 1, 3, idx}, profile.Mode)
			pp.AddCounter64([]int{2, 1, 4, idx}, profile.Counters.ConformedPackets)
			pp.AddCounter64([]int{2, 1, 5, idx}, profile.Counters.ConformedBytes)
			pp.AddCounter64([]int{2, 1, 6, idx}, profile.Counters.YellowPackets)
			pp.AddCounter64([]int{2, 1, 7, idx}, profile.Counters.YellowBytes)
			pp.AddCounter64([]int{2, 1, 8, idx}, profile.Counters.ExceededPackets)
			pp.AddCounter64([]int{2, 1, 9, idx}, profile.Counters.ExceededBytes)
		}
	})
}

// func addEntries(name string, idx int, p Interface) {
// }
