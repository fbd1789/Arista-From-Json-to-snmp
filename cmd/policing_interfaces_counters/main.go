package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/arista-northwest/go-passpersist/passpersist"
	"github.com/arista-northwest/go-passpersist/utils/arista"
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
//     "cpuInterfaces": {}
// }`)

// var mockIfIndexMap map[string]int = map[string]int{
// 	"Ethernet48/1.1792": 4811792,
// 	"Ethernet48/1.1758": 4811758,
// 	"Ethernet48/1.1588": 4811588,
// }

type PolicingInterfaceCounters struct {
	Interfaces map[string]Interface
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

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	data := &PolicingInterfaceCounters{}

	//w, _ := syslog.New(syslog.LOG_LOCAL4, utils.ProgName())
	w := os.Stdout
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
			// pp.AddOctetString([]int{1, 1, 9, idx}, Float64ToBytes(profile.Counters.ConformedBitsRate))
			// pp.AddOctetString([]int{1, 1, 10, idx}, Float64ToBytes(profile.Counters.ExceededBitsRate))
		}
	})
}
