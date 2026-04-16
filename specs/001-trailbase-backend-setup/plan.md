# Implementation Plan: Trailbase Backend Setup (Minimal Starter)

**Branch**: `001-trailbase-backend-setup` | **Date**: 2026-04-16 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/001-trailbase-backend-setup/spec.md`

## Summary

Deploy a Trailbase instance on fly.io (Stockholm `arn`) as the project's canonical
backend. Create a single `club_info` table seeded with public organizational contact
data. Add a new SSR `/kontakt` page to alvestass.se that fetches and renders this
data, using Cloudflare edge caching for performance and stale-if-error for the
unavailability fallback. This proves the Astro → Cloudflare Workers → Trailbase
pipeline end-to-end with zero personal data exposure.

## Technical Context

**Language/Version**: TypeScript 5.9.3 / Astro 5.17.1
**Primary Dependencies**: @astrojs/cloudflare 12.6.12, Trailbase (Docker), fly.io CLI
**Storage**: Trailbase (SQLite via libSQL) on fly.io `arn` — 1 GB persistent volume
**Testing**: Manual browser verification (no automated test framework installed; TODO(TEST_FRAMEWORK))
**Target Platform**: Cloudflare Workers (Astro SSR) + fly.io free tier (Trailbase)
**Project Type**: Web application — hybrid static + SSR frontend + separate backend service
**Performance Goals**: SC-003 ≤ 3 s page load (mobile); edge-cached SSR responses will be sub-200 ms
**Constraints**: fly.io free tier (shared CPU, 256 MB RAM, 3 GB volume); Cloudflare Worker memory limits; Astro hybrid output mode required
**Scale/Scope**: Single `club_info` record; trivial read load (small Swedish swim club)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Gate | Status |
|-----------|------|--------|
| I. Code Quality | TypeScript strict mode — zero errors; `astro check` passes; no `any`; no dead code | ✅ New files (`trailbase.ts`, `kontakt.astro`) must be type-safe. Existing files untouched. |
| II. Testing Standards | `pnpm build` must pass; manual browser test of full update flow required | ✅ Build gate enforced. Manual test required (quickstart.md step 5). |
| III. UX Consistency | Swedish text; Bulma + Sass; mobile-first; no new JS frameworks | ✅ `/kontakt` page uses existing Bulma components and Swedish copy. No new JS library. |
| IV. Performance | SSR only where static generation is impossible; Lighthouse ≥ 90; correct cache headers | ✅ SSR justified (live backend data required). `Cache-Control: public, max-age=300, stale-if-error=86400` MUST be set. |
| V. Backend Architecture | Trailbase is sole backend; schema via migrations; Cloudflare Workers delegate to Trailbase REST; no custom auth | ✅ This feature IMPLEMENTS Principle V. Secrets via `wrangler secret put`. |
| VI. External Data | N/A — Trailbase is the club's own backend, not Idrottsarenan | ✅ Not applicable. |
| VII. GDPR | No personal data in initial data set; EU region (arn); no personal-data gate triggered | ✅ Data is organizational info only. `TODO(INTEGRITETSPOLICY)` and `TODO(GDPR_REGISTER)` not triggered by this feature. |

**Post-design re-check**: Hybrid output mode (`output: 'hybrid'`) is a minimal change to `astro.config.mjs` that affects no existing pages. Confirmed no violations introduced.

## Project Structure

### Documentation (this feature)

```text
specs/001-trailbase-backend-setup/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── club-info-api.md # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
trailbase/                          # NEW: Trailbase backend service
├── Dockerfile                      # Trailbase container image
├── fly.toml                        # fly.io config — region arn, volume mount
└── migrations/
    └── 0001_initial.sql            # club_info schema + Alvesta SS seed row

src/
├── lib/
│   └── trailbase.ts                # NEW: typed REST client for Trailbase
├── pages/
│   └── kontakt.astro               # NEW: SSR page — fetches + renders ClubInfo
└── env.d.ts                        # UPDATED: add TRAILBASE_URL

# Untouched — existing content collections and components unchanged
src/components/ClubInfo.astro       (no change)
src/content/club/                   (no change)
src/pages/index.astro               (no change)

astro.config.mjs                    # UPDATED: output: 'hybrid'
```

**Structure Decision**: Web application with a dedicated backend service directory (`trailbase/`) and a thin integration layer in the existing Astro `src/`. No new Astro project; the backend is an independently deployed service.

## Phase 0: Research

*See [research.md](research.md) for full findings. Summary:*

| Unknown | Decision | File |
|---------|----------|------|
| Trailbase REST endpoint path | `/api/collections/v1/club_info` (verify after deploy) | contracts/club-info-api.md |
| Data freshness mechanism | SSR + Cloudflare `max-age=300, stale-if-error=86400` | research.md §2 |
| Astro output mode change | `output: 'hybrid'` — minimal, no regression | research.md §3 |
| fly.io config for Trailbase | `shared-cpu-1x`, 256 MB, 1 GB volume, region `arn` | research.md §4 |
| Schema management | SQL migration file committed to git | research.md §5 |

## Phase 1: Design & Contracts

*See [data-model.md](data-model.md) and [contracts/club-info-api.md](contracts/club-info-api.md).*

### Key Design Decisions

**`src/lib/trailbase.ts`** — typed REST client:
- Exports `fetchClubInfo(baseUrl: string): Promise<ClubInfo>`
- Wraps `fetch()` with the correct endpoint path
- Returns `ClubInfo | null` (null if the record doesn't exist yet)
- Throws on network errors (let the caller decide on fallback)

**`src/pages/kontakt.astro`** — SSR contact page:
- `export const prerender = false`
- Fetches `ClubInfo` in the frontmatter; wraps in try/catch
- Sets `Cache-Control: public, max-age=300, stale-if-error=86400`
- If `clubInfo` is null or fetch failed, renders placeholder text (not an error page)
- UI: Bulma card with name, tagline, founding year, address block, phone, email, short description

**`trailbase/Dockerfile`**:
```dockerfile
FROM ghcr.io/trailbase-core/trail:latest
```
Trailbase reads its data directory from `/app/data` (fly.io volume mount point).

**`trailbase/fly.toml`** — key sections:
```toml
app = "alvestass-trailbase"
primary_region = "arn"

[build]
  dockerfile = "Dockerfile"

[mounts]
  source = "trailbase_data"
  destination = "/app/data"

[[services]]
  internal_port = 4000
  protocol = "tcp"
  [[services.ports]]
    port = 443
    handlers = ["tls", "http"]
```

### Constitution Check (post-design)

All principles still satisfied. The `output: 'hybrid'` change is backward-compatible. The `kontakt.astro` SSR page has explicit `Cache-Control` headers per Principle IV. No personal data introduced.
