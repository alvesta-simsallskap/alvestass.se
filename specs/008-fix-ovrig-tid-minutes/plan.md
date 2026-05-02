# Implementation Plan: Fix Övrig Tid Minutes Bug

**Branch**: `008-fix-ovrig-tid-minutes` | **Date**: 2026-04-30 | **Spec**: [spec.md](spec.md)  
**Input**: Feature specification from `/specs/008-fix-ovrig-tid-minutes/spec.md`

## Summary

The minutes field of an "Övrig tid" row in the time-report form is optional in practice — users frequently leave it blank when reporting whole-hour work — but the form-data parser requires it to be truthy. An empty string for `m` (or `h`) causes the entire row to be silently dropped. The fix tightens the validity check to require only `date` and `desc`, and coerces empty `h`/`m` strings to `"0"` before parsing. As a complementary UX improvement, the Alpine.js model for newly added rows will default both fields to `0` instead of `""`, preventing the issue from arising in the first place.

## Technical Context

**Language/Version**: TypeScript 5.9 (strict mode), Go 1.24.4 (admin CLI — not touched by this fix)  
**Primary Dependencies**: Astro 5.17.1, Alpine.js 3.15.3, Bulma 1.0.4 + Sass  
**Storage**: N/A — pure client/worker logic change, no Trailbase interaction  
**Testing**: No unit-test framework yet (TODO(TEST_FRAMEWORK)); changes MUST be verifiable by inspection and manual end-to-end test  
**Target Platform**: Cloudflare Workers (SSR API route) + browser (Alpine.js)  
**Project Type**: Web application (Astro/Cloudflare edge)  
**Performance Goals**: N/A — no hot path affected  
**Constraints**: `pnpm build` (wrangler types + astro check + astro build) MUST pass; TypeScript strict, zero `any`  
**Scale/Scope**: Two small files — `src/lib/timeReportValidation.ts` and `src/pages/tidrapport.astro`

## Constitution Check

| Principle | Gate | Status | Notes |
|-----------|------|--------|-------|
| I — Code Quality | TypeScript strict, zero errors, no `any`, no dead code | ✅ PASS | Fix is a small logic change; no new types introduced |
| II — Testing Standards | `pnpm build` must pass; `timeReportValidation.ts` must have unit tests; e2e manual test required for time-report workflow | ⚠️ CONDITIONAL | TODO(TEST_FRAMEWORK) is open — unit tests cannot be added yet. Fix MUST be simple enough to be verified by inspection. Manual e2e test of full form submission (including the "empty minutes" case) is **required** before marking done. |
| III — UX Consistency | UI change must be verified at mobile/tablet/desktop; Swedish text | ✅ PASS | Default-value change is invisible to the user (0 in a number field looks the same as empty); no new text or layout |
| IV — Performance | No new client-side bundles, static assets, or SSR paths | ✅ PASS | Trivial Alpine model init change |
| V — Backend Architecture | No Trailbase schema changes, no raw SQL | ✅ PASS | No backend changes |
| VI — External Integration | No Idrottsarenan data involved | ✅ PASS | N/A |
| VII — GDPR | No personal data introduced or re-processed | ✅ PASS | N/A |

**Conditional gate resolution**: The Principle II unit-test gate is deferred to TODO(TEST_FRAMEWORK). The change to `parseTimeReportForm` is a 3-line logic simplification with no branching complexity that cannot be verified by reading the code. Full end-to-end manual verification is mandatory before merge.

## Project Structure

### Documentation (this feature)

```text
specs/008-fix-ovrig-tid-minutes/
├── plan.md              # This file
├── research.md          # Phase 0 — no unknowns; minimal notes
├── data-model.md        # Phase 1 — no new entities
└── tasks.md             # Phase 2 output (/speckit.tasks — not yet created)
```

### Source Code (files changed)

```text
src/
├── lib/
│   └── timeReportValidation.ts   # FR-001..007: coerce empty h/m to "0", fix validity check
└── pages/
    └── tidrapport.astro          # UX: default h and m to 0 in new extraTimes rows
```

No new files, no new directories.

## Complexity Tracking

No constitution violations requiring justification.
