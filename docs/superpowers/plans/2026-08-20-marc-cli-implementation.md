# marc CLI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn this repo into `marc`, an installable Go CLI that renders any Markdown file to PDF via a dotfiles-managed, named template, replacing the hardcoded `render.sh` pipeline.

**Architecture:** A single-package Go binary (`package main`, flat files at repo root) that: (1) resolves a template and the pandoc/Chromium binaries from XDG config, (2) preprocesses the Markdown (`\newpage` → pandoc pagebreak divs, title extraction), (3) shells out to `pandoc` then headless `chromium` to produce the PDF, matching `render.sh`'s existing flags exactly. Each unit (config, binary resolution, template resolution, path defaulting, preprocessing) is a small pure-logic file with its own tests; only the final shell-out pipeline and CLI wiring are integration-verified by hand, per the spec.

**Tech Stack:** Go (toolchain installed: go1.26.6), `github.com/BurntSushi/toml` (confirmed dependency for `config.toml` parsing — no other third-party dependencies), `pandoc` and a Chromium-based browser as external runtime dependencies (already installed: `pandoc` 3.10.2, `/usr/bin/chromium`).

**Spec:** `docs/superpowers/specs/2026-08-20-marc-cli-design.md`

## Global Constraints

- Single Markdown file input only (no batch rendering) — v1.
- Output defaults to `<input-basename>.pdf` next to the input file; `-o/--output` overrides it.
- Templates are dotfiles-managed and never shipped in this repo: resolved from `$XDG_CONFIG_HOME/marc/templates/<name>/template.html`, falling back to `~/.config/marc/templates/<name>/template.html` when `XDG_CONFIG_HOME` is unset.
- `config.toml` lives at `$XDG_CONFIG_HOME/marc/config.toml` (same fallback) and is entirely optional. v1 supports exactly two keys: `pandoc_bin`, `chromium_bin`.
- If `-t/--template` is omitted: prompt interactively when stdin is a TTY; if stdin is not a TTY, error — no silent default template.
- Dependency: `github.com/BurntSushi/toml` only. No other third-party Go modules.
- `go.mod` module path: `marc` (personal tool, not published).
- Chromium binary auto-detect order: `chromium`, `chromium-browser`, `google-chrome`, `google-chrome-stable`, `brave-browser`.
- Pandoc/Chromium invocation flags are ported verbatim from `render.sh`: pandoc gets `--from markdown --template <path> --embed-resources --standalone` (plus `-M title=<extracted title>` when a title is found); Chromium gets `--headless --disable-gpu --no-sandbox --run-all-compositor-stages-before-draw --no-pdf-header-footer --print-to-pdf-no-header --print-to-pdf=<path> file://<html-path>`.
- Cleanup sequencing per the spec: delete the confirmed Renaissance-doc list first (Task 1); delete `render.sh`, `template.tex`, `goal.pdf`, `.playwright-mcp/` only after `marc` is built and manually verified working (Task 8).

---

### Task 1: Remove Renaissance-specific content

**Files:**
- Delete: `Readme.md`, `Readme (1).md`, `docker-compose.yaml`, `.env`, `output.html`, `Systemdokumentation Renaissance (vorläufig).pdf`, `Installationsprotokoll Renaissance zweiter Teil vom 30.07.2026.pdf`, `.attachments.1259797/`, `attachments.1259797.zip`, `935-101.pdf`

**Interfaces:** None — this task only removes files, no code.

All of these files are currently untracked in git (confirmed via `git status` before this plan was written), so removing them produces no diff to commit — they simply vanish from the working tree.

- [ ] **Step 1: Delete the confirmed files and directories**

```bash
rm -f "Readme.md" "Readme (1).md" "docker-compose.yaml" ".env" "output.html" \
      "Systemdokumentation Renaissance (vorläufig).pdf" \
      "Installationsprotokoll Renaissance zweiter Teil vom 30.07.2026.pdf" \
      "attachments.1259797.zip" "935-101.pdf"
rm -rf ".attachments.1259797"
```

- [ ] **Step 2: Verify they're gone and git has nothing to commit for this**

```bash
ls Readme.md docker-compose.yaml .env 2>&1 | grep -q "No such file" && echo "confirmed removed"
git status --short
```

Expected: the `ls` check prints "confirmed removed"; `git status --short` shows no new changes related to these files (they were untracked, so nothing appears).

---

### Task 2: Project scaffolding, `\newpage` preprocessing, and title extraction

**Files:**
- Create: `go.mod`
- Create: `Makefile`
- Create: `preprocess.go`
- Test: `preprocess_test.go`

**Interfaces:**
- Produces: `PreprocessNewpage(src string) string`, `ExtractTitle(src string) string` — both used by `render.go` in Task 7.

- [ ] **Step 1: Initialize the Go module**

```bash
go mod init marc
```

Expected: creates `go.mod` with `module marc` and a `go` directive.

- [ ] **Step 2: Set the Go version floor in go.mod**

Open `go.mod` and confirm/set the version line to:

```
go 1.23
```

- [ ] **Step 3: Create the Makefile**

```makefile
.PHONY: build install test

build:
	go build -o bin/marc .

install: build
	mkdir -p $(HOME)/.local/bin
	cp bin/marc $(HOME)/.local/bin/marc

test:
	go test ./...
```

Note: `make build`/`make install` won't succeed yet — there's no `func main()` until Task 7 (`go build` requires one for package main; `go test` does not). Use `make test` / `go test ./...` to verify tasks 2–6.

- [ ] **Step 4: Write the failing tests for `PreprocessNewpage`**

Create `preprocess_test.go`:

```go
package main

import "testing"

func TestPreprocessNewpage_replacesMarker(t *testing.T) {
	input := "Section one.\n\n\\newpage\n\n## Section two\n"
	want := "Section one.\n\n::: pagebreak\n:::\n\n## Section two\n"
	got := PreprocessNewpage(input)
	if got != want {
		t.Errorf("PreprocessNewpage() = %q, want %q", got, want)
	}
}

func TestPreprocessNewpage_multipleMarkers(t *testing.T) {
	input := "a\\newpageb\\newpagec"
	want := "a::: pagebreak\n:::b::: pagebreak\n:::c"
	got := PreprocessNewpage(input)
	if got != want {
		t.Errorf("PreprocessNewpage() = %q, want %q", got, want)
	}
}

func TestPreprocessNewpage_noMarker(t *testing.T) {
	input := "nothing to replace here"
	got := PreprocessNewpage(input)
	if got != input {
		t.Errorf("PreprocessNewpage() = %q, want unchanged %q", got, input)
	}
}
```

- [ ] **Step 5: Run the tests to verify they fail**

```bash
go test ./... -run TestPreprocessNewpage -v
```

Expected: FAIL — `PreprocessNewpage` is undefined.

- [ ] **Step 6: Implement `PreprocessNewpage`**

Create `preprocess.go`:

```go
package main

import "strings"

// PreprocessNewpage rewrites literal `\newpage` markers in Markdown
// source into pandoc's `::: pagebreak :::` fenced-div syntax, matching
// the sed step from the original render.sh script. `\newpage` is not
// native Markdown/pandoc syntax; the pagebreak div, handled via
// template.html's raw-attribute CSS, is.
func PreprocessNewpage(src string) string {
	return strings.ReplaceAll(src, `\newpage`, "::: pagebreak\n:::")
}
```

- [ ] **Step 7: Run the tests to verify they pass**

```bash
go test ./... -run TestPreprocessNewpage -v
```

Expected: PASS (all 3 subtests).

- [ ] **Step 8: Write the failing tests for `ExtractTitle`**

Append to `preprocess_test.go`:

```go
func TestExtractTitle_findsFirstHeading(t *testing.T) {
	src := "intro text\n# Systemdokumentation Renaissance\n\nmore text\n# Second Heading\n"
	got := ExtractTitle(src)
	want := "Systemdokumentation Renaissance"
	if got != want {
		t.Errorf("ExtractTitle() = %q, want %q", got, want)
	}
}

func TestExtractTitle_noHeadingReturnsEmpty(t *testing.T) {
	src := "no headings here\njust text\n"
	got := ExtractTitle(src)
	if got != "" {
		t.Errorf("ExtractTitle() = %q, want empty string", got)
	}
}

func TestExtractTitle_ignoresSubheadings(t *testing.T) {
	src := "## Not a title\n# Actual Title\n"
	got := ExtractTitle(src)
	want := "Actual Title"
	if got != want {
		t.Errorf("ExtractTitle() = %q, want %q", got, want)
	}
}
```

- [ ] **Step 9: Run the tests to verify they fail**

```bash
go test ./... -run TestExtractTitle -v
```

Expected: FAIL — `ExtractTitle` is undefined.

- [ ] **Step 10: Implement `ExtractTitle`**

Append to `preprocess.go`:

```go
// ExtractTitle returns the text of the first top-level Markdown
// heading (a line beginning with "# ") in src, or "" if none is
// found. Mirrors render.sh's `grep '^# ' | head -1` step, used to set
// the rendered document's <title>.
func ExtractTitle(src string) string {
	for _, line := range strings.Split(src, "\n") {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}
```

- [ ] **Step 11: Run all tests to verify they pass**

```bash
go test ./... -v
```

Expected: PASS (all 6 subtests across both functions).

- [ ] **Step 12: Commit**

```bash
git add go.mod Makefile preprocess.go preprocess_test.go
git commit -m "feat: scaffold marc Go module, port newpage/title preprocessing"
```

---

### Task 3: Config loading

**Files:**
- Create: `config.go`
- Test: `config_test.go`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: `type Config struct { PandocBin, ChromiumBin string }`, `ConfigDir() (string, error)`, `LoadConfig(dir string) (Config, error)` — used by `main.go` in Task 7.

- [ ] **Step 1: Add the TOML dependency**

```bash
go get github.com/BurntSushi/toml
```

Expected: adds an entry to `go.mod` and creates/updates `go.sum`.

- [ ] **Step 2: Write the failing tests**

Create `config_test.go`:

```go
package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDir_usesXDGConfigHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdgtest")
	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() error = %v", err)
	}
	want := filepath.Join("/tmp/xdgtest", "marc")
	if dir != want {
		t.Errorf("ConfigDir() = %q, want %q", dir, want)
	}
}

func TestConfigDir_fallsBackToHomeConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "/tmp/hometest")
	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() error = %v", err)
	}
	want := filepath.Join("/tmp/hometest", ".config", "marc")
	if dir != want {
		t.Errorf("ConfigDir() = %q, want %q", dir, want)
	}
}

func TestLoadConfig_missingFileReturnsZeroValue(t *testing.T) {
	dir := t.TempDir()
	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg != (Config{}) {
		t.Errorf("LoadConfig() = %+v, want zero value", cfg)
	}
}

func TestLoadConfig_parsesBothFields(t *testing.T) {
	dir := t.TempDir()
	content := "pandoc_bin = \"/usr/bin/pandoc\"\nchromium_bin = \"/usr/bin/chromium\"\n"
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed writing test config: %v", err)
	}
	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	want := Config{PandocBin: "/usr/bin/pandoc", ChromiumBin: "/usr/bin/chromium"}
	if cfg != want {
		t.Errorf("LoadConfig() = %+v, want %+v", cfg, want)
	}
}

func TestLoadConfig_partialFieldsLeaveOthersZero(t *testing.T) {
	dir := t.TempDir()
	content := "pandoc_bin = \"/usr/bin/pandoc\"\n"
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed writing test config: %v", err)
	}
	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	want := Config{PandocBin: "/usr/bin/pandoc"}
	if cfg != want {
		t.Errorf("LoadConfig() = %+v, want %+v", cfg, want)
	}
}

func TestLoadConfig_malformedTOMLReturnsError(t *testing.T) {
	dir := t.TempDir()
	content := "this is not valid toml = = =\n"
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed writing test config: %v", err)
	}
	if _, err := LoadConfig(dir); err == nil {
		t.Error("LoadConfig() error = nil, want error for malformed TOML")
	}
}
```

- [ ] **Step 3: Run the tests to verify they fail**

```bash
go test ./... -run 'TestConfigDir|TestLoadConfig' -v
```

Expected: FAIL — `Config`, `ConfigDir`, `LoadConfig` are undefined.

- [ ] **Step 4: Implement config loading**

Create `config.go`:

```go
package main

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds optional overrides read from config.toml.
type Config struct {
	PandocBin   string `toml:"pandoc_bin"`
	ChromiumBin string `toml:"chromium_bin"`
}

// ConfigDir returns marc's config directory: $XDG_CONFIG_HOME/marc if
// XDG_CONFIG_HOME is set, otherwise $HOME/.config/marc.
func ConfigDir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "marc"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "marc"), nil
}

// LoadConfig reads <dir>/config.toml. If the file does not exist, it
// returns a zero-value Config and no error. Any other read or parse
// error is returned as-is.
func LoadConfig(dir string) (Config, error) {
	var cfg Config
	path := filepath.Join(dir, "config.toml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
```

- [ ] **Step 5: Run the tests to verify they pass**

```bash
go test ./... -v
```

Expected: PASS for all tests so far (Task 2's and Task 3's).

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum config.go config_test.go
git commit -m "feat: load marc config.toml (pandoc_bin/chromium_bin overrides)"
```

---

### Task 4: Binary resolution

**Files:**
- Create: `binary.go`
- Test: `binary_test.go`

**Interfaces:**
- Consumes: nothing from earlier tasks (takes a `lookup func(string) (string, error)` so it doesn't depend on `Config` directly).
- Produces: `ResolveBinary(toolLabel, configured string, candidates []string, lookup func(string) (string, error)) (string, error)` — used by `main.go` in Task 7 with `exec.LookPath` as `lookup`.

- [ ] **Step 1: Write the failing tests**

Create `binary_test.go`:

```go
package main

import (
	"errors"
	"testing"
)

func TestResolveBinary_usesConfiguredWhenSet(t *testing.T) {
	lookup := func(name string) (string, error) {
		if name == "/custom/pandoc" {
			return "/custom/pandoc", nil
		}
		return "", errors.New("not found")
	}
	path, err := ResolveBinary("pandoc", "/custom/pandoc", []string{"pandoc"}, lookup)
	if err != nil {
		t.Fatalf("ResolveBinary() error = %v", err)
	}
	if path != "/custom/pandoc" {
		t.Errorf("ResolveBinary() = %q, want %q", path, "/custom/pandoc")
	}
}

func TestResolveBinary_configuredButUnusableReturnsError(t *testing.T) {
	lookup := func(name string) (string, error) {
		return "", errors.New("no such file")
	}
	_, err := ResolveBinary("pandoc", "/bad/path", []string{"pandoc"}, lookup)
	if err == nil {
		t.Error("ResolveBinary() error = nil, want error for unusable configured path")
	}
}

func TestResolveBinary_triesCandidatesInOrder(t *testing.T) {
	lookup := func(name string) (string, error) {
		if name == "chromium-browser" {
			return "/usr/bin/chromium-browser", nil
		}
		return "", errors.New("not found")
	}
	path, err := ResolveBinary("chromium", "", []string{"chromium", "chromium-browser", "google-chrome"}, lookup)
	if err != nil {
		t.Fatalf("ResolveBinary() error = %v", err)
	}
	if path != "/usr/bin/chromium-browser" {
		t.Errorf("ResolveBinary() = %q, want %q", path, "/usr/bin/chromium-browser")
	}
}

func TestResolveBinary_noneFoundReturnsError(t *testing.T) {
	lookup := func(name string) (string, error) {
		return "", errors.New("not found")
	}
	_, err := ResolveBinary("chromium", "", []string{"chromium", "chromium-browser"}, lookup)
	if err == nil {
		t.Error("ResolveBinary() error = nil, want error when no candidate resolves")
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

```bash
go test ./... -run TestResolveBinary -v
```

Expected: FAIL — `ResolveBinary` is undefined.

- [ ] **Step 3: Implement `ResolveBinary`**

Create `binary.go`:

```go
package main

import "fmt"

// ResolveBinary determines which executable path to use for a tool.
// If configured is non-empty, it is checked via lookup and returned
// (or an error if lookup fails). Otherwise each name in candidates is
// tried via lookup, in order; the first that resolves is returned. If
// nothing resolves, an error names toolLabel and every candidate
// tried.
func ResolveBinary(toolLabel, configured string, candidates []string, lookup func(string) (string, error)) (string, error) {
	if configured != "" {
		path, err := lookup(configured)
		if err != nil {
			return "", fmt.Errorf("configured %s binary %q not usable: %w", toolLabel, configured, err)
		}
		return path, nil
	}
	var tried []string
	for _, name := range candidates {
		if path, err := lookup(name); err == nil {
			return path, nil
		}
		tried = append(tried, name)
	}
	return "", fmt.Errorf("no %s binary found (tried: %v); set %s_bin in config.toml or install one on PATH", toolLabel, tried, toolLabel)
}
```

- [ ] **Step 4: Run the tests to verify they pass**

```bash
go test ./... -v
```

Expected: PASS for all tests so far.

- [ ] **Step 5: Commit**

```bash
git add binary.go binary_test.go
git commit -m "feat: resolve pandoc/chromium binaries from config or PATH"
```

---

### Task 5: Template resolution

**Files:**
- Create: `template.go`
- Test: `template_test.go`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: `ErrNoTemplate` (error value), `TemplatesDir(configDir string) string`, `ListTemplates(templatesDir string) ([]string, error)`, `TemplatePath(templatesDir, name string) (string, error)`, `IsInteractive(f *os.File) bool`, `PromptTemplate(names []string, in io.Reader, out io.Writer) (string, error)` — all used by `main.go` in Task 7.

Note: `IsInteractive` is a one-line `os.File.Stat()` wrapper around real terminal detection — like the pandoc/Chromium shell-outs, it's not meaningfully unit-testable in isolation and isn't tested here; it's exercised manually in Task 7.

- [ ] **Step 1: Write the failing tests**

Create `template_test.go`:

```go
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mustWriteTemplate(t *testing.T, templatesDir, name string) {
	t.Helper()
	dir := filepath.Join(templatesDir, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("failed creating template dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "template.html"), []byte("<html></html>"), 0o644); err != nil {
		t.Fatalf("failed writing template.html: %v", err)
	}
}

func TestListTemplates_returnsSortedNames(t *testing.T) {
	dir := t.TempDir()
	mustWriteTemplate(t, dir, "work")
	mustWriteTemplate(t, dir, "private")
	names, err := ListTemplates(dir)
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}
	want := []string{"private", "work"}
	if len(names) != len(want) || names[0] != want[0] || names[1] != want[1] {
		t.Errorf("ListTemplates() = %v, want %v", names, want)
	}
}

func TestListTemplates_ignoresDirsWithoutTemplateHTML(t *testing.T) {
	dir := t.TempDir()
	mustWriteTemplate(t, dir, "work")
	if err := os.MkdirAll(filepath.Join(dir, "empty"), 0o755); err != nil {
		t.Fatalf("failed creating empty dir: %v", err)
	}
	names, err := ListTemplates(dir)
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}
	if len(names) != 1 || names[0] != "work" {
		t.Errorf("ListTemplates() = %v, want [work]", names)
	}
}

func TestListTemplates_missingDirReturnsEmpty(t *testing.T) {
	names, err := ListTemplates(filepath.Join(t.TempDir(), "does-not-exist"))
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}
	if len(names) != 0 {
		t.Errorf("ListTemplates() = %v, want empty", names)
	}
}

func TestTemplatePath_found(t *testing.T) {
	dir := t.TempDir()
	mustWriteTemplate(t, dir, "work")
	path, err := TemplatePath(dir, "work")
	if err != nil {
		t.Fatalf("TemplatePath() error = %v", err)
	}
	want := filepath.Join(dir, "work", "template.html")
	if path != want {
		t.Errorf("TemplatePath() = %q, want %q", path, want)
	}
}

func TestTemplatePath_notFoundReturnsError(t *testing.T) {
	dir := t.TempDir()
	if _, err := TemplatePath(dir, "missing"); err == nil {
		t.Error("TemplatePath() error = nil, want error for missing template")
	}
}

func TestPromptTemplate_selectsByExactName(t *testing.T) {
	names := []string{"private", "work"}
	in := strings.NewReader("work\n")
	var out strings.Builder
	got, err := PromptTemplate(names, in, &out)
	if err != nil {
		t.Fatalf("PromptTemplate() error = %v", err)
	}
	if got != "work" {
		t.Errorf("PromptTemplate() = %q, want %q", got, "work")
	}
}

func TestPromptTemplate_selectsByNumber(t *testing.T) {
	names := []string{"private", "work"}
	in := strings.NewReader("2\n")
	var out strings.Builder
	got, err := PromptTemplate(names, in, &out)
	if err != nil {
		t.Fatalf("PromptTemplate() error = %v", err)
	}
	if got != "work" {
		t.Errorf("PromptTemplate() = %q, want %q", got, "work")
	}
}

func TestPromptTemplate_invalidSelectionReturnsError(t *testing.T) {
	names := []string{"private", "work"}
	in := strings.NewReader("bogus\n")
	var out strings.Builder
	if _, err := PromptTemplate(names, in, &out); err == nil {
		t.Error("PromptTemplate() error = nil, want error for invalid selection")
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

```bash
go test ./... -run 'TestListTemplates|TestTemplatePath|TestPromptTemplate' -v
```

Expected: FAIL — `ListTemplates`, `TemplatePath`, `PromptTemplate` are undefined.

- [ ] **Step 3: Implement template resolution**

Create `template.go`:

```go
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// ErrNoTemplate indicates no template name was given and the session
// is non-interactive, so no prompt could be shown.
var ErrNoTemplate = fmt.Errorf("no template specified; pass -t/--template (session is non-interactive)")

// TemplatesDir returns the templates directory beneath configDir.
func TemplatesDir(configDir string) string {
	return filepath.Join(configDir, "templates")
}

// ListTemplates returns the names of available templates in
// templatesDir: subdirectories that contain a template.html file,
// sorted alphabetically. If templatesDir does not exist, it returns
// an empty slice and no error.
func ListTemplates(templatesDir string) ([]string, error) {
	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(templatesDir, e.Name(), "template.html")); err == nil {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

// TemplatePath returns the template.html path for name within
// templatesDir. Returns an error if that file doesn't exist.
func TemplatePath(templatesDir, name string) (string, error) {
	path := filepath.Join(templatesDir, name, "template.html")
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("template %q not found (looked for %s)", name, path)
	}
	return path, nil
}

// IsInteractive reports whether f looks like a terminal, as opposed
// to a pipe or redirected file.
func IsInteractive(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// PromptTemplate writes a numbered list of names to out, reads a
// single line from in, and returns the selected name. The line may be
// either the exact template name or its 1-based list number.
func PromptTemplate(names []string, in io.Reader, out io.Writer) (string, error) {
	fmt.Fprintln(out, "Available templates:")
	for i, n := range names {
		fmt.Fprintf(out, "  %d) %s\n", i+1, n)
	}
	fmt.Fprint(out, "Select a template: ")

	scanner := bufio.NewScanner(in)
	if !scanner.Scan() {
		return "", fmt.Errorf("no template selection received")
	}
	choice := strings.TrimSpace(scanner.Text())

	for _, n := range names {
		if n == choice {
			return n, nil
		}
	}
	if idx, err := strconv.Atoi(choice); err == nil && idx >= 1 && idx <= len(names) {
		return names[idx-1], nil
	}
	return "", fmt.Errorf("invalid template selection %q", choice)
}
```

- [ ] **Step 4: Run the tests to verify they pass**

```bash
go test ./... -v
```

Expected: PASS for all tests so far.

- [ ] **Step 5: Commit**

```bash
git add template.go template_test.go
git commit -m "feat: resolve and interactively prompt for marc templates"
```

---

### Task 6: Output path helper

**Files:**
- Create: `paths.go`
- Test: `paths_test.go`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: `DefaultOutputPath(inputPath string) string` — used by `main.go` in Task 7.

- [ ] **Step 1: Write the failing tests**

Create `paths_test.go`:

```go
package main

import "testing"

func TestDefaultOutputPath_replacesExtension(t *testing.T) {
	got := DefaultOutputPath("/home/jlk/docs/Readme.md")
	want := "/home/jlk/docs/Readme.pdf"
	if got != want {
		t.Errorf("DefaultOutputPath() = %q, want %q", got, want)
	}
}

func TestDefaultOutputPath_relativePath(t *testing.T) {
	got := DefaultOutputPath("notes.markdown")
	want := "notes.pdf"
	if got != want {
		t.Errorf("DefaultOutputPath() = %q, want %q", got, want)
	}
}

func TestDefaultOutputPath_noExtension(t *testing.T) {
	got := DefaultOutputPath("README")
	want := "README.pdf"
	if got != want {
		t.Errorf("DefaultOutputPath() = %q, want %q", got, want)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

```bash
go test ./... -run TestDefaultOutputPath -v
```

Expected: FAIL — `DefaultOutputPath` is undefined.

- [ ] **Step 3: Implement `DefaultOutputPath`**

Create `paths.go`:

```go
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
```

- [ ] **Step 4: Run the tests to verify they pass**

```bash
go test ./... -v
```

Expected: PASS for all tests so far.

- [ ] **Step 5: Commit**

```bash
git add paths.go paths_test.go
git commit -m "feat: default PDF output path from input basename"
```

---

### Task 7: Render pipeline, CLI wiring, and manual verification

**Files:**
- Create: `render.go`
- Create: `main.go`

**Interfaces:**
- Consumes: `PreprocessNewpage`, `ExtractTitle` (Task 2), `Config`, `ConfigDir`, `LoadConfig` (Task 3), `ResolveBinary` (Task 4), `ErrNoTemplate`, `TemplatesDir`, `ListTemplates`, `TemplatePath`, `IsInteractive`, `PromptTemplate` (Task 5), `DefaultOutputPath` (Task 6).
- Produces: the `marc` binary itself (`func main()`). Nothing else consumes this — it's the top of the call graph.

This task is not unit-tested — per the spec, the pandoc/Chromium shell-out pipeline is a thin wrapper around external processes, verified by actually running the tool (Step 5 below), not by `go test`.

- [ ] **Step 1: Implement the render pipeline**

Create `render.go`:

```go
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
	if out, err := exec.Command(opts.ChromiumBin, chromiumArgs...).CombinedOutput(); err != nil {
		return fmt.Errorf("chromium failed: %w\n%s", err, out)
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
```

- [ ] **Step 2: Implement the CLI entry point**

Create `main.go`:

```go
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
)

var chromiumCandidates = []string{
	"chromium",
	"chromium-browser",
	"google-chrome",
	"google-chrome-stable",
	"brave-browser",
}

func main() {
	os.Exit(run())
}

func run() int {
	var templateName string
	var outputPath string

	fs := flag.NewFlagSet("marc", flag.ContinueOnError)
	fs.StringVar(&templateName, "template", "", "template name (dotfiles-managed)")
	fs.StringVar(&templateName, "t", "", "shorthand for -template")
	fs.StringVar(&outputPath, "output", "", "output PDF path (default: input basename with .pdf)")
	fs.StringVar(&outputPath, "o", "", "shorthand for -output")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: marc [-t|--template <name>] [-o|--output <path>] <input.md>")
		fs.PrintDefaults()
	}
	if err := fs.Parse(os.Args[1:]); err != nil {
		return 2
	}

	if fs.NArg() != 1 {
		fs.Usage()
		return 2
	}
	inputPath := fs.Arg(0)

	if _, err := os.Stat(inputPath); err != nil {
		fmt.Fprintf(os.Stderr, "marc: %v\n", err)
		return 1
	}

	configDir, err := ConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "marc: resolving config directory: %v\n", err)
		return 1
	}

	cfg, err := LoadConfig(configDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "marc: loading config: %v\n", err)
		return 1
	}

	pandocBin, err := ResolveBinary("pandoc", cfg.PandocBin, []string{"pandoc"}, exec.LookPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "marc: %v\n", err)
		return 1
	}
	chromiumBin, err := ResolveBinary("chromium", cfg.ChromiumBin, chromiumCandidates, exec.LookPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "marc: %v\n", err)
		return 1
	}

	templatesDir := TemplatesDir(configDir)
	if templateName == "" {
		names, err := ListTemplates(templatesDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "marc: listing templates: %v\n", err)
			return 1
		}
		if len(names) == 0 {
			fmt.Fprintf(os.Stderr, "marc: no templates found in %s\n", templatesDir)
			return 1
		}
		if !IsInteractive(os.Stdin) {
			fmt.Fprintf(os.Stderr, "marc: %v\n", ErrNoTemplate)
			return 1
		}
		templateName, err = PromptTemplate(names, os.Stdin, os.Stdout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "marc: %v\n", err)
			return 1
		}
	}

	templatePath, err := TemplatePath(templatesDir, templateName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "marc: %v\n", err)
		return 1
	}

	if outputPath == "" {
		outputPath = DefaultOutputPath(inputPath)
	}

	if err := Render(RenderOptions{
		InputPath:    inputPath,
		OutputPath:   outputPath,
		TemplatePath: templatePath,
		PandocBin:    pandocBin,
		ChromiumBin:  chromiumBin,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "marc: %v\n", err)
		return 1
	}

	fmt.Printf("Rendered %s\n", outputPath)
	return 0
}
```

- [ ] **Step 3: Run all unit tests to confirm nothing broke**

```bash
go test ./... -v
```

Expected: PASS for every test from Tasks 2–6 (this task adds no new unit tests).

- [ ] **Step 4: Build the binary**

```bash
make build
```

Expected: produces `bin/marc` with no errors.

- [ ] **Step 5: Manually verify the full pipeline in an isolated fake HOME**

`template.html` (the existing pandoc HTML template in this repo) is still present at this point in the plan — reuse it as the sample template for this manual check.

```bash
MARC_TEST_HOME=$(mktemp -d)
mkdir -p "$MARC_TEST_HOME/.config/marc/templates/test"
cp template.html "$MARC_TEST_HOME/.config/marc/templates/test/template.html"
cat > "$MARC_TEST_HOME/sample.md" <<'EOF'
# Sample Document

Some text on the first page.

\newpage

## Second page

More text here.
EOF

HOME="$MARC_TEST_HOME" XDG_CONFIG_HOME="$MARC_TEST_HOME/.config" \
  ./bin/marc -t test "$MARC_TEST_HOME/sample.md"

file "$MARC_TEST_HOME/sample.pdf"
```

Expected: prints `Rendered <path>/sample.pdf`; `file` reports a valid PDF document (e.g. `PDF document, version ...`).

- [ ] **Step 6: Manually verify the error paths**

```bash
# Unknown template name -> clear "not found" error, exit 1
HOME="$MARC_TEST_HOME" XDG_CONFIG_HOME="$MARC_TEST_HOME/.config" \
  ./bin/marc -t doesnotexist "$MARC_TEST_HOME/sample.md"; echo "exit: $?"

# No template given, non-interactive stdin -> ErrNoTemplate, exit 1
HOME="$MARC_TEST_HOME" XDG_CONFIG_HOME="$MARC_TEST_HOME/.config" \
  ./bin/marc "$MARC_TEST_HOME/sample.md" < /dev/null; echo "exit: $?"

# Missing input file -> stat error, exit 1
HOME="$MARC_TEST_HOME" XDG_CONFIG_HOME="$MARC_TEST_HOME/.config" \
  ./bin/marc -t test "$MARC_TEST_HOME/does-not-exist.md"; echo "exit: $?"
```

Expected: all three print a `marc: ...` error naming the actual problem and exit with code 1.

- [ ] **Step 7: Clean up the manual test sandbox**

```bash
rm -rf "$MARC_TEST_HOME"
```

- [ ] **Step 8: Commit**

```bash
git add render.go main.go
git commit -m "feat: wire up marc CLI (render pipeline + flag parsing)"
```

---

### Task 8: Final cleanup and CLAUDE.md rewrite

**Files:**
- Delete: `render.sh`, `template.tex`, `goal.pdf`, `.playwright-mcp/`
- Modify: `CLAUDE.md`

**Interfaces:** None.

`render.sh`, `template.tex`, `goal.pdf`, and `.playwright-mcp/` are all currently untracked in git, same as Task 1's deletions — removing them produces no diff to commit on its own; only the `CLAUDE.md` rewrite is committed.

- [ ] **Step 1: Confirm Task 7's manual verification passed**

Don't proceed with this task until Task 7's Steps 5–6 both succeeded — this cleanup assumes `marc` fully replaces `render.sh`.

- [ ] **Step 2: Delete the superseded files**

```bash
rm -f render.sh template.tex goal.pdf
rm -rf .playwright-mcp
```

- [ ] **Step 3: Rewrite CLAUDE.md for the marc CLI project**

Replace `CLAUDE.md`'s content (keep the required prefix) to describe:
- What `marc` is: a Go CLI that renders any Markdown file to PDF via pandoc + headless Chromium, using a dotfiles-managed, named template (`~/.config/marc/templates/<name>/template.html`, or `$XDG_CONFIG_HOME/marc/templates/<name>/` if set).
- Commands: `make build`, `make install`, `make test` (or `go test ./...`).
- Architecture/file map: `main.go` (CLI wiring), `render.go` (pandoc/Chromium pipeline), `config.go` (XDG config + `config.toml`), `binary.go` (pandoc/Chromium binary resolution), `template.go` (template resolution + interactive prompt), `paths.go` (default output path), `preprocess.go` (`\newpage` rewrite + title extraction) — each with its own `_test.go`.
- The one external dependency: `github.com/BurntSushi/toml`.
- That templates are never stored in this repo — they live in the user's dotfiles.

- [ ] **Step 4: Verify the build still works after cleanup**

```bash
make build && ./bin/marc -h
```

Expected: builds cleanly; `-h` prints the usage line (exit code 2 is expected from `flag`'s default `-h` handling — that's fine, this is just confirming the binary still runs).

- [ ] **Step 5: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: rewrite CLAUDE.md for the marc CLI, drop superseded render.sh"
```
