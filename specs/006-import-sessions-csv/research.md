# Research: CSV Import of Time Report Sessions

## CSV Parsing in Go

**Decision**: Use `encoding/csv` from the Go standard library.

**Rationale**: The stdlib reader handles RFC 4180 CSV correctly (quoted fields with embedded commas, CRLF/LF line endings). No third-party dependency is needed. The `csv.Reader` returns `[]string` rows; column order is resolved by reading the header row and building a name→index map, satisfying FR-013 (any column order accepted).

**Alternatives considered**: None — the stdlib package covers all requirements and is already available.

---

## Upsert Strategy via Trailbase SDK

**Decision**: Fetch all existing `time_report_sessions` records for each distinct `month_key` present in the CSV using `RecordApi.List` with a `FilterColumn`; build an in-memory map keyed by `(month_key, training_group, date, title)` → `{id, hours, minutes}`; then classify each CSV row as insert, update, or skip.

**Rationale**: Trailbase has no native UPSERT endpoint. Fetching all records for relevant months upfront allows O(1) lookup per CSV row and avoids one round-trip per row. The `time_report_sessions` table has an index on `(month_key, training_group)`, so filtered list queries are efficient. A CSV for a single month will typically have tens of rows — the entire dataset fits comfortably in memory.

**Alternatives considered**:
- **One query per CSV row**: Would work but creates N round-trips for N rows; unnecessary given the small data volume.
- **No pre-fetch, attempt create and handle duplicate error**: Trailbase returns an HTTP error on constraint violation but does not expose a structured error body reliable enough to distinguish "already exists" from other errors.

---

## Pagination When Fetching Sessions

**Decision**: Fetch all pages using cursor-based pagination. Issue `List` calls with `cursor` from the previous response until `cursor` is `nil`. Use a page size of 1000 (pass `Limit: ptr(uint64(1000))`).

**Rationale**: The `ListResponse` returned by the SDK includes an optional `cursor` field for the next page. The `time_report_sessions` table is not expected to have more than a few hundred rows per month, so one page is almost always sufficient; cursor following is a safety net.

**Alternatives considered**: Fetching without a limit and relying on Trailbase's default page size — risky because the default may be smaller than the dataset.

---

## TUI Flow for File Path Input

**Decision**: Add a new menu item "Importera tidrapportpass" that launches an interactive Bubbletea model. The model first prompts for the CSV file path using the existing `textinput` component (same pattern as `ui/update.go`), then proceeds through validation → summary → confirmation → execution phases.

**Rationale**: Consistent with the existing menu-driven UX pattern. The `textinput` component already supports path entry. No shell argument parsing is added because the spec says the import is accessible from the main menu (FR-012), and adding a subcommand would change the CLI's interaction model without justification.

**Alternatives considered**: Accepting the file path as a CLI flag/argument — would require restructuring `main.go` to support subcommands and deviate from the current menu-driven design.

---

## Session Classification Logic

**Decision**: A CSV row is classified as:
- **skip** — composite key exists, hours and minutes are identical to the stored record.
- **update** — composite key exists, hours or minutes differ.
- **insert** — composite key not found in existing data.

All validation happens before the confirmation prompt. If any row is invalid, the import is aborted (FR-009). The confirmation prompt shows counts for inserts, updates, and skips. After confirmation, inserts are applied first, then updates.

**Rationale**: Skipping truly unchanged rows reduces unnecessary write operations and makes the pre-confirmation summary more useful to the administrator.

**Alternatives considered**: Updating even unchanged rows — simpler code but wastes round-trips and makes the summary less informative.

---

## Code Structure

**Decision**: New `internal/importer/` package containing `csv.go` (parsing and per-row validation) and `csv_test.go` (unit tests). Session-specific Trailbase operations go in `internal/trailbase/sessions.go`. The TUI model goes in `internal/ui/import.go`.

**Rationale**: Keeps the importer logic independently testable (FR per constitution Principle II). The `trailbase` package is the right home for typed API wrappers, consistent with `client.go`. The `ui` package owns all Bubbletea models.

**Alternatives considered**: Putting everything in one file — would make unit-testing the parsing logic impractical without a running Trailbase instance.
