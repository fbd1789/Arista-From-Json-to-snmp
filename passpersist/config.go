package passpersist

import (
	"context"
	"log/slog"
	"time"

	"github.com/sethvargo/go-envconfig"
)

type ConfigFunc func(*Config)

type Config struct {
	BaseOid         Oid           `env:"BASE_OID, overwrite"`
	RefreshInterval time.Duration `env:"REFRESH_INTERVAL, overwrite"`
	ConsoleDebug    bool          `env:"CONSOLE_DEBUG, overwrite"`
}

func NewConfigWithDefaults(ctx context.Context) Config {
	//ctx := context.Background()
	c := Config{
		BaseOid:         MustNewOid(NetSnmpExtendMib),
		RefreshInterval: time.Second * 60,
		ConsoleDebug:    false,
	}

	if err := envconfig.Process(ctx, &c); err != nil {
		slog.Warn("failed to process env", slog.Any("error", err))
	}
	//fmt.Println("CONFIG_DEFAULTS:", c)
	return c
}

func WithConsoleDebug(cfg *Config) {
	cfg.ConsoleDebug = true
}

func WithBaseOid(oid Oid) ConfigFunc {
	return func(c *Config) {
		c.BaseOid = oid
	}
}

func WithRefreshInterval(interval time.Duration) ConfigFunc {
	return func(c *Config) {
		c.RefreshInterval = interval
	}
}
