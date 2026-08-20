package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RenderOptions bundles everything needed to render a single Markdown
// file to PDF.
type RenderOptions struct {
	InputPath    string
	OutputPath   string
	TemplatePath string
	PandocBin    string
	ChromiumBin  string
}

// Render runs the full pandoc -> Chromium pipeline: it reads
// InputPath, rewrites \newpage markers, extracts a document title,
// invokes pandoc against TemplatePath to produce a self-contained
// HTML file, invokes Chromium in headless mode to print that HTML to
// PDF, and moves the result to OutputPath. Intermediate files live in
// a temp directory that is removed before Render returns.
func Render(opts RenderOptions) error {
	raw, err := os.ReadFile(opts.InputPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", opts.InputPath, err)
	}

	title := ExtractTitle(string(raw))
	processed := PreprocessNewpage(string(raw))

	tmpDir, err := os.MkdirTemp("", "marc-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	htmlPath := filepath.Join(tmpDir, "output.html")

	pandocArgs := []string{
		"--from", "markdown",
		"--template", opts.TemplatePath,
		"--embed-resources",
		"--standalone",
		"-o", htmlPath,
	}
	if title != "" {
		pandocArgs = append(pandocArgs, "-M", "title="+title)
	}
	pandocCmd := exec.Command(opts.PandocBin, pandocArgs...)
	// Run from the input file's directory so relative asset paths
	// (e.g. local images) in the Markdown resolve the same way they
	// would if you ran pandoc by hand from that directory - matching
	// render.sh, which always ran pandoc from the project dir.
	pandocCmd.Dir = filepath.Dir(opts.InputPath)
	pandocCmd.Stdin = strings.NewReader(processed)
	if out, err := pandocCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pandoc failed: %w\n%s", err, out)
	}

	tmpPDFPath := filepath.Join(tmpDir, "output.pdf")
	chromiumArgs := []string{
		"--headless",
		"--disable-gpu",
		"--no-sandbox",
		"--run-all-compositor-stages-before-draw",
		"--no-pdf-header-footer",
		"--print-to-pdf-no-header",
		"--print-to-pdf=" + tmpPDFPath,
		"file://" + htmlPath,
	}
	out, err := exec.Command(opts.ChromiumBin, chromiumArgs...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("chromium failed: %w\n%s", err, out)
	}
	if _, err := os.Stat(tmpPDFPath); err != nil {
		return fmt.Errorf("chromium exited successfully but produced no PDF; its output:\n%s", out)
	}

	if err := moveFile(tmpPDFPath, opts.OutputPath); err != nil {
		return fmt.Errorf("moving rendered PDF to %s: %w", opts.OutputPath, err)
	}
	return nil
}

// moveFile moves src to dst, falling back to copy+remove if a direct
// rename fails - os.Rename can't cross filesystem boundaries, and the
// system temp dir (where src lives) is commonly a different
// filesystem (e.g. tmpfs) than the destination.
func moveFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return err
	}
	return os.Remove(src)
}
