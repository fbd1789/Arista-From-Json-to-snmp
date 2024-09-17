package utils

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"
)

func MustLoadMockDataFile(d any, path string) {
	file, err := os.Open(path)
	if err != nil {
		slog.Error("failed to open file", "path", path, slog.Any("error", err))
		os.Exit(1)
	}
	defer file.Close()

	raw, err := io.ReadAll(file)
	if err != nil {
		slog.Error("failed to read file", slog.Any("error", err))
		os.Exit(1)
	}
	if err := json.Unmarshal(raw, &d); err != nil {
		slog.Error("failed to parse json", slog.Any("error", err))
		os.Exit(1)
	}
}
