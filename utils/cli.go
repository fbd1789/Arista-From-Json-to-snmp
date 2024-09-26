package utils

import (
	"flag"

	"log/slog"

	"github.com/arista-northwest/go-passpersist/utils/logger"
)

func CommonCLI(version string, tag string, buildDate string) {
	ver := flag.Bool("v", false, "display version")
	debug := flag.Bool("debug", false, "override extension logging and enable console debugging")
	console := flag.Bool("console", false, "enable console logging")
	level := flag.String("level", "INFO", "set logging level")
	flag.Parse()

	if *ver {
		DisplayVersionAndExit(version, buildDate, tag)
	}

	if *debug {
		logger.EnableConsoleLogger(slog.LevelDebug, true)
	}

	if *console {
		var l slog.Level
		l.UnmarshalText([]byte(*level))
		logger.EnableConsoleLogger(l, false)
	}
}
