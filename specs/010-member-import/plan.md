# Implementation Plan: Member Register – Data Model and Initial Import

**Branch**: `010-member-import` | **Date**: 2026-05-19 | **Spec**: [spec.md](spec.md)  
**Input**: Feature specification from `specs/010-member-import/spec.md`

## Summary

Establish a Trailbase database schema for active club members, guardians, training groups, and family links, then build a one-shot admin CLI command that imports from two existing CSV exports (IdrottOnline and WeUnite Grupplista). The import deduplicates on IdrottsID (IID), strips time slots from swim-school group names, links guardians, and records family constellations. Production import is gated on GDPR pre-conditions (see Constitution Check).

## Technical Context

**Language/Version**: Go 1.24.4 (admin CLI import command); SQL (Trailbase SQLite migration)  
**Primary Dependencies**: `encoding/csv` (stdlib), `regexp` (stdlib), Trailbase Go SDK `v0.0.0-20260501081523-08228bb11ccf`  
**Storage**: Trailbase SQLite on fly.io (region: arn / Stockholm) — six new tables  
**Testing**: `go test` — unit tests for CSV parsers, group-name normaliser, and deduplication logic  
**Target Platform**: macOS / Linux (admin CLI); fly.io (Trailbase migration applied via Trailbase admin)  
**Project Type**: CLI tool extension + database schema migration  
**Scale/Scope**: ~270 active persons, ~30 training groups, ~350 guardian links (observed in source data)  
**Constraints**: WeUnite CSV uses semicolon delimiter and UTF-8 encoding; personnummer used transiently as a linking key between source files — never stored in Trailbase

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-checked after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I — Code Quality | ✅ Pass | All Go types explicit; no `any`; existing patterns followed |
| II — Testing | ✅ Pass | go test covers parsers and normalisation; `pnpm build` not affected |
| III — UX Consistency | ✅ Pass | CLI-only feature; Swedish text in all user-facing CLI output |
| IV — Performance | ✅ Pass | CLI tool; no Lighthouse scope |
| V — Trailbase | ✅ Pass | Schema via migration files; Trailbase Go SDK for all writes; no raw SQL in Go |
| VI — Idrottsarenan | ✅ Pass | Bulk CSV exports (not live API); minimum fields; purpose documented; IID is the federation-assigned identifier |
| VII — GDPR | ⚠️ **GATED** | Feature stores personal data — see below |

### GDPR Gate (Principle VII)

This feature introduces six new Trailbase tables containing personal data (names, contact details, dates of birth). Development of the schema and CLI tool may proceed; **the import MUST NOT be run in production until all pre-conditions below are satisfied.**

| Requirement | Status |
|-------------|--------|
| Legal basis documented in data-model.md | ✅ Done (Art. 6(1)(b) + (f)) |
| Data minimisation — only active members, minimum fields | ✅ Satisfied by design |
| Retention period in every migration comment | ✅ See data-model.md |
| Deletion path documented | ✅ CLI `delete-member` contract + CASCADE |
| Access control — data behind Trailbase auth | ⚠️ Planned in future auth feature; tables created without public API access |
| `docs/gdpr-register.md` updated for new tables | ⚠️ **Must add entries for all six tables before production import** |
| `/integritetspolicy` page published | ⚠️ `TODO(INTEGRITETSPOLICY)` — not yet created; **blocks production** |
| DPA: fly.io | ✅ Noted in existing gdpr-register.md |
| DPA: Cloudflare | ✅ Confirmed in existing gdpr-register.md |

## Project Structure

### Documentation (this feature)

```text
specs/010-member-import/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── contracts/
│   ├── cli-import-members.md     # Phase 1 output — CLI command spec
│   └── trailbase-member-tables.md # Phase 1 output — table access patterns
└── tasks.md             # Phase 2 output (/speckit-tasks — not created by /speckit-plan)
```

### Source Code (repository root)

```text
trailbase/migrations/
└── U1779235200__create_member_register.sql   # Six new tables

tools/admin-cli/
├── cmd/alvestass-admin/main.go               # Add MenuImportMembers case
├── internal/
│   ├── memberimporter/
│   │   ├── idrottonline.go   # IdrottOnline CSV parser
│   │   ├── weunite.go        # WeUnite Grupplista CSV parser
│   │   ├── normalize.go      # Group name normalization + active-member filter
│   │   ├── model.go          # Intermediate types (RawMember, RawGroup, etc.)
│   │   └── importer.go       # Orchestration: parse → merge → deduplicate → import
│   ├── trailbase/
│   │   ├── client.go         # (existing)
│   │   ├── sessions.go       # (existing)
│   │   └── members.go        # NEW: member/guardian/training-group CRUD via SDK
│   └── ui/
│       ├── import.go         # (existing — session import UI)
│       └── importmembers.go  # NEW: member import UI (file-path prompts, summary)

docs/
└── gdpr-register.md     # Add entries for the six new tables (pre-production gate)
```

**Structure Decision**: Single CLI binary extended with a new menu item and internal package. Follows the established pattern from `006-import-sessions-csv`.
