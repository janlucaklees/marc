package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// pandocInputFormat is the pandoc reader (plus extensions) used for
// the input Markdown. Pandoc's markdown only starts a list after a
// blank line, so a list directly below a paragraph would otherwise be
// folded into that paragraph as inline text.
const pandocInputFormat = "markdown+lists_without_preceding_blankline"

// RenderOptions bundles everything needed to render a single Markdown
// file to PDF.
type RenderOptions struct {
	InputPath    string
	OutputPath   string
	TemplatePath string
	PandocBin    string
	ChromiumBin  string
	Unpaged      bool
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
		"--from", pandocInputFormat,
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

	if opts.Unpaged {
		if err := makeUnpaged(opts.ChromiumBin, htmlPath, tmpDir); err != nil {
			return err
		}
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

// makeUnpaged rewrites the HTML at htmlPath so that printing it yields
// one page as tall as the content. It first loads a measuring copy in
// headless Chromium (--dump-dom, with a virtual-time budget so async
// content like Mermaid can finish) to learn the content height, then
// writes the final HTML with a matching @page size.
func makeUnpaged(chromiumBin, htmlPath, tmpDir string) error {
	html, err := os.ReadFile(htmlPath)
	if err != nil {
		return err
	}
	measureHTML, err := MeasurementHTML(string(html))
	if err != nil {
		return err
	}
	measurePath := filepath.Join(tmpDir, "measure.html")
	if err := os.WriteFile(measurePath, []byte(measureHTML), 0o644); err != nil {
		return err
	}

	cmd := exec.Command(chromiumBin,
		"--headless",
		"--disable-gpu",
		"--no-sandbox",
		fmt.Sprintf("--window-size=%d,1000", unpagedWidthPx),
		"--virtual-time-budget=10000",
		"--dump-dom",
		"file://"+measurePath,
	)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	dom, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("chromium (measuring pass) failed: %w\n%s", err, stderr.String())
	}
	height, err := ParseMeasuredHeight(string(dom))
	if err != nil {
		return err
	}

	final, err := UnpagedHTML(string(html), height)
	if err != nil {
		return err
	}
	return os.WriteFile(htmlPath, []byte(final), 0o644)
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
