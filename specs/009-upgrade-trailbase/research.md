# Research: Upgrade Trailbase to v0.26.8

**Date**: 2026-05-05  
**Feature**: 009-upgrade-trailbase

## Release Notes Summary (v0.26.4 – v0.26.8)

Source: https://github.com/trailbaseio/trailbase/releases

| Version | Date | Key Changes | Breaking? |
|---------|------|-------------|-----------|
| v0.26.8 | 2026-04-30 | Experimental PostgreSQL connection types (additive only); `AsyncReactive` primitive; dependency updates | No |
| v0.26.7 | 2026-04-28 | CLI regression fix for `user`/`admin` commands; `--version` fix for CI builds; new `&skip_cursor` listing param; geospatial admin UI fixes | No |
| v0.26.6 | 2026-04-23 | Admin UI WKT geometry support; `execute_batch()` added internally; dependency updates | No |
| v0.26.5 | 2026-04-22 | Major internal refactor of `trailbase_sqlite::Connection` (async transaction API, `SyncConnection`, dedicated backup API, new `refinery` driver); statement cache size increase | No |
| v0.26.4 | 2026-04-17 | WASM transaction model improved; rate limits and body size limits now configurable; admin UI account timestamps; `kanal` → `flume` switch | No |

## Impact Analysis for This Project

| Area | Impact | Detail |
|------|--------|--------|
| REST API (`src/lib/trailbase.ts`) | None | No API contract changes in any release |
| Authentication | None | No auth changes |
| SQL schema / migrations | None | Migration runner internals changed in v0.26.5 but format/interface is identical |
| Go client SDK | Update needed | SDK commit advances from 2026-04-21 to 2026-05-01 snapshot |
| Docker image | Redeploy only | `trailbase/trailbase:latest` already in Dockerfile; just redeploy |
| Admin CLI | Build + test | `go build ./...` and `go test ./...` must pass with updated SDK |
| CLAUDE.md | Text update | Version string `0.26.3` → `0.26.8` |

## Go Client SDK Update

- **Previous**: `v0.0.0-20260421205927-716548d63a07` (2026-04-21 snapshot)
- **New**: `v0.0.0-20260501081523-08228bb11ccf` (2026-05-01 snapshot, corresponds to post-v0.26.8)
- **How resolved**: `go get github.com/trailbaseio/trailbase/client/go/trailbase@main`
- **go.mod**: Already updated (the `go get` ran during research phase)
- **go.sum**: Updated automatically by `go get`

## Decisions

### D1: No Dockerfile change needed
**Decision**: Keep `trailbase/trailbase:latest` as-is.  
**Rationale**: The Dockerfile already tracks latest. Redeploying on fly.io (`fly deploy`) triggers a fresh pull of the latest image, which is v0.26.8.  
**Alternatives considered**: Pinning to a specific tag (e.g., `trailbase/trailbase:0.26.8`) for reproducibility. Rejected because there is no evidence the image is published with version tags; the project has always used `:latest`.

### D2: No schema migration needed
**Decision**: No new migration file.  
**Rationale**: Zero schema changes in v0.26.4–v0.26.8. Existing migrations in `trailbase/migrations/` remain valid.

### D3: No data backup procedure in scope
**Decision**: Rely on fly.io volume persistence.  
**Rationale**: The `trailbase_data` volume is decoupled from the Docker image. A `fly deploy` swaps only the container binary, not the volume. The upgrade risk is equivalent to a process restart.  
**Recommendation**: Taking a fly.io volume snapshot before deploying is good operational practice but is out of scope for this spec (no data at risk from a backwards-compatible binary upgrade).
