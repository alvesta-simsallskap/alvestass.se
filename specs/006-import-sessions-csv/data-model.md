# Data Model: CSV Import of Time Report Sessions

## Existing Entity: TimeReportSession

Already defined in `trailbase/migrations/U1776427200__create_time_report_sessions.sql`. No schema migration required.

| Field           | Type    | Constraints                                                                 |
|-----------------|---------|-----------------------------------------------------------------------------|
| `id`            | INTEGER | Primary key (auto-assigned by Trailbase)                                    |
| `month_key`     | TEXT    | NOT NULL, non-empty (e.g. `"2026-04"`)                                      |
| `training_group`| TEXT    | NOT NULL, one of: `simskola`, `tavlingA`, `tavlingB`, `teknik`, `masters`, `vuxencrawl` |
| `date`          | TEXT    | NOT NULL, non-empty (format `YYYY-MM-DD` by convention)                     |
| `title`         | TEXT    | NOT NULL, non-empty                                                         |
| `hours`         | INTEGER | NOT NULL, ≥ 0                                                               |
| `minutes`       | INTEGER | NOT NULL, ≥ 0 AND < 60, default 0                                           |

**Composite key** (used for upsert identity): `(month_key, training_group, date, title)`

---

## New Go Type: `SessionRow` (in `internal/importer/csv.go`)

Represents one parsed and validated row from the CSV file, before any Trailbase interaction.

```
SessionRow {
  MonthKey      string   // required, non-empty
  TrainingGroup string   // required, one of the 6 allowed values
  Date          string   // required, non-empty
  Title         string   // required, non-empty
  Hours         int      // required, ≥ 0
  Minutes       int      // required, ≥ 0 and < 60; defaults to 0 if column absent
}
```

---

## New Go Type: `ParseError` (in `internal/importer/csv.go`)

Represents a validation error for one CSV row.

```
ParseError {
  Line    int    // 1-based line number in the CSV file (including header)
  Column  string // column name where the error occurred (empty for row-level errors)
  Value   string // the offending value
  Message string // Swedish human-readable description
}
```

---

## New Go Type: `ExistingSession` (in `internal/importer/csv.go`)

Represents a session record already stored in Trailbase, used as input to `ComputeDiff`.
Defined in the `importer` package so `ComputeDiff` has no dependency on the `trailbase` package,
avoiding a circular import.

```
ExistingSession {
  ID            int64  // Trailbase record ID (needed for updates)
  MonthKey      string
  TrainingGroup string
  Date          string
  Title         string
  Hours         int
  Minutes       int
}
```

---

## New Go Type: `ImportDiff` (in `internal/importer/csv.go`)

The result of comparing CSV rows against existing Trailbase records. Computed before the confirmation prompt.

```
ImportDiff {
  Inserts []SessionRow      // rows with no matching existing record
  Updates []SessionUpdate   // rows where the key matches but values differ
  Skipped []SessionRow      // rows where the key matches and values are identical
}
```

---

## New Go Type: `SessionUpdate` (in `internal/importer/csv.go`)

Pairs a CSV row (new values) with the Trailbase record ID it will overwrite.

```
SessionUpdate {
  ID  int64      // Trailbase record ID of the existing session
  Row SessionRow // new values from the CSV
}
```

---

## New Go Type: `ImportResult` (in `internal/importer/csv.go`)

The outcome after applying a confirmed import.

```
ImportResult {
  Inserted int
  Updated  int
  Skipped  int
}
```

---

## Internal Go Type: `TimeReportSession` (in `internal/trailbase/sessions.go`)

Go struct mirroring the `time_report_sessions` table for use with the Trailbase SDK.
Internal to the `trailbase` package — callers receive `[]importer.ExistingSession` from
`ListAllSessions`, not this type directly.

```
TimeReportSession {
  ID            int    `json:"id"`
  MonthKey      string `json:"month_key"`
  TrainingGroup string `json:"training_group"`
  Date          string `json:"date"`
  Title         string `json:"title"`
  Hours         int    `json:"hours"`
  Minutes       int    `json:"minutes"`
}
```

---

## Validation Rules (applied in `internal/importer/csv.go`)

| Rule | Field | Constraint | Swedish error message |
|------|-------|------------|-----------------------|
| R-01 | `month_key` | non-empty string | "month_key får inte vara tomt" |
| R-02 | `training_group` | one of the 6 allowed values | "träningsgrupp '{value}' är inte giltig" |
| R-03 | `date` | non-empty string | "datum får inte vara tomt" |
| R-04 | `title` | non-empty string | "titel får inte vara tomt" |
| R-05 | `hours` | parseable integer ≥ 0 | "timmar måste vara ett heltal ≥ 0" |
| R-06 | `minutes` | parseable integer ≥ 0 and < 60 | "minuter måste vara ett heltal mellan 0 och 59" |
| R-07 | header | all required columns present | "obligatorisk kolumn saknas: '{column}'" |

Required CSV columns: `month_key`, `training_group`, `date`, `title`, `hours`. Optional: `minutes` (defaults to 0 if absent).

**Note on R-07**: R-07 is a **header-level** check performed once in `ParseCSV` when the header row is read. It produces a single `ParseError` with `Line: 1` and causes immediate abort before any rows are processed. Rules R-01 through R-06 are **per-row** checks performed inside `validateRow`.
