# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repository is

`marc` is a small Go CLI that renders a Markdown file to PDF via `pandoc` (Markdown → self-contained HTML) followed by headless Chromium (HTML → PDF), using a named, dotfiles-managed HTML template. There is no application server, no web frontend, no Docker deployment here — just the CLI source, its tests, and the Makefile that builds/installs it.

Usage: `marc [-t|--template <name>] [-i|--interactive] [-u|--unpaged] [-o|--output <path>] <input.md>`

- `-u/--unpaged` renders a single page as tall as the content instead of paginated A4 (see `unpaged.go`). Chromium has no page-size flag, so `render.go` first loads a measuring copy via `--dump-dom` to get the content height, then injects a matching `@page` size and prints. It also injects CSS that overrides the template's `@media print` layout so the PDF looks like the on-screen view.

- `-t/--template <name>` selects a template by name directly, erroring if it doesn't exist. Takes priority over everything below if given.
- `-i/--interactive` forces the template picker for this run, ignoring `config.toml`'s `default_template`.
- With neither flag: `marc` uses `config.toml`'s `default_template` (implicitly `"default"` if the key or file is absent — see "Templates and config live outside this repo"). If that resolves to an empty string, it behaves like `-i`: prompts if the session is interactive, errors if not (e.g. piped/scripted).
- `-o/--output` sets the output PDF path. If omitted, it defaults to the input path with its extension replaced by `.pdf` (see `paths.go`).

## Commands

- `make build` — builds the binary to `bin/marc`.
- `make install` — builds, then runs `go install ./cmd/marc` (drops the binary in `$(go env GOBIN)`, falling back to `~/go/bin`).
- `make test` (or `go test ./...`) — runs all package tests.

These Makefile targets are the only defined workflow commands — use them rather than invoking `go build`/`go install` directly.

## Architecture

All source lives in package `main` under `cmd/marc/`, one `.go` file per concern — kept separate from the repo-root project files (`go.mod`, `Makefile`, `CLAUDE.md`, `docs/`). Six of the eight have a sibling `_test.go` (`binary.go`, `config.go`, `paths.go`, `preprocess.go`, `template.go`, `unpaged.go`) — pure logic covered by unit tests. `main.go` and `render.go` deliberately have no `_test.go`: they're thin wrappers around external processes (pandoc, Chromium) and CLI wiring, and are instead verified by manual end-to-end testing (build the binary, render a real Markdown file, confirm a real PDF comes out and the error paths behave) rather than `go test`.

- `cmd/marc/main.go` — CLI entry point: flag parsing (`-t/--template`, `-i/--interactive`, `-o/--output`), wiring config/binary/template resolution together, and invoking `Render`. This is orchestration only; the actual logic lives in the files below.
- `cmd/marc/render.go` — the pandoc/Chromium pipeline itself: `RenderOptions` and `Render`. Reads the input file, preprocesses it, shells out to pandoc to produce self-contained HTML in a temp dir, shells out to headless Chromium to print that HTML to PDF, then moves the result to the output path.
- `cmd/marc/config.go` — resolves marc's XDG config directory (`ConfigDir`) and loads the optional `config.toml` from it (`LoadConfig`). Also embeds the starter config (`cmd/marc/config.example.toml`, via `//go:embed`) and `BootstrapDefaultConfig`, which writes it out on first run — see "First-time setup" below.
- `cmd/marc/binary.go` — resolves which pandoc/Chromium executable to run (`ResolveBinary`): a configured override from `config.toml` takes priority, otherwise it searches a list of candidate names on `PATH`.
- `cmd/marc/template.go` — template resolution and the interactive template picker: `ErrNoTemplate`, `TemplatesDir`, `ListTemplates`, `TemplatePath`, `IsInteractive`, `PromptTemplate`. Used by `main.go` when `-t` is omitted. Also embeds the starter template (`cmd/marc/template.html`, via `//go:embed`) and `BootstrapDefaultTemplate`, which writes it out on first run — see "First-time setup" below.
- `cmd/marc/paths.go` — `DefaultOutputPath`, the input-basename-with-`.pdf` fallback used when `-o` is omitted.
- `cmd/marc/preprocess.go` — Markdown preprocessing before it's handed to pandoc: `PreprocessNewpage` rewrites literal `\newpage` markers into pandoc's `::: pagebreak :::` fenced-div syntax (pandoc has no native page-break syntax), and `ExtractTitle` pulls the first `# ` heading out to use as the rendered document's `<title>`. This mirrors what the project's old `render.sh` script used to do by hand via `sed`/`grep` before `marc` replaced it.

## Dependencies

This project has exactly one third-party Go dependency: `github.com/BurntSushi/toml`, used by `config.go` to parse `config.toml`. Everything else is standard library. Don't add further dependencies without asking — see the global coding charter.

## Templates and config live outside this repo

Templates are **never** stored in this repository. `marc` reads them from the user's dotfiles, at:

```
$XDG_CONFIG_HOME/marc/templates/<name>/template.html
```

falling back to `~/.config/marc/templates/<name>/template.html` if `XDG_CONFIG_HOME` is unset. A "template" is a directory under `templates/` named `<name>` containing a `template.html` (a self-contained pandoc HTML template). `ListTemplates` discovers template names by scanning for such directories; `TemplatePath` resolves a chosen name to its `template.html` path.

The same config directory optionally holds a `config.toml` with three keys, all optional:

```toml
pandoc_bin = "/path/to/pandoc"
chromium_bin = "/path/to/chromium"
default_template = "default"
```

If `pandoc_bin`/`chromium_bin` is absent, `marc` falls back to searching `PATH` for a list of candidate binary names (`pandoc` for the pandoc binary; `chromium`, `chromium-browser`, `google-chrome`, `google-chrome-stable`, `brave-browser` in that order for the browser binary — see `chromiumCandidates` in `main.go`).

`default_template` controls what `-t`-omitted runs use (see the Usage section above). Its absence is meaningfully different from an explicit empty string: absent (no `config.toml`, or a `config.toml` predating this key) falls back to `"default"`; explicitly set to `""` means always prompt. `LoadConfig` distinguishes the two via `toml.MetaData.IsDefined`, exposed as `Config.DefaultTemplateSet`.

When extending or debugging `marc`, remember: nothing under `templates/` or a real `config.toml` should ever be added to this repo — those belong to the user's dotfiles, not to the tool's source. The one exception is the starter assets covered next, which live in the tool's source deliberately, as embedded resources rather than user config.

## First-time setup is automatic

A fresh checkout has no templates in the user's dotfiles yet. Rather than erroring, `marc` bootstraps: when `-t` is omitted and no templates exist at all in the templates dir, it writes an embedded starter template to `<templates-dir>/default/template.html`, writes an embedded starter `config.toml` (with `default_template = "default"` active, plus commented-out `pandoc_bin`/`chromium_bin` examples) to the config dir if one isn't already there, prints a note that it did so, and renders immediately using that template — no manual setup step required for the common case.

The source of these starter assets is tracked in the repo, unlike real templates/config: `cmd/marc/template.html` and `cmd/marc/config.example.toml`, embedded via `//go:embed` in `template.go`/`config.go` respectively (see `BootstrapDefaultTemplate`, `BootstrapDefaultConfig`). Edit those files to change what gets bootstrapped for new users.

This only fires when the templates dir has zero templates in it (the genuine fresh-install case). Passing an explicit `-t name` that doesn't exist behaves as before — a resolution error listing what's available.
