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
