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
	Config.BaseOid = MustNewOid(DEFAULT_BASE_OID)

}

var Config ConfigT

type ConfigT struct {
	BaseOid  *Oid
	Refresh  time.Duration
	LogLevel int
}
