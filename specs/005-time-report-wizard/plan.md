# Implementation Plan: Two-Step Time Report Wizard

**Branch**: `005-time-report-wizard` | **Date**: 2026-04-26 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/005-time-report-wizard/spec.md`

## Summary

Converts the time report form into a two-step wizard: step 1 resolves the
instructor's identity via an email lookup (new server-side API endpoint); step
2 renders a role-scoped version of the existing form using Alpine.js conditional
rendering. The instructor data model gains a `travel_compensation` boolean and
the `swim_school_rate` NOT NULL constraint is relaxed to allow coach-only
instructors.

## Technical Context

**Language/Version**: TypeScript 5.9 (strict mode), Astro 5.17.1  
**Primary Dependencies**: Alpine.js 3.15.3, Bulma 1.0.4 + Sass, Trailbase 0.26.3  
**Storage**: Trailbase SQLite on fly.io (arn) — schema migration required  
**Testing**: Manual browser verification (`pnpm dev`); build gate (`pnpm build`)  
**Target Platform**: Cloudflare Workers (SSR pages and API routes)  
**Performance Goals**: Step 1 lookup completes in < 2 s on mobile; page load unchanged  
**Constraints**: No new client-side bundles; Alpine.js only for interactivity  
**Scale/Scope**: ~30 instructors; one active reporting month at a time

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-checked after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I — TypeScript strict | **PASS** | `swim_school_rate: number \| null` requires null guard in `salary.ts`; straightforward fix |
| II — Build gate | **PASS** | `pnpm build` is the merge gate; UI must be verified in browser |
| III — UX consistency | **PASS** | Bulma boxes, Alpine.js state machine, all text in Swedish |
| IV — Performance | **PASS** | One additional fetch per session (step 1 lookup); page load unaffected |
| V — Trailbase | **PASS** | Schema change via new migration file; new API endpoint uses service user token |
| VI — External data | **N/A** | No Idrottsarenan integration |
| VII — GDPR | **GATE — see below** | `travel_compensation` is personal data; endpoint must not log email |

**GDPR Gate (Principle VII)**:
- `travel_compensation` constitutes personal data (employment terms — whether
  the employer covers travel costs is a contractual term).
- Legal basis: Contractual necessity (Art. 6(1)(b) GDPR) — same as existing
  salary rate fields.
- Retention: Until end of employment + 1 year (same as salary rates).
- Deletion path: Admin deletes the instructor row via Trailbase admin UI
  (existing path — no new path needed).
- The new `/api/lookup-instructor` endpoint returns only role flags and
  `travel_compensation` — no salary rates exposed (data minimization).
- The endpoint MUST NOT log the email address in the Worker log output.
- TODO(GDPR_REGISTER): `travel_compensation` must be added to the GDPR register
  before go-live.

## Project Structure

### Documentation (this feature)

```text
specs/005-time-report-wizard/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── lookup-instructor.md
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code Changes

```text
trailbase/
└── migrations/
    └── U1776686400__update_instructors.sql   ← NEW migration

src/
├── lib/
│   ├── types.ts                              ← update Instructor interface
│   └── salary.ts                             ← null guard for swim_school_rate
├── pages/
│   ├── api/
│   │   ├── lookup-instructor.ts              ← NEW endpoint (step 1)
│   │   └── send-time-report.ts               ← minor: conditional field handling
│   └── tidrapport.astro                      ← major rewrite (two-step wizard)
```

No new components, no new npm packages, no new Cloudflare secrets.
