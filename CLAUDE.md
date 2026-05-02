# Claude Code Context — alvestass.se

## Project Overview

Website for Alvesta Simsällskap (founded 1921), a Swedish swim club.
Content-driven Astro site deployed on Cloudflare's edge network.

## Tech Stack

- **Frontend**: Astro 5.17.1 (static by default, per-page SSR via `prerender = false`)
- **CSS**: Bulma 1.0.4 + Sass (.scss) — Bulma is the ONLY layout/component framework
- **Interactivity**: Alpine.js 3.15.3 — the ONLY client-side JS framework permitted
- **Icons**: Material Symbols Rounded (already included)
- **Edge runtime**: Cloudflare Workers via @astrojs/cloudflare 12.6.12
- **Backend**: Trailbase on fly.io (`arn` — Stockholm) — the ONLY database/persistence layer
- **Secrets manager**: `wrangler secret put` — NEVER commit secrets to git

## README

`README.md` must always be kept up to date. When any of the following change, update the README in the same task:
- Tech stack or versions
- Development commands
- Project structure
- Admin CLI features, build steps, or SDK dependencies
- Deployment process

## Key Commands

```bash
pnpm dev          # wrangler types && astro dev
pnpm build        # wrangler types && astro check && astro build  (MUST pass before merge)
pnpm preview      # astro preview
```

## Environment Secrets (Cloudflare Worker)

Set via `wrangler secret put <NAME>`:

| Secret | Purpose |
|--------|---------|
| `MJ_APIKEY_PUBLIC` | Mailjet public key |
| `MJ_APIKEY_PRIVATE` | Mailjet private key |
| `TURNSTILE_SECRET_KEY` | Cloudflare Turnstile |
| `TRAILBASE_URL` | `https://alvestass-trailbase.fly.dev` |

Local dev: add to `.dev.vars` (git-ignored).

## Project Structure

```
src/
├── components/       Astro components (PascalCase)
├── config/           Site settings, time report config
├── content/          Astro content collections (markdown)
│   ├── club/         Club info cards (05-kontakt.md has contact info — GDPR flag: 04-styrelse.md has personal data)
│   ├── swim-school/
│   └── training-groups/
├── lib/
│   ├── email.ts      Mailjet integration
│   ├── salary.ts     Salary calculation logic
│   ├── timeReportValidation.ts
│   ├── trailbase.ts  Trailbase REST client (typed)
│   └── types.ts
├── pages/
│   ├── api/send-time-report.ts   SSR — email endpoint
│   ├── kontakt.astro             SSR — club contact from Trailbase
│   └── index.astro               Static — main page
└── env.d.ts          Cloudflare runtime type declarations

trailbase/            Trailbase backend service
├── Dockerfile
├── fly.toml          Region: arn (Stockholm)
└── migrations/
    └── 0001_initial.sql
```

## Architecture Rules (from constitution)

1. **TypeScript strict mode** — zero errors, no `any`, `astro check` MUST pass
2. **`pnpm build` is the merge gate** — failing build = blocking error
3. **Bulma + Sass only** — no inline styles, no other CSS frameworks
4. **Alpine.js only** — no React, Vue, or other JS frameworks client-side
5. **Swedish text** — all user-visible content in Swedish
6. **Trailbase is the sole backend** — no other databases; use Trailbase REST API from Workers; secrets via `wrangler secret put`
7. **SSR only when necessary** — add `export const prerender = false` with explicit justification; set `Cache-Control` headers correctly
8. **GDPR** — no personal data in URLs/logs; personal data MUST be behind auth; `/integritetspolicy` page required before any personal data goes live

## Active Features

| Spec | Status | Key files |
|------|--------|-----------|
| 001-trailbase-backend-setup | In planning | specs/001-trailbase-backend-setup/ |

## GDPR Notes

- `src/content/club/04-styrelse.md` contains board member names (personal data) — not yet handled per Principle VII; flag for future migration
- `TODO(INTEGRITETSPOLICY)`: `/integritetspolicy` page must exist before any personal data feature goes live
- `TODO(GDPR_REGISTER)`: `docs/gdpr-register.md` must be created before any personal data is stored

## Active Technologies
- TypeScript 5.9 (strict mode) + Astro 5.17, @astrojs/cloudflare 12.6, Trailbase v0.26.3, Alpine.js 3.15, Bulma 1.0 (002-time-report-trailbase)
- Trailbase SQLite on fly.io (arn); three new tables (002-time-report-trailbase)
- Trailbase SQLite on fly.io (region: arn / Stockholm) — three new tables (002-time-report-trailbase)
- Go 1.23+ + `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/bubbles`, `github.com/charmbracelet/lipgloss`, `github.com/stretchr/testify` (tests only) (003-admin-cli)
- Local config file at `$UserConfigDir/alvestass-admin/config.json` (permissions `0600`); no local database (003-admin-cli)
- TypeScript 5.9 (strict mode); Astro 5.17.1 + Astro content collections, Alpine.js 3.15.3, Bulma 1.0.4 + Sass (004-frontend-mpa)
- No new storage; all content from existing `src/content/` collections (004-frontend-mpa)
- TypeScript 5.9 (strict mode), Astro 5.17.1 + Alpine.js 3.15.3, Bulma 1.0.4 + Sass, Trailbase 0.26.3 (005-time-report-wizard)
- Trailbase SQLite on fly.io (arn) — schema migration required (005-time-report-wizard)
- Go 1.24.4 (see `tools/admin-cli/go.mod`) + `github.com/charmbracelet/bubbletea` v1.1.1, `github.com/charmbracelet/bubbles` v0.20.0, `github.com/charmbracelet/lipgloss` v1.0.0, `github.com/trailbaseio/trailbase/client/go/trailbase` (Trailbase SDK), `encoding/csv` (stdlib) (006-import-sessions-csv)
- Trailbase SQLite on fly.io (arn) — `time_report_sessions` table (existing, no migration needed) (006-import-sessions-csv)
- TypeScript 5.9 (strict mode), Go 1.24.4 (admin CLI — not touched by this fix) + Astro 5.17.1, Alpine.js 3.15.3, Bulma 1.0.4 + Sass (008-fix-ovrig-tid-minutes)
- N/A — pure client/worker logic change, no Trailbase interaction (008-fix-ovrig-tid-minutes)

## Recent Changes
- 002-time-report-trailbase: Added TypeScript 5.9 (strict mode) + Astro 5.17, @astrojs/cloudflare 12.6, Trailbase v0.26.3, Alpine.js 3.15, Bulma 1.0
