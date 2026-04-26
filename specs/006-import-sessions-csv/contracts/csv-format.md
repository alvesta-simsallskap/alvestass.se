# Contract: CSV File Format

## Overview

The import command accepts a UTF-8 encoded CSV file conforming to RFC 4180.

## Header Row

The file MUST begin with a header row. Column names are case-sensitive.

**Required columns** (in any order):

| Column name      | Type    | Description |
|------------------|---------|-------------|
| `month_key`      | string  | Month identifier, e.g. `2026-04` |
| `training_group` | string  | One of the 6 allowed values (see below) |
| `date`           | string  | Session date, convention `YYYY-MM-DD` |
| `title`          | string  | Session title / description |
| `hours`          | integer | Duration hours, ≥ 0 |

**Optional columns**:

| Column name | Type    | Default | Description |
|-------------|---------|---------|-------------|
| `minutes`   | integer | `0`     | Duration minutes, 0–59 |

Additional columns beyond those listed are silently ignored.

## Allowed Values for `training_group`

| Value        | Meaning |
|--------------|---------|
| `simskola`   | Simskola |
| `tavlingA`   | Tävling A |
| `tavlingB`   | Tävling B |
| `teknik`     | Teknik |
| `masters`    | Masters |
| `vuxencrawl` | Vuxencrawl |

## Example

```csv
month_key,training_group,date,title,hours,minutes
2026-04,simskola,2026-04-02,Nybörjare grupp 1,1,30
2026-04,tavlingA,2026-04-02,Morgonträning,2,0
2026-04,teknik,2026-04-03,Teknikpass,1,45
```

## Encoding and Delimiters

- Encoding: UTF-8 (with or without BOM)
- Delimiter: comma (`,`)
- Line endings: LF (`\n`) or CRLF (`\r\n`)
- Quoted fields: double-quote (`"`) per RFC 4180; embedded commas and newlines within quoted fields are supported

## Composite Key

The combination of `month_key + training_group + date + title` uniquely identifies a session for upsert purposes. If the same composite key appears more than once in the file, the last occurrence wins.

## Error Behaviour

If any row fails validation, the CLI reports all errors with their 1-based line number (counting the header as line 1) and exits without importing any rows.
