# go-passpersist

Golang implementation of SNMP's Pass-Persist protocol


### Example

```
package main

import (
	"context"
	"fmt"
	"log/syslog"
	"time"

	"github.com/arista-northwest/go-passpersist/passpersist"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// console logger breaks the passpersis protocol even though it writes to stderr
	//passpersist.EnableConsoleLogger("debug")
	// or send log to syslog
	//passpersist.EnableSyslogLogger("debug", syslog.LOG_LOCAL4, "passpersist-hello")
	passpersist.RefreshInterval = 60 * time.Second

	pp := passpersist.NewPassPersist()

	pp.Run(ctx, func(pp *passpersist.PassPersist) {
		pp.AddString([]int{0}, "Hello from PassPersist")
		pp.AddString([]int{1}, "You found a secret message!")
	})
}

```
# Arista-From-Json-to-snmp
