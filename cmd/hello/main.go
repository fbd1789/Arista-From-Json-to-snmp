package main

import (
	"context"
	"fmt"
	"time"

	"github.com/arista-northwest/go-passpersist/passpersist"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//console logger breaks the passpersis protocol even though it writes to stderr
	//passpersist.EnableConsoleLogger("debug")
	//passpersist.EnableSyslogLogger("debug", syslog.LOG_LOCAL4, "passpersist-hello")

	oid := passpersist.MustNewOid(passpersist.NetSnmpExtendMib)

	cfg := passpersist.MustNewConfig(
		passpersist.WithBaseOid(oid),
		passpersist.WithRefreshInterval(time.Second),
	)

	pp := passpersist.NewPassPersist(cfg)

	pp.Run(ctx, func(pp *passpersist.PassPersist) {
		pp.AddString([]int{0}, "Hello from PassPersist")
		pp.AddString([]int{1}, "You found a secret message!")

		for i := 2; i <= 10; i++ {
			for j := 1; j <= 10; j++ {
				pp.AddString([]int{i, j}, fmt.Sprintf("Value: %d.%d", i, j))
			}
		}
	})
}
