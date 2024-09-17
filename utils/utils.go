package utils

import (
	"os"
	"path/filepath"
)

func ProgName() string {
	return filepath.Base(os.Args[0])
}
