# Alvesta Simsällskap — alvestass.se

Website for Alvesta Simsällskap (founded 1921). Content-driven Astro site on Cloudflare's edge network.

## Tech stack

- **Astro 5** — static by default, per-page SSR via Cloudflare Workers
- **Bulma 1 + Sass** — sole CSS/layout framework
- **Alpine.js 3** — sole client-side JS framework
- **Trailbase** — backend on fly.io (Stockholm), sole persistence layer
- **Mailjet** — transactional email (time reports)
- **Cloudflare Turnstile** — bot protection on forms

## Development

```bash
pnpm install
pnpm dev        # wrangler types && astro dev
pnpm build      # wrangler types && astro check && astro build
pnpm preview    # astro preview
```

Local secrets go in `.dev.vars` (git-ignored). See `CLAUDE.md` for the full secret list.

## Project structure

```
src/
├── components/        Astro components (PascalCase)
├── config/            Site settings, time-report config
├── content/           Markdown content collections
│   ├── club/
│   ├── swim-school/
│   └── training-groups/
├── lib/               Shared logic (trailbase client, email, salary calc)
└── pages/             Astro pages + API routes

trailbase/             Backend service (Dockerfile, fly.toml, migrations)
tools/
└── admin-cli/         Standalone admin CLI (Go) — see below
```

## Admin CLI

A standalone Go executable for administrators to manage backend data without requiring direct database access.

**Features (v1):**
- Interactive update of club contact information
- Consistency check with extensible `Checker` interface (future: member groups, time-report coverage)
- First-run wizard stores session tokens locally; credentials are never persisted
- Uses the [official Trailbase Go SDK](https://github.com/trailbaseio/trailbase/tree/main/client/go/trailbase) for automatic token refresh

**Platforms:** macOS (Apple Silicon + Intel), Windows 10/11 — single binary, no installation.

```bash
cd tools/admin-cli
go mod tidy
go build -o alvestass-admin ./cmd/alvestass-admin
./alvestass-admin
```

Cross-platform release builds:

```bash
brew install goreleaser
goreleaser build --clean --snapshot
# Binaries land in tools/admin-cli/dist/
```

See [`specs/003-admin-cli/quickstart.md`](specs/003-admin-cli/quickstart.md) for the full manual test procedure.

## Calendars

Static `.ics` files in `public/`, with a `_headers` file controlling `Content-Type`.

## Deployment

Cloudflare Pages + Workers. Secrets are set via:

```bash
wrangler secret put <NAME>
```
