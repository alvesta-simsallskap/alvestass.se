# Quickstart: CSV Import of Time Report Sessions

## Prerequisites

- Admin CLI built and configured (`alvestass-admin` in PATH, or run via `go run`)
- A valid session (first-run setup completed)
- A CSV file following the format documented in `contracts/csv-format.md`

## Running the Import

1. Launch the CLI:
   ```bash
   ./alvestass-admin
   # or during development:
   cd tools/admin-cli && go run ./cmd/alvestass-admin
   ```

2. From the main menu, select **"Importera tidrapportpass"** (use arrow keys and Enter, or press the assigned number key).

3. Enter the path to the CSV file when prompted (absolute or relative path accepted).

4. The CLI fetches existing sessions, validates all rows, and displays a summary:
   ```
   Förhandsgranskning:
     Infogningar:  12
     Uppdateringar: 3
     Oförändrade:   5
   
   Vill du genomföra importen? (j/n)
   ```

5. Press **j** to confirm or **n** / **Esc** to cancel.

6. On success the CLI reports:
   ```
   ✓ Import klar: 12 tillagda, 3 uppdaterade, 5 oförändrade.
   ```

## Preparing a CSV File

Minimum required columns: `month_key`, `training_group`, `date`, `title`, `hours`.

```csv
month_key,training_group,date,title,hours,minutes
2026-04,simskola,2026-04-02,Nybörjare grupp 1,1,30
2026-04,tavlingA,2026-04-02,Morgonträning,2,0
```

See `contracts/csv-format.md` for the full specification.

## Allowed Training Groups

`simskola`, `tavlingA`, `tavlingB`, `teknik`, `masters`, `vuxencrawl`

## Handling Validation Errors

If the file contains errors the CLI lists each one with its line number and exits without making changes:

```
Fel i CSV-filen:
  Rad 3: träningsgrupp "ungdomA" är inte giltig
  Rad 5: timmar måste vara ett heltal ≥ 0
Ingen data importerades.
```

Fix the errors and re-run.

## Development

Build and test:
```bash
cd tools/admin-cli
go test ./internal/importer/...   # unit tests for CSV parsing
go build ./cmd/alvestass-admin    # verify compilation
```
