package utils

import (
	"path/filepath"
)

func ResolveFilename(prepath, filename string) string {
	return filepath.Join(prepath, filename)
}
