# Data Model: Time Report Trailbase Migration

**Feature**: 002-time-report-trailbase
**Date**: 2026-04-16

---

## Entity: TimeReportConfig

**Storage**: Trailbase table `time_report_config` (SQLite STRICT)
**Cardinality**: Exactly one row
**Access**: Authenticated read (service user token); admin-only write
**GDPR**: No personal data

### Fields

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | INTEGER | PRIMARY KEY | Single row, id = 1 |
| `active_month_key` | TEXT | NOT NULL | ISO month, e.g. `2026-04` |
| `active_month_display` | TEXT | NOT NULL | Swedish name, e.g. `april 2026` |
| `extra_time_simskola` | INTEGER | NOT NULL DEFAULT 30 | Minutes added per swim school session |
| `extra_time_training` | INTEGER | NOT NULL DEFAULT 15 | Minutes added per training session |
| `half_day_salary` | INTEGER | NOT NULL DEFAULT 500 | SEK for half-day competition |
| `full_day_salary` | INTEGER | NOT NULL DEFAULT 1000 | SEK for full-day competition |
| `overnight_salary` | INTEGER | NOT NULL DEFAULT 300 | SEK for overnight stay |

### SQL Migration

```sql
-- trailbase/migrations/U1776600000__create_time_report_config.sql
-- GDPR: No personal data — operational configuration only

CREATE TABLE IF NOT EXISTS time_report_config (
  id                    INTEGER PRIMARY KEY,
  active_month_key      TEXT    NOT NULL CHECK(active_month_key != ''),
  active_month_display  TEXT    NOT NULL CHECK(active_month_display != ''),
  extra_time_simskola   INTEGER NOT NULL DEFAULT 30,
  extra_time_training   INTEGER NOT NULL DEFAULT 15,
  half_day_salary       INTEGER NOT NULL DEFAULT 500,
  full_day_salary       INTEGER NOT NULL DEFAULT 1000,
  overnight_salary      INTEGER NOT NULL DEFAULT 300
) STRICT;

INSERT INTO time_report_config (
  id, active_month_key, active_month_display
) VALUES (
  1, '2026-04', 'april 2026'
);
```

### TypeScript Interface

```typescript
export interface TimeReportConfig {
  id: number;
  active_month_key: string;       // e.g. "2026-04"
  active_month_display: string;   // e.g. "april 2026"
  extra_time_simskola: number;    // minutes
  extra_time_training: number;    // minutes
  half_day_salary: number;        // SEK
  full_day_salary: number;        // SEK
  overnight_salary: number;       // SEK
}
```

---

## Entity: TimeReportSession

**Storage**: Trailbase table `time_report_sessions` (SQLite STRICT)
**Cardinality**: Many rows (50–100 per month)
**Access**: Authenticated read (service user token); admin-only write
**GDPR**: No personal data

### Fields

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | INTEGER | PRIMARY KEY | Auto-assigned |
| `month_key` | TEXT | NOT NULL | ISO month, e.g. `2026-04` — matches `time_report_config.active_month_key` |
| `training_group` | TEXT | NOT NULL | One of: `simskola`, `tavlingA`, `tavlingB`, `teknik`, `masters`, `vuxencrawl` |
| `date` | TEXT | NOT NULL | ISO date, e.g. `2026-04-15` |
| `title` | TEXT | NOT NULL | Display name, e.g. `Träning`, `Simskola`, `ÖGP Växjö` |
| `hours` | INTEGER | NOT NULL | Duration hours; codes: 10=half-day, 15=overnight, 20=full-day |
| `minutes` | INTEGER | NOT NULL DEFAULT 0 | Duration minutes (0–59) |

### Duration Codes

| `hours` value | Meaning | Salary treatment |
|---------------|---------|-----------------|
| 10 | Half-day competition | Flat `half_day_salary` (not hourly) |
| 15 | Overnight stay | Flat `overnight_salary` (not hourly) |
| 20 | Full-day competition | Flat `full_day_salary` (not hourly) |
| Any other | Normal session | Hourly rate × (hours + minutes/60 + prep time) |

### SQL Migration

```sql
-- trailbase/migrations/U1776427200__create_time_report_sessions.sql
-- GDPR: No personal data — training schedule only

CREATE TABLE IF NOT EXISTS time_report_sessions (
  id             INTEGER PRIMARY KEY,
  month_key      TEXT    NOT NULL CHECK(month_key != ''),
  training_group TEXT    NOT NULL CHECK(training_group IN (
                   'simskola', 'tavlingA', 'tavlingB',
                   'teknik', 'masters', 'vuxencrawl'
                 )),
  date           TEXT    NOT NULL CHECK(date != ''),
  title          TEXT    NOT NULL CHECK(title != ''),
  hours          INTEGER NOT NULL CHECK(hours >= 0),
  minutes        INTEGER NOT NULL DEFAULT 0 CHECK(minutes >= 0 AND minutes < 60)
) STRICT;

CREATE INDEX IF NOT EXISTS idx_sessions_month_group
  ON time_report_sessions (month_key, training_group);
```

### TypeScript Interface

```typescript
export interface TimeReportSession {
  id: number;
  month_key: string;        // e.g. "2026-04"
  training_group: string;   // e.g. "tavlingA"
  date: string;             // e.g. "2026-04-15"
  title: string;            // e.g. "Träning"
  hours: number;            // 10=half-day, 15=overnight, 20=full-day, else normal
  minutes: number;
}

// Grouped by training group (shape used by the Astro page and salary module)
export type SessionsByTrainingGroup = {
  simskola: TimeReportSession[];
  tavlingA: TimeReportSession[];
  tavlingB: TimeReportSession[];
  teknik: TimeReportSession[];
  masters: TimeReportSession[];
  vuxencrawl: TimeReportSession[];
};
```

---

## Entity: Instructor

**Storage**: Trailbase table `instructors` (SQLite STRICT)
**Cardinality**: Many rows (~30 active instructors)
**Access**: Authenticated read (service user); admin-only write
**GDPR**: Personal data — see notes below

### GDPR Documentation

| Field | Personal data? | Legal basis | Retention |
|-------|---------------|-------------|-----------|
| `email` | Yes — contact identifier | Contractual necessity (employment) | Until end of employment + 1 year |
| `swim_school_rate` | Yes — salary data | Contractual necessity (employment) | Until end of employment + 1 year |
| `coach_rate` | Yes — salary data | Contractual necessity (employment) | Until end of employment + 1 year |

Must be registered in `docs/gdpr-register.md` (TODO: GDPR_REGISTER) before go-live.
Deletion path: admin deletes row via Trailbase admin UI.

### Fields

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | INTEGER | PRIMARY KEY | Auto-assigned |
| `email` | TEXT | NOT NULL UNIQUE | Lowercase email address (lookup key) |
| `swim_school_rate` | INTEGER | NOT NULL | SEK per hour for swim school sessions |
| `coach_rate` | INTEGER | DEFAULT NULL | SEK per hour for coaching; NULL if no coaching duties |

### SQL Migration

```sql
-- trailbase/migrations/U1776513600__create_instructors.sql
-- GDPR: Personal data — email and salary rates
-- Legal basis: Contractual necessity (employment relationship)
-- Retention: Until end of employment + 1 year for accounting
-- Access: Authenticated read (service user); admin-only write
-- Deletion path: Admin deletes row via Trailbase admin UI

CREATE TABLE IF NOT EXISTS instructors (
  id               INTEGER PRIMARY KEY,
  email            TEXT    NOT NULL UNIQUE CHECK(email != ''),
  swim_school_rate INTEGER NOT NULL CHECK(swim_school_rate > 0),
  coach_rate       INTEGER CHECK(coach_rate IS NULL OR coach_rate > 0)
) STRICT;
```

### TypeScript Interface

```typescript
export interface Instructor {
  id: number;
  email: string;
  swim_school_rate: number;
  coach_rate: number | null;
}
```

> **Note**: The existing `Instructor` interface (formerly `Employee`, now renamed to `Instructor`) in `src/lib/types.ts` uses
> `swimSchoolRate` and `coachRate` (camelCase). The Trailbase-fetched version
> uses `swim_school_rate` / `coach_rate` (snake_case). The fetch function in
> `src/lib/trailbase.ts` should return the snake_case interface; salary.ts
> will be updated to accept it.

---

## Data Flow Summary

```
Trailbase (fly.io arn) — ALL endpoints require service user token
├── time_report_config [authenticated read — service user JWT]
│     └── fetched by: tidrapport.astro SSR frontmatter (page load)
├── time_report_sessions [authenticated read — service user JWT]
│     └── fetched by: tidrapport.astro SSR frontmatter (page load)
└── instructors [authenticated read — service user JWT]
      └── fetched by: send-time-report.ts at form submission time

Cloudflare Worker secrets
├── TRAILBASE_URL               (existing)
├── TRAILBASE_SERVICE_EMAIL     (new)
└── TRAILBASE_SERVICE_PASSWORD  (new)

Auth flow per Worker invocation:
  POST /api/auth/v1/token → auth_token (JWT)
  Reuse auth_token for all Trailbase calls within the same invocation
```

---

## Removed Artifacts (post-migration)

| File | Replaced by |
|------|-------------|
| `src/config/time-report-items.json` | `time_report_sessions` table |
| `src/config/time-report-settings.ts` | `time_report_config` table + `import.meta.env.DEV` |
| Hardcoded `EMPLOYEES` array in `send-time-report.ts` | `instructors` table |
