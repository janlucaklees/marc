# marc: Markdown → PDF CLI — Design

**Status:** Approved (design), pending implementation plan
**Date:** 2026-08-20

## Purpose

Turn the Markdown→PDF rendering pipeline currently hardcoded in `render.sh` (pandoc → HTML → headless Chromium → PDF) into a general-purpose, installable Linux CLI tool: `marc`. It must accept any Markdown file and a named template (dotfiles-managed, not shipped with the tool), so the same binary can render, e.g., a work-branded PDF and a private-branded PDF from different template directories.

This repurposes the current `marc` project folder, which previously held Renaissance-specific documentation and deployment artifacts (see "Cleanup" below) unrelated to this tool.

## Non-goals (v1)

- Multi-file batch input (`marc a.md b.md`) — single file only.
- Per-template settings beyond the HTML template itself (page size, pandoc metadata defaults, etc.) — the template directory exists to allow this later without a format change, but v1 only reads `template.html` from it.
- CI/non-interactive-first design — the interactive template prompt is a convenience for local use; non-interactive sessions are handled by erroring clearly, not by special-casing CI.
- Packaging beyond `make install` (no GoReleaser, no distro packaging).

## CLI interface

```
marc [-t|--template <name>] [-o|--output <path>] <input.md>
```

- `<input.md>` — required, positional.
- `-t/--template <name>` — selects `~/.config/marc/templates/<name>/template.html`. If omitted:
  - If stdin is a TTY: list available template names (subdirectories of the templates dir) and prompt for a choice.
  - If stdin is not a TTY: error, telling the user to pass `-t` explicitly.
- `-o/--output <path>` — overrides the default output location (`<input-basename>.pdf` next to the input file).

## Config layout

XDG-style, dotfiles-managed, not part of this repo:

```
$XDG_CONFIG_HOME/marc/            (falls back to ~/.config/marc/)
  config.toml                     # optional
  templates/
    work/template.html
    private/template.html
    .../template.html
```

`config.toml` (all fields optional, v1 scope limited to binary overrides):

```toml
pandoc_bin = "/usr/bin/pandoc"       # optional override; default: search PATH
chromium_bin = "/usr/bin/chromium"   # optional override; default: search PATH
```

## Components (Go, single module)

- **`main.go`** — flag parsing, orchestration, top-level error reporting.
- **`config.go`** — loads and parses `config.toml`; resolves the config directory via `XDG_CONFIG_HOME`/`~/.config`.
- **`template.go`** — resolves a template by name; lists available templates; handles the interactive prompt (TTY check + selection) when no template is given.
- **`render.go`** — the pipeline itself:
  1. Preprocess input Markdown: rewrite `\newpage` markers into pandoc `::: pagebreak :::` divs (ported from `render.sh`'s `sed` step).
  2. Run `pandoc` with the resolved `template.html`, `--embed-resources --standalone`, writing to a temp HTML file.
  3. Run headless Chromium (`--headless --disable-gpu --no-sandbox --run-all-compositor-stages-before-draw --no-pdf-header-footer --print-to-pdf-no-header --print-to-pdf=<temp-or-final>`) against the temp HTML file.
  4. Move the resulting PDF to the final output path; delete the temp HTML file (best-effort, even on error paths where it exists).

## Error handling

Each of these is a distinct, actionable error message (not a raw exec error dump):

- Input file missing/unreadable.
- Named template not found → lists available template names.
- No template given + stdin not a TTY.
- `pandoc`/`chromium` not found (checks configured path first, then PATH) → names which binary and where it looked.
- `pandoc`/`chromium` exits non-zero → surfaces their stderr output.
- Output path unwritable (e.g. directory doesn't exist).

## Build & install

`Makefile` at repo root:
- `make build` — `go build -o bin/marc ./...`
- `make install` — builds, then copies `bin/marc` to `~/.local/bin/marc`

Repo root becomes the Go module root (`go.mod`).

## Testing

`go test` covers pure-logic units only:
- `\newpage` → pagebreak-div preprocessing transform.
- Template resolution (found / not found / listing).
- Config parsing (present / absent / partial `config.toml`).

The pandoc/Chromium shell-outs are not unit-tested — they're thin wrappers around external processes, verified by running the tool for real. No end-to-end test harness.

## Cleanup sequencing

1. **Now (pre-implementation, already confirmed with user):** delete Renaissance-specific content no longer relevant to this repo: `Readme.md`, `Readme (1).md`, `docker-compose.yaml`, `.env`, `output.html`, `Systemdokumentation Renaissance (vorläufig).pdf`, `Installationsprotokoll Renaissance zweiter Teil vom 30.07.2026.pdf`, `.attachments.1259797/`, `attachments.1259797.zip`, `935-101.pdf` (unrelated personal file).
2. **During implementation:** port `template.html`'s content into a dotfiles template directory (outside this repo) as the first named template (e.g. `work` or `default` — user to place it).
3. **After `marc` fully replaces the old pipeline and is verified working:** delete `render.sh`, `template.tex`, `goal.pdf`, `.playwright-mcp/` from this repo.

## Open items for the implementation plan

- Exact Chromium binary name search order for auto-detect (`chromium`, `chromium-browser`, `google-chrome`, `brave-browser`, ...).
- Whether `go.mod` module path matters (this is a personal single-binary tool, not published — a simple local module name is fine unless the user wants otherwise).
