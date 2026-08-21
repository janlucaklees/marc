package main

import (
	"path/filepath"
	"strings"
)

// DefaultOutputPath returns the default PDF output path for a given
// Markdown input path: the same directory and basename, with the
// extension replaced by .pdf.
func DefaultOutputPath(inputPath string) string {
	ext := filepath.Ext(inputPath)
	base := strings.TrimSuffix(inputPath, ext)
	return base + ".pdf"
}
