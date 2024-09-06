package passpersist

import (
	"time"
)

const (
	AristaExperimentalMib = "1.3.6.1.4.1.30065.4"
	NetSnmpExtendMib      = "1.3.6.1.4.1.8072.1.3.1"
)

type Config struct {
	BaseOid         Oid
	RefreshInterval time.Duration
}

func MustNewConfig(opts ...func(*Config)) *Config {
	cfg := &Config{
		BaseOid:         MustNewOid(NetSnmpExtendMib),
		RefreshInterval: time.Second * 60,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

func WithBaseOid(oid Oid) func(*Config) {
	return func(c *Config) {
		c.BaseOid = oid
	}
}

func WithRefreshInterval(interval time.Duration) func(*Config) {
	return func(c *Config) {
		c.RefreshInterval = interval
	}
}
