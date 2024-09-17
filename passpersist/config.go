package passpersist

import (
	"time"
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

	for _, f := range opts {
		f(cfg)
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
