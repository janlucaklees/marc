# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repository is

This is **not** the Renaissance application source code. Renaissance (a clinical-trial DQA ticketing system) lives in a separate, private GitHub repo (`janlucaklees/renaissance`), built as Docker images and published to `ghcr.io`. This repository holds the **documentation and deployment artifacts** for that system:

- German-language Markdown docs, rendered to HTML/PDF for handoff to CRO IT/compliance.
- The `docker-compose.yaml` used to deploy prebuilt Renaissance images on a target host.

## Commands

**Render a doc to HTML/PDF:**

```
./render.sh
```

This is the only defined workflow command — always use it rather than invoking `pandoc`/`chromium` directly. It:
1. Runs `Readme.md` through `pandoc` with `template.html`, producing `output.html` (self-contained, embedded resources).
2. Renders `output.html` to `output.pdf` via headless `chromium`.
3. Rewrites `\newpage` markers into pandoc `::: pagebreak :::` divs before conversion (a workaround for `\newpage` not being native Markdown/pandoc syntax).

`INPUT` is hardcoded to `Readme.md` inside the script. To render `Readme (1).md` (the Installationsprotokoll) instead, temporarily point `INPUT` at it or copy it over `Readme.md` before running — don't add a second script; ask before changing `render.sh` itself since it's a shared tool.

## Layout

- `Readme.md` — **Systemdokumentation Renaissance**: the source-of-truth architecture/ops doc (purpose, architecture diagram, tech stack, data model, hosting requirements, security posture, backup status, validation status). This is the primary file to edit when system facts change.
- `Readme (1).md` — **Installationsprotokoll**: a dated log of one specific VM deployment (Photon OS/VMware host at the CRO). Treat as a historical record, not a living doc — don't retrofit it when the system changes; add a new dated protocol instead.
- `template.html` — pandoc HTML template used by `render.sh`. Self-contained styling for the rendered doc.
- `template.tex` — a pandoc LaTeX template; present but **not** wired into `render.sh` (which renders via the HTML→chromium path). Don't assume it's live tooling.
- `docker-compose.yaml` / `.env` — the production deployment stack: Traefik (reverse proxy, sole external entrypoint), Postgres, `renaissance` (backend, `ghcr.io/janlucaklees/renaissance`), `frontend` (SvelteKit SSR, `ghcr.io/janlucaklees/renaissance-frontend`). Pulls versioned prebuilt images — nothing here is built from source. `.env` in this repo holds placeholder (`changeme`) values, not real secrets.
- `output.html` / `output.pdf` / `Systemdokumentation Renaissance (vorläufig).pdf` / `Installationsprotokoll ... .pdf` — rendered outputs of the two Markdown docs at various points in time. Regenerate via `render.sh` rather than hand-editing.
- `.attachments.1259797/`, `attachments.1259797.zip` — screenshots referenced by `Readme (1).md`.
- `goal.pdf` — reference PDF showing the intended rendered look/format for the Installationsprotokoll (compare against when adjusting `template.html`).

## Conventions

- Docs are written in German; keep additions consistent with existing terminology (e.g. "Klärungsbedarf", "Studien", "Teilnehmer") rather than introducing English mid-document.
- Renaissance versions are referenced by both semver tag (e.g. `v0.12.1`) and commit hash — keep both in sync when updating version references in `Readme.md`.
- Rendered `.pdf`/`.html` files are build artifacts of the corresponding `.md` source; when a Markdown doc changes, regenerate its rendered output via `render.sh` rather than editing the PDF/HTML by hand.
