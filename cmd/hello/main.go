package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/arista-northwest/go-passpersist/passpersist"
	"github.com/arista-northwest/go-passpersist/utils"
)

var (
	date    string
	tag     string
	version string
)

func init() {
	//logger.EnableSyslogger(syslog.LOG_LOCAL4, slog.LevelInfo)
}

// func redirectStderr(f *os.File) {
// 	err := syscall.Dup2(int(f.Fd()), int(os.Stderr.Fd()))
// 	if err != nil {
// 		log.Fatalf("Failed to redirect stderr to file: %v", err)
// 	}
// }

func main() {
	defer utils.CapPanic()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	utils.CommonCLI(version, tag, date)

	var opts []passpersist.Option

	b, _ := utils.GetBaseOIDFromSNMPdConfig()
	if b != nil {
		opts = append(opts, passpersist.WithBaseOID(*b))
	}
	opts = append(opts, passpersist.WithRefresh(time.Second*300))

	pp := passpersist.NewPassPersist(opts...)

	pp.Run(ctx, func(pp *passpersist.PassPersist) {
		slog.Debug("updating...")
		pp.AddString([]int{0}, "Hello from PassPersist")
		pp.AddString([]int{1}, "You found a secret message!")
	})
}
