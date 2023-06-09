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
	conf := &passpersist.ConfigT{
		BaseOid:  passpersist.MustNewOid(passpersist.DEFAULT_BASE_OID),
		Refresh:  1 * time.Second,
		LogLevel: 5,
	}
	pp := passpersist.NewPassPersist(conf)
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
