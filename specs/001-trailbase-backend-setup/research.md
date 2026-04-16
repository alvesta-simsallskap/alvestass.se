# Research: Trailbase Backend Setup (Minimal Starter)

**Feature**: 001-trailbase-backend-setup
**Date**: 2026-04-16
**Status**: Complete

---

## Decision 1: Trailbase REST API Access Pattern

**Decision**: Use Trailbase's auto-generated Records REST API with a public (unauthenticated) read access policy on the `club_info` table.

**Endpoint pattern** (verify exact path from Trailbase admin UI after deployment):
```
GET https://<app>.fly.dev/api/collections/v1/club_info
```
Single record by ID:
```
GET https://<app>.fly.dev/api/collections/v1/club_info/<record-id>
```

**Access policy**: Trailbase allows per-table read/write access rules. The `club_info` table MUST be configured with an open read policy (no JWT required) so the Cloudflare Worker can fetch without credentials. Write access remains admin-only.

**Rationale**: The data is entirely non-personal organizational contact info; requiring an API key for reads adds complexity with no security benefit. Admin writes are always protected by Trailbase's built-in auth.

**Alternative considered**: Require a read API key — rejected. It would require rotating secrets and storing them in Cloudflare Worker secrets for data that is already public.

---

## Decision 2: Data Freshness Strategy (SC-002: ≤5 min propagation)

**Decision**: SSR the `/kontakt` page with Cloudflare edge cache headers:
```
Cache-Control: public, max-age=300, stale-if-error=86400
```

**How it works**:
- Cloudflare caches the rendered page at the edge for 5 minutes (`max-age=300`)
- After 5 minutes the next request triggers a fresh Trailbase fetch and re-caches
- If Trailbase is unreachable, Cloudflare serves the previous cached version for up to 24 hours (`stale-if-error=86400`) — satisfying the "show last known data" fallback

**Rationale**: Meets SC-002 (≤5 min), SC-003 (fast cached loads), and the fallback edge case with a single header. No client-side JS needed, no Alpine.js fetch, no KV store.

**Alternative considered**: Client-side Alpine.js fetch to a `/api/club-info` endpoint — rejected. Adds a client-side network request, requires CORS configuration on Trailbase, and contradicts Principle IV (minimize client-side JS).

**Alternative considered**: Cloudflare KV with periodic Worker update — rejected. Significantly more infrastructure for a PoC with no material benefit over edge caching.

---

## Decision 3: Astro Output Mode

**Decision**: Change `astro.config.mjs` to `output: 'hybrid'`. Add `export const prerender = false` to `/kontakt.astro` only.

**Rationale**: `output: 'hybrid'` makes all existing pages static by default — zero regression on performance, build, or caching for the rest of the site. Only the new Kontakt page opts into SSR, which is justified because it serves live backend data (constitution Principle IV gate: "SSR only where static generation is technically impossible").

**Alternative considered**: `output: 'server'` — rejected. Makes the entire site SSR, breaking static generation for all existing pages.

---

## Decision 4: fly.io Deployment for Trailbase

**Decision**: Use the official Trailbase Docker image with a fly.io `arn` (Stockholm) persistent volume for SQLite storage.

**fly.io setup summary**:
- App name: `alvestass-trailbase` (or similar)
- Region: `arn` (Stockholm) — GDPR data-residency (FR-007)
- Machine: `shared-cpu-1x` with 256 MB RAM (free tier)
- Volume: 1 GB persistent volume mounted at `/app/data` for the SQLite database
- HTTP service on port `4000` (Trailbase default)

**Dockerfile approach**: Use `ghcr.io/trailbase-core/trail:latest` and configure via environment variables or a `traildepot/` config directory.

**Rationale**: fly.io free tier with arn region satisfies the spec constraints (FR-007, Assumptions). The official Docker image minimises configuration surface.

**Volume mounting note**: The SQLite file MUST be on a persistent fly.io volume — if the machine restarts without a volume, all data is lost. The volume is the single most critical deployment step.

---

## Decision 5: Trailbase Schema Management

**Decision**: Define the initial schema as a SQL migration file at `trailbase/migrations/0001_initial.sql`. Apply via the Trailbase admin UI or CLI during first deployment.

**Rationale**: Constitution Principle V requires all schema changes through Trailbase's migration system. Having the SQL under version control satisfies this even before formal migration tooling is wired up.

**Note on Trailbase migrations**: Trailbase manages schema through its internal migration runner. The SQL file is the canonical source of truth checked into git; the actual execution happens inside the Trailbase admin UI or by placing the file in the `traildepot/migrations/` directory that Trailbase scans on startup.

---

## Decision 6: Source Code Layout

The Trailbase backend configuration lives in a top-level `trailbase/` directory (separate from the Astro `src/`). The Astro website gets a thin typed client module.

```
trailbase/                        # Backend service — fly.io + schema
├── fly.toml                      # fly.io app config (region arn, volume mount)
├── Dockerfile                    # Trailbase container
└── migrations/
    └── 0001_initial.sql          # club_info table + seed row

src/
├── lib/
│   └── trailbase.ts              # NEW: typed REST client for Trailbase
├── pages/
│   └── kontakt.astro             # NEW: SSR page — fetches + renders ClubInfo
└── env.d.ts                      # UPDATED: add TRAILBASE_URL
```

**Existing files untouched**: `src/components/ClubInfo.astro`, `src/content/club/`, `src/pages/index.astro` — these content-collection-based cards continue to render existing markdown content unchanged.

---

## GDPR Flag (non-blocking for this PoC)

The existing file `src/content/club/04-styrelse.md` contains personal data (board member names). This is not in scope for this PoC but MUST be addressed in a future feature per constitution Principle VII. Flagged here for the processing register (TODO: GDPR_REGISTER).
