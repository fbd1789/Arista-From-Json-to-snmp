package main

import (
	"context"
	"log/slog"
	"log/syslog"
	"net"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"time"

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

type MACAddress struct {
	net.HardwareAddr
}

func (m *MACAddress) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), "\"")
	if str == "" {
		return nil
	}
	mac, err := net.ParseMAC(str)
	if err != nil {
		return err
	}
	*m = MACAddress{mac}
	return nil
}

type EOSTimeTicks struct {
	time.Duration
}

func (t *EOSTimeTicks) UnmarshalJSON(b []byte) error {
	ticks, err := strconv.ParseFloat(string(b), 64)
	if err != nil {
		return err
	}
	tmp := time.Duration(time.Nanosecond * time.Duration(ticks*1_000_000_000))
	*t = EOSTimeTicks{tmp}
	return nil
}

type EOSTimestamp struct {
	time.Time
}

func (t *EOSTimestamp) UnmarshalJSON(b []byte) error {
	ts, err := strconv.ParseFloat(string(b), 64)
	if err != nil {
		return err
	}
	tmp := time.Unix(0, int64(ts*1_000_000_000))
	*t = EOSTimestamp{tmp}
	return nil
}

type ShowVersion struct {
	ImageFormatVersion   string       `json:"imageFormatVersion,omitempty"`
	CEOSToolsVersion     string       `json:"cEosToolsVersion,omitempty"`
	Uptime               EOSTimeTicks `json:"uptime,omitempty"`
	ModelName            string       `json:"modelName,omitempty"`
	KernelVersion        string       `json:"kernelVersion,omitempty"`
	InternalVersion      string       `json:"internalVersion,omitempty"`
	MemoryTotal          int          `json:"memTotal,omitempty"`
	Manufacturer         string       `json:"mfgName,omitempty"`
	SerialNumber         string       `json:"serialNumber,omitempty"`
	SystemMACAddress     MACAddress   `json:"systemMacAddress,omitempty"`
	BootupTimeStamp      EOSTimestamp `json:"bootupTimestamp,omitempty"`
	MemoryFree           int          `json:"memFree,omitempty"`
	Version              string       `json:"version,omitempty"`
	ConfiguredMACAddress MACAddress   `json:"configMacAddress,omitempty"`
	IsInternalVersion    bool         `json:"isIntlVersion,omitempty"`
	ImageOptimization    string       `json:"imageOptimization,omitempty"`
	InternalBuildID      string       `json:"internalBuildId,omitempty"`
	HardwareRevision     string       `json:"hardwareRevision,omitempty"`
	HardwareMACAddress   MACAddress   `json:"hwMacAddress,omitempty"`
	Architecture         string       `json:"architecture,omitempty"`
}

func init() {
	logger.EnableSyslogger(syslog.LOG_LOCAL4, slog.LevelInfo)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	utils.CommonCLI(version, tag, date)

	var opts []passpersist.Option
	baseOID, _ := utils.GetBaseOIDFromSNMPdConfig()
	if baseOID != nil {
		opts = append(opts, passpersist.WithBaseOID(*baseOID))
	}

	pp := passpersist.NewPassPersist(opts...)

	pp.Run(ctx, func(pp *passpersist.PassPersist) {
		var data ShowVersion
		err := arista.EosCommandJson("show version", data)
		if err != nil {
			slog.Error("failed to run command", slog.Any("error", err))
			os.Exit(1)
		}
		pp.AddString([]int{255, 1}, data.Version)
		pp.AddInt([]int{255, 2}, int32(data.MemoryFree))
		pp.AddCounter32([]int{255, 3}, uint32(4294967295))
		pp.AddCounter64([]int{255, 4}, uint64(18446744073709551615))
		pp.AddOID([]int{255, 5}, passpersist.MustNewOID("1.3.6.1.4.1.30065.4.224"))
		pp.AddOctetString([]int{255, 6}, []byte{'0', 'b', 'c', 'd'})
		pp.AddIP([]int{255, 7}, netip.MustParseAddr("1.2.3.4"))
		pp.AddIPV6([]int{255, 8}, netip.MustParseAddr("dead:beef:1:2:3::4"))
		pp.AddGauge([]int{255, 9}, uint32(4294967295))
		pp.AddTimeTicks([]int{255, 10}, data.Uptime.Duration)
	})
}
