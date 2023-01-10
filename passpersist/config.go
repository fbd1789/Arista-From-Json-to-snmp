package passpersist

import (
	"time"
)

const (
	DEFAULT_REFRESH   = time.Second * 30
	DEFAULT_LOG_LEVEL = 4                         // warn
	DEFAULT_BASE_OID  = "1.3.6.1.4.1.30065.4.224" // enterprises::arista::aristaExperiment.224
	// 224 = 112 + 122 = pp
)

func init() {
	setDefaults()
}

func setDefaults() {
	Config.Refresh = DEFAULT_REFRESH
	Config.LogLevel = DEFAULT_LOG_LEVEL
	Config.BaseOid, _ = OIDFromString(DEFAULT_BASE_OID)

}

// func FlagInit() {
// 	var err error

// 	flag.DurationVar(&Config.Refresh, "refresh", DEFAULT_REFRESH, "refresh timer")
// 	flag.IntVar(&Config.LogLevel, "level", DEFAULT_LOG_LEVEL, "logging level (-1 - 5)")

// 	var baseOid OID
// 	o := flag.String("base-oid", DEFAULT_BASE_OID, "base OID")

// 	baseOid, err = OIDFromString(*o)
// 	if err != nil {
// 		log.Fatal().Msgf("invalid OID: '%s'", *o)
// 	}
// 	Config.BaseOid = baseOid

// 	flag.Parse()
// }

var Config ConfigT

type ConfigT struct {
	BaseOid  OID
	Refresh  time.Duration
	LogLevel int
}
