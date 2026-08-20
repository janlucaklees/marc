# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repository is

`marc` is a small Go CLI that renders a Markdown file to PDF via `pandoc` (Markdown → self-contained HTML) followed by headless Chromium (HTML → PDF), using a named, dotfiles-managed HTML template. There is no application server, no web frontend, no Docker deployment here — just the CLI source, its tests, and the Makefile that builds/installs it.

Usage: `marc [-t|--template <name>] [-o|--output <path>] <input.md>`

- `-t/--template` selects a template by name. If omitted and the session is interactive, `marc` lists available templates and prompts for a choice; if omitted and non-interactive (e.g. piped/scripted), it errors rather than guessing.
- `-o/--output` sets the output PDF path. If omitted, it defaults to the input path with its extension replaced by `.pdf` (see `paths.go`).

## Commands

- `make build` — builds the binary to `bin/marc`.
- `make install` — builds, then copies the binary to `$HOME/.local/bin/marc`.
- `make test` (or `go test ./...`) — runs all package tests.

These Makefile targets are the only defined workflow commands — use them rather than invoking `go build`/`go install` directly.

## Architecture

Everything lives in package `main` at the repo root, one `.go` file per concern. Five of the seven have a sibling `_test.go` (`binary.go`, `config.go`, `paths.go`, `preprocess.go`, `template.go`) — pure logic covered by unit tests. `main.go` and `render.go` deliberately have no `_test.go`: they're thin wrappers around external processes (pandoc, Chromium) and CLI wiring, and are instead verified by manual end-to-end testing (build the binary, render a real Markdown file, confirm a real PDF comes out and the error paths behave) rather than `go test`.

- `main.go` — CLI entry point: flag parsing (`-t/--template`, `-o/--output`), wiring config/binary/template resolution together, and invoking `Render`. This is orchestration only; the actual logic lives in the files below.
- `render.go` — the pandoc/Chromium pipeline itself: `RenderOptions` and `Render`. Reads the input file, preprocesses it, shells out to pandoc to produce self-contained HTML in a temp dir, shells out to headless Chromium to print that HTML to PDF, then moves the result to the output path.
- `config.go` — resolves marc's XDG config directory (`ConfigDir`) and loads the optional `config.toml` from it (`LoadConfig`).
- `binary.go` — resolves which pandoc/Chromium executable to run (`ResolveBinary`): a configured override from `config.toml` takes priority, otherwise it searches a list of candidate names on `PATH`.
- `template.go` — template resolution and the interactive template picker: `ErrNoTemplate`, `TemplatesDir`, `ListTemplates`, `TemplatePath`, `IsInteractive`, `PromptTemplate`. Used by `main.go` when `-t` is omitted.
- `paths.go` — `DefaultOutputPath`, the input-basename-with-`.pdf` fallback used when `-o` is omitted.
- `preprocess.go` — Markdown preprocessing before it's handed to pandoc: `PreprocessNewpage` rewrites literal `\newpage` markers into pandoc's `::: pagebreak :::` fenced-div syntax (pandoc has no native page-break syntax), and `ExtractTitle` pulls the first `# ` heading out to use as the rendered document's `<title>`. This mirrors what the project's old `render.sh` script used to do by hand via `sed`/`grep` before `marc` replaced it.

## Dependencies

This project has exactly one third-party Go dependency: `github.com/BurntSushi/toml`, used by `config.go` to parse `config.toml`. Everything else is standard library. Don't add further dependencies without asking — see the global coding charter.

## Templates and config live outside this repo

Templates are **never** stored in this repository. `marc` reads them from the user's dotfiles, at:

```
$XDG_CONFIG_HOME/marc/templates/<name>/template.html
```

falling back to `~/.config/marc/templates/<name>/template.html` if `XDG_CONFIG_HOME` is unset. A "template" is a directory under `templates/` named `<name>` containing a `template.html` (a self-contained pandoc HTML template). `ListTemplates` discovers template names by scanning for such directories; `TemplatePath` resolves a chosen name to its `template.html` path.

The same config directory optionally holds a `config.toml` with two keys, both optional:

```toml
pandoc_bin = "/path/to/pandoc"
chromium_bin = "/path/to/chromium"
```

If a key is absent, `marc` falls back to searching `PATH` for a list of candidate binary names (`pandoc` for the pandoc binary; `chromium`, `chromium-browser`, `google-chrome`, `google-chrome-stable`, `brave-browser` in that order for the browser binary — see `chromiumCandidates` in `main.go`).

When extending or debugging `marc`, remember: nothing under `templates/` or a real `config.toml` should ever be added to this repo — those belong to the user's dotfiles, not to the tool's source.
