package main

import (
	"context"
	"time"

	"github.com/arista-northwest/go-snmppasspersist/passpersist"
)

func main() {
	passpersist.Config.Refresh = time.Second * 5
	pp := passpersist.NewPassPersist(&passpersist.Config)
	ctx := context.Background()
	pp.Run(ctx, func(pp *passpersist.PassPersist) {
		pp.AddString([]int{255, 0}, "Hello")
		pp.AddInt([]int{255, 1}, 42)
		pp.AddString([]int{255, 2}, "!")
	})
}
