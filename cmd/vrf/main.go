package main

import (
	"context"
	"log/slog"
	"log/syslog"
	"time"
	"strconv"
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

// Définition de la structure pour les protocoles (protocols) 
type Protocol struct {
	RoutingState  string `json:"routingState"`
	ProtocolState string `json:"protocolState"`
	Supported     bool   `json:"supported"`
}

// Définition de la structure pour les VRF (Virtual Routing and Forwarding)
type Vrf struct {
	RouteDistinguisher string              `json:"routeDistinguisher"`
	VrfState           string              `json:"vrfState"`
	InterfacesV6       []string            `json:"interfacesV6"`
	InterfacesV4       []string            `json:"interfacesV4"`
	Interfaces         []string            `json:"interfaces"`
	Protocols          map[string]Protocol `json:"protocols"`
}

// Définition de la structure principale (VRFs map)
type Vrfs struct {
	Vrfs map[string]Vrf `json:"vrfs"`
}

func init() {
	logger.EnableSyslogger(syslog.LOG_LOCAL4, slog.LevelInfo)
}

func main() {
	defer utils.CapPanic()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	utils.CommonCLI(version, tag, date)

	data := &Vrfs{}

	var opts []passpersist.Option

	b, _ := utils.GetBaseOIDFromSNMPdConfig()
	if b != nil {
		opts = append(opts, passpersist.WithBaseOID(*b))
	}
	opts = append(opts, passpersist.WithRefresh(time.Second*300))

	pp := passpersist.NewPassPersist(opts...)

	pp.Run(ctx, func(pp *passpersist.PassPersist) {
		slog.Debug("show vrf...")
		if err := arista.EosCommandJson("show vrf", &data); err != nil {
			slog.Error("failed to run eos command", slog.Any("error", err))
			return
		}
		index :=10
		for vrfName, vrfData := range data.Vrfs{
			pp.AddString([]int{index}, vrfName)
			pp.AddString([]int{index, 1}, vrfData.RouteDistinguisher)
			pp.AddString([]int{index, 2}, vrfData.VrfState)
			for protoName, protoData := range vrfData.Protocols{
				pp.AddString([]int{index, 3}, protoName)
				pp.AddString([]int{index, 3, 1}, protoData.RoutingState)
				pp.AddString([]int{index, 3, 2}, protoData.ProtocolState)
				pp.AddString([]int{index, 3, 3}, strconv.FormatBool(protoData.Supported))
			}
			index++
		} 
		// pp.AddCounter64([]int{1, 1}, 34)
	})
}
   