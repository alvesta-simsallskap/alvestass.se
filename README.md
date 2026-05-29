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
pnpm build      # astro check && astro build  (must pass before merge)
pnpm preview    # wrangler types && astro preview
pnpm test       # vitest run
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

## Trailbase backend

The Trailbase instance runs on fly.io (region: `arn` / Stockholm) and is the sole persistence layer.

### Deploying changes

Any time you add or modify a migration file in `trailbase/migrations/`, redeploy the backend:

```bash
cd trailbase
fly deploy
```

This rebuilds the Docker image with the updated migration files baked in, deploys it to fly.io, and Trailbase automatically applies any pending migrations on startup.

### How migrations work

- Migration files live in `trailbase/migrations/` and follow Trailbase's naming convention (`U{unix_timestamp}__{description}.sql`).
- The Dockerfile copies them into the image at `/app/seed-migrations/`.
- On startup, `docker-entrypoint.sh` copies any new files to `/app/traildepot/migrations/` (existing files are never overwritten), then Trailbase runs all pending migrations.
- Never edit or rename an already-applied migration — add a new file instead.

### Adding a schema change

1. Create a new file: `trailbase/migrations/U{timestamp}__{description}.sql`
2. Write the migration SQL (Trailbase uses SQLite; wrap destructive changes in a transaction).
3. Run `cd trailbase && fly deploy`.
4. Verify in the Trailbase admin UI that the migration was applied.

### Admin UI

The Trailbase admin UI is available at the fly.io app URL. Use it to inspect tables, manage instructor records, and verify migrations.

## Deployment

Cloudflare Pages + Workers. Secrets are set via:

```bash
wrangler secret put <NAME>
```
