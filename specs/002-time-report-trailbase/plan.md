# Implementation Plan: Time Report Trailbase Migration

**Branch**: `002-time-report-trailbase` | **Date**: 2026-04-16 | **Spec**: [spec.md](spec.md)  
**Input**: Feature specification from `/specs/002-time-report-trailbase/spec.md`

## Summary

Migrate the time-report feature's data sources (monthly schedule, instructor salary
rates, and configuration constants) from hardcoded TypeScript files to Trailbase.
Three new database tables (`time_report_sessions`, `instructors`,
`time_report_config`) replace `time-report-items.json` and
`time-report-settings.ts`. The `tidrapport.astro` page becomes SSR. The
`send-time-report.ts` Worker endpoint fetches instructor rates at submission time
via a service user token. No personal data (names, emails, salary estimates) is
stored in the database — email remains the sole submission record.

## Technical Context

**Language/Version**: TypeScript 5.9 (strict mode)  
**Primary Dependencies**: Astro 5.17, @astrojs/cloudflare 12.6, Trailbase v0.26.3
REST API, Alpine.js 3.15, Bulma 1.0, Mailjet  
**Storage**: Trailbase SQLite on fly.io (region: arn / Stockholm) — three new tables  
**Testing**: Vitest (must be installed — see Constitution Gate II below)  
**Target Platform**: Cloudflare Workers (Edge Runtime, Node.js compatibility mode)  
**Project Type**: web-service (SSR Cloudflare Worker + static Astro pages)  
**Performance Goals**: Time report form loads in ≤ 3 s on mobile (SC-004)  
**Constraints**: TypeScript strict mode, zero `astro check` errors (SC-005);
`pnpm build` is the merge gate  
**Scale/Scope**: ~26 instructors, O(50) sessions per month; single-tenant

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-checked after Phase 1 design.*

### Gate I — Code Quality (Principle I)

| Check | Status | Action |
|-------|--------|--------|
| TypeScript strict mode | ⚠️ VIOLATION (pre-existing) | `salary.ts` uses `any` in two places — MUST be fixed as part of this feature |
| `Employee` type name | ⚠️ VIOLATION (pre-existing) | Rename `Employee` → `Instructor` in `types.ts`, `salary.ts`, `send-time-report.ts` |
| Dead code removal | REQUIRED | Remove `time-report-items.json`, `time-report-settings.ts`, hardcoded `EMPLOYEES` array after migration |

### Gate II — Testing (Principle II)

| Check | Status | Action |
|-------|--------|--------|
| `salary.ts` unit tests | ❌ MISSING | MUST have unit tests before new logic is added |
| `timeReportValidation.ts` unit tests | ❌ MISSING | MUST have unit tests |
| TODO(TEST_FRAMEWORK) | ❌ BLOCKING | Vitest MUST be installed as the first implementation task |

**Gate II blocks all `src/lib/` changes.** Vitest setup is Task 1.

### Gate III — UX (Principle III)

| Check | Status | Action |
|-------|--------|--------|
| Swedish UI text | ✓ | Error messages must be in Swedish (FR-010) |
| No inline styles | ⚠️ EXISTING | Pre-existing inline styles in `tidrapport.astro` (e.g., `style="cursor:pointer;"`) are out of scope for this feature |
| Alpine.js + Bulma only | ✓ | No new JS frameworks introduced |

### Gate IV — Performance (Principle IV)

| Check | Status | Action |
|-------|--------|--------|
| SSR justification | ✓ | `tidrapport.astro` requires Trailbase data at request time; `prerender = false` is mandatory |
| Cache-Control headers | REQUIRED | Set `Cache-Control: no-store` on `tidrapport.astro` (schedule must always be fresh); set correctly on the API endpoint |

### Gate V — Backend Architecture (Principle V)

| Check | Status | Action |
|-------|--------|--------|
| Trailbase sole backend | ✓ | All new data in Trailbase |
| Schema via migrations | ✓ | Three new migration SQL files |
| Worker delegates to REST API | ✓ | No raw SQL in `src/` |
| Trailbase built-in auth | ✓ | Service user token for all backend calls |
| Secrets via `wrangler secret put` | ✓ | `TRAILBASE_SERVICE_EMAIL` + `TRAILBASE_SERVICE_PASSWORD` added as new secrets |

### Gate VII — GDPR (Principle VII) — **TWO BLOCKING PRE-CONDITIONS**

| Check | Status | Action |
|-------|--------|--------|
| `/integritetspolicy` page | ❌ **BLOCKING** | MUST exist before go-live |
| `docs/gdpr-register.md` | ❌ **BLOCKING** | MUST exist before go-live |
| Legal basis for `instructors` | ✓ | Contractual necessity (employment) — stated in spec |
| Retention period | ✓ | "Until termination + 1 year" — MUST appear in migration comment |
| Deletion path | ✓ | Trailbase admin UI |
| Personal data not in source code | ✓ | Email/salary data removed from `send-time-report.ts` |
| No personal data in logs/URLs | ✓ | Email used only as lookup key, never logged |

**GDPR gate resolution**: Both GDPR blocking items are included as explicit tasks
in the implementation. The `instructors` table MUST NOT receive production data
until `/integritetspolicy` is live and `docs/gdpr-register.md` is committed.

### Constitution Check Summary

No constitution violations block the design itself. Two GDPR pre-conditions and the
Vitest requirement must be resolved as the first implementation tasks. All other
gates are satisfied by the design below.

## Architectural Decisions

### Service User Authentication

**Service user role**: Regular (non-admin) Trailbase user. All three tables
(`time_report_config`, `time_report_sessions`, `instructors`) use
"Authenticated" read access — meaning any Trailbase user created by an admin
can read them. Since Trailbase has no public registration, this is effectively
"admin-controlled access list", satisfying FR-004's intent of "not publicly
accessible".

**Authentication flow**: Service user logs in via `POST /api/auth/v1/token`
once per Worker invocation. The returned `auth_token` (JWT) is reused for
all subsequent Trailbase calls within that invocation. No KV caching is needed.

**New Worker secrets**: `TRAILBASE_SERVICE_EMAIL`, `TRAILBASE_SERVICE_PASSWORD`

### `salary.ts` Refactoring

`salary.ts` currently imports `timeReportItems` JSON and settings constants at
module level. After migration, all these values come from Trailbase. The refactor
converts the module to **pure functions** that accept schedule data and config
values as parameters — removing all top-level imports of config files. This also
fixes the `any` type violations.

The `findTimeItem` function signature changes from
`(section: string, value: string)` to
`(schedule: SessionSchedule, trainingGroup: string, value: string)` where `SessionSchedule`
is a typed map of `TrainingTrainingGroupKey → Session[]`.

### `tidrapport.astro` → SSR

The page gains `export const prerender = false`. Schedule and config are fetched
server-side via the extended `trailbase.ts` client. If the Trailbase call fails, the
page renders an error notice in Swedish instead of a broken form. The schedule data
is passed as props to `TimeReportCheckboxGroup` components (no change to the
component interface).

## Project Structure

### Documentation (this feature)

```text
specs/002-time-report-trailbase/
├── plan.md          ← this file
├── research.md      ← Phase 0 output
├── data-model.md    ← Phase 1 output
├── contracts/
│   └── trailbase-api.md  ← Phase 1 output
├── quickstart.md    ← Phase 1 output
└── tasks.md         ← Phase 2 output (/speckit.tasks)
```

### Source Code Changes

```text
src/
├── lib/
│   ├── trailbase.ts             EXTEND — add auth, schedule, config, instructor fns
│   ├── salary.ts                REFACTOR — pure fns, accept schedule + config params,
│   │                                       fix `any` types
│   ├── types.ts                 MODIFY — rename Employee→Instructor; add Session,
│   │                                     TrainingTrainingGroupKey, SessionSchedule, TimeReportConfig
│   └── timeReportValidation.ts  UNCHANGED
├── pages/
│   ├── tidrapport.astro         MODIFY — add prerender=false, SSR fetch, error state
│   ├── api/send-time-report.ts  MODIFY — remove hardcoded EMPLOYEES, fetch from TB
│   └── integritetspolicy.astro  CREATE — GDPR gate (minimal page)
└── env.d.ts                     MODIFY — add TRAILBASE_SERVICE_EMAIL, TRAILBASE_SERVICE_PASSWORD secrets

trailbase/
└── migrations/
    ├── U1776427200__create_time_report_sessions.sql  CREATE
    ├── U1776513600__create_instructors.sql           CREATE
    └── U1776600000__create_time_report_config.sql    CREATE

docs/
└── gdpr-register.md             CREATE — GDPR gate

src/config/
├── time-report-items.json       DELETE after migration
└── time-report-settings.ts      DELETE after migration

tests/
├── salary.test.ts               CREATE — unit tests
└── timeReportValidation.test.ts CREATE — unit tests

vitest.config.ts                 CREATE — Vitest setup
package.json                     MODIFY — add vitest dev dependency
```

## Complexity Tracking

| Situation | Justification |
|-----------|--------------|
| `prerender = false` on `tidrapport.astro` | Required by Principle IV/V: page must serve Trailbase data at request time; static generation is technically impossible for live schedule |
| Regular service user reading all three tables | All tables use "Authenticated" read; service user is non-admin, satisfying FR-004 (no public access) and FR-005 (Worker reads instructor rates) simultaneously |
