package passpersist

import (
	"context"
	"log/slog"
	"time"

	"github.com/sethvargo/go-envconfig"
)

type ConfigFunc func(*Config)

type Config struct {
	BaseOID         OID           `json:"base-oid" env:"BASE_OID, overwrite"`
	RefreshInterval time.Duration `json:"refresh-interval" env:"REFRESH_INTERVAL, overwrite"`
	ConsoleDebug    bool          `json:"console-debug" env:"CONSOLE_DEBUG, overwrite"`
}

func NewConfigWithDefaults(ctx context.Context) Config {
	//ctx := context.Background()
	c := Config{
		BaseOID:         MustNewOID(NetSnmpExtendMib),
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

func WithBaseOID(oid OID) ConfigFunc {
	return func(c *Config) {
		c.BaseOID = oid
	}
}

func WithRefreshInterval(interval time.Duration) ConfigFunc {
	return func(c *Config) {
		c.RefreshInterval = interval
	}
}
