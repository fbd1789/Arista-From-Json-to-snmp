package passpersist

import (
	"log/syslog"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	// default to no logging
	DisableLogging()
}

const (
	AristaExperimentalMib = "1.3.6.1.4.1.30065.4"
	NetSnmpExtendMib      = "1.3.6.1.4.1.8072.1.3.1"
)

var (
	BaseOid = MustNewOid(NetSnmpExtendMib).MustAppend([]int{224})
	// 224 = 112 + 112 = pp
	RefreshInterval = time.Second * 60
)

func SetLogLevel(s string) error {
	l, err := zerolog.ParseLevel(s)
	if err != nil {
		return err
	}
	zerolog.SetGlobalLevel(l)
	return nil
}

func LogLevel() string {
	return zerolog.GlobalLevel().String()
}

func DisableLogging() {
	SetLogLevel("disabled")
}

func EnableConsoleLogger(level string) {
	SetLogLevel(level)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
}

func EnableSyslogLogger(level string, prio syslog.Priority, tag string) {
	sw, err := syslog.New(prio, tag)
	if err != nil {
		panic(err)
	}

	SetLogLevel(level)
	log.Logger = zerolog.New(zerolog.SyslogLevelWriter(sw))
}
