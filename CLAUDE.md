# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Website for Alvesta Simsällskap (founded 1921), a Swedish swim club.
Content-driven Astro site deployed on Cloudflare's edge network, backed by a
Trailbase service on fly.io. A standalone Go admin CLI manages backend data.

## Tech Stack

- **Frontend**: Astro 6.1.7 (static by default, per-page SSR via `export const prerender = false`)
- **CSS**: Bulma 1.0.4 + Sass (.scss) — the ONLY layout/component framework
- **Interactivity**: Alpine.js 3.15.3 — the ONLY client-side JS framework permitted
- **Icons**: Material Symbols Rounded (`@fontsource/material-symbols-rounded`)
- **Edge runtime**: Cloudflare Workers via `@astrojs/cloudflare` 13.1.10
- **Backend**: Trailbase v0.26.8 on fly.io (`arn` — Stockholm) — the ONLY database/persistence layer
- **Admin CLI**: Go 1.24.4 + Bubble Tea TUI (`tools/admin-cli/`)
- **Email**: Mailjet (transactional, time reports); **Bot protection**: Cloudflare Turnstile
- **Language**: TypeScript 5.9.3 (strict mode)
- **Secrets**: `wrangler secret put` — NEVER commit secrets to git

## Key Commands

```bash
pnpm dev          # wrangler types && astro dev
pnpm build        # astro check && astro build   (the merge gate — MUST pass)
pnpm preview      # wrangler types && astro preview
pnpm test         # vitest run  (tests in tests/*.test.ts)
```

Run a single frontend test: `pnpm vitest run tests/salary.test.ts` (add `-t "<name>"` to filter).

`wrangler types` regenerates `worker-configuration.d.ts` (Cloudflare runtime bindings). It runs
automatically in `dev`/`preview` but NOT in `build`; run it manually if bindings change.

### Admin CLI (Go)

```bash
cd tools/admin-cli
go test ./...                                   # unit tests (CSV parsers, normalisers, dedup)
go build -o alvestass-admin ./cmd/alvestass-admin
./alvestass-admin                               # interactive TUI; first run prompts for login
goreleaser build --clean --snapshot             # cross-platform release binaries → dist/
```

## Environment Secrets (Cloudflare Worker)

Set via `wrangler secret put <NAME>`; local dev values go in `.dev.vars` (git-ignored, see `.dev.vars.example`):

| Secret | Purpose |
|--------|---------|
| `MJ_APIKEY_PUBLIC` / `MJ_APIKEY_PRIVATE` | Mailjet keys |
| `TURNSTILE_SECRET_KEY` | Cloudflare Turnstile |
| `TRAILBASE_URL` | `https://alvestass-trailbase.fly.dev` |

## Architecture

### Rendering model
Pages are static-prerendered by default. SSR is opt-in per file via `export const prerender = false`
and is currently used only where live Trailbase data or form handling is required:
- `src/pages/kontakt.astro` — club contact info from Trailbase
- `src/pages/tidrapport.astro` — instructor time-report wizard
- `src/pages/api/send-time-report.ts` — Mailjet send endpoint
- `src/pages/api/lookup-instructor.ts` — instructor lookup for the wizard

Every SSR route must set correct `Cache-Control` headers and keep personal data out of URLs/logs.

### Content
Markdown content collections are defined in `src/content.config.ts` (glob loaders + Zod schemas)
across three collections: `club/`, `swim-school/`, `training-groups/`. Display order is driven by
an `order` frontmatter field. All user-visible text is in **Swedish**.

### Backend (Trailbase)
`src/lib/trailbase.ts` is the typed REST client; the Worker is the only thing that talks to Trailbase.
Schema lives in `trailbase/migrations/` (SQLite, naming `U{unix_timestamp}__{description}.sql`).
Migrations are baked into the Docker image and applied on startup — **never edit or rename an applied
migration; add a new file.** Deploy schema changes with `cd trailbase && fly deploy`, then verify in
the Trailbase admin UI.

### Admin CLI (`tools/admin-cli/`)
Single Go binary (Bubble Tea TUI) for administrators, using the official Trailbase Go SDK (auto token
refresh; credentials never persisted — only session tokens are stored locally at
`$UserConfigDir/alvestass-admin/config.json`, mode `0600`). Internal packages: `config`,
`trailbase` (SDK CRUD: sessions, members), `importer` (session CSV), `memberimporter` (IdrottOnline +
WeUnite member CSV import), `validate` (consistency checkers), `ui` (TUI screens). Importers do
parse → normalise (strip swim-school time slots, filter active members) → deduplicate on IdrottsID →
write via SDK.

### Key frontend logic (`src/lib/`)
`salary.ts` (salary estimation), `timeReportValidation.ts`, `email.ts` (Mailjet), `types.ts`.

## Architecture Rules (from constitution)

1. **TypeScript strict mode** — zero errors, no `any`; `astro check` MUST pass
2. **`pnpm build` is the merge gate** — failing build = blocking error
3. **Bulma + Sass only** — no inline styles, no other CSS frameworks
4. **Alpine.js only** — no React/Vue/other client-side JS frameworks
5. **Swedish** — all user-visible content (web + CLI) in Swedish
6. **Trailbase is the sole backend** — no other databases; REST API from Workers; SDK from the CLI; no raw SQL in Go
7. **SSR only when necessary** — justify each `prerender = false`; set `Cache-Control` correctly
8. **GDPR** — no personal data in URLs/logs; personal data behind auth

## GDPR Notes

- `/integritetspolicy` (`src/pages/integritetspolicy.astro`) and the Art. 30 register
  (`docs/gdpr-register.md`) now **exist** — keep the register updated whenever a table storing
  personal data is added.
- `src/content/club/04-styrelse.md` contains board member names (personal data) — flagged for
  future migration behind auth.
- **Member register (feature 010) is GATED for production**: the six member tables created by
  `trailbase/migrations/U1779235200__create_member_register.sql` must NOT be imported in production
  until (1) all six tables are documented in `docs/gdpr-register.md`, (2) they are configured as
  authenticated-only in the Trailbase admin UI, and (3) `/integritetspolicy` is live. Personnummer
  is used only transiently as a CSV linking key and is NEVER stored in Trailbase.

## Spec-Kit Workflow

Features are developed via spec-kit (`.specify/`, `specs/NNN-name/`): each has `spec.md` → `plan.md`
→ `tasks.md`. Use the `speckit-*` skills for that workflow. README must stay in sync when tech
stack/versions, dev commands, structure, admin-CLI features, build steps, SDK deps, or deployment
change — in the same task.

## Active Technologies
- TypeScript 5.9 (strict mode) + Astro 6.1.7, @astrojs/cloudflare 13.1, Trailbase v0.26.8, Alpine.js 3.15, Bulma 1.0
- Go 1.24.4 (admin CLI) + Bubble Tea / Bubbles / Lip Gloss, Trailbase Go SDK, `encoding/csv` (stdlib)
- Trailbase SQLite on fly.io (region: arn / Stockholm)

## Recent Changes
- 010-member-import: Member register schema (six tables) + admin-CLI member import (IdrottOnline + WeUnite CSV); production gated on GDPR pre-conditions
- 009-upgrade-trailbase: Trailbase upgraded v0.26.3 → v0.26.8; Go SDK updated
- 008-fix-ovrig-tid-minutes: time-report "övrig tid" minutes fix (pure client/worker logic)

<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan
at specs/010-member-import/plan.md
<!-- SPECKIT END -->
