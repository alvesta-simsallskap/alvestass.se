# CLI Command Schema

**Feature**: 003-admin-cli  
**Date**: 2026-04-22

---

## Overview

The Admin CLI presents a menu-driven interactive interface. It also supports a minimal set of non-interactive flags for scripting. All user-visible strings are in Swedish.

---

## Invocation

```
alvestass-admin [flags]
```

| Flag | Description |
|------|-------------|
| `--help`, `-h` | Print usage and exit |
| `--version` | Print version string (`003-admin-cli vX.Y.Z`) and exit |
| `--config <path>` | Override config file path (default: `$UserConfigDir/alvestass-admin/config.json`) |

No subcommand arguments — all operations are selected via the interactive menu.

---

## Interactive Menu Structure

```
╔══════════════════════════════════╗
║  Alvesta Simsällskap — Admin CLI ║
╚══════════════════════════════════╝

Välj en åtgärd:

  [1] Uppdatera kontaktuppgifter
  [2] Kontrollera data
  [3] Hjälp
  [4] Avsluta

>
```

---

## Operation Contracts

### [1] Uppdatera kontaktuppgifter (Update)

**Input**: Field selection and new values entered interactively.

**Behaviour**:
1. Fetch current `club_info` record via `GET /api/records/v1/club_info/1`; display current values
2. Present numbered field selector for all editable fields
3. For each selected field: prompt for new value, show current value as default
4. Validate all changed fields before showing confirm prompt (FR-13)
5. Show diff of all changes; prompt `Spara ändringar? (j/n)`
6. On confirm: `PATCH /api/records/v1/club_info/1` with only changed fields

**Exit conditions**:
- `n` at any confirm prompt → return to Main Menu without changes
- `Ctrl+C` at any point → return to Main Menu without changes (FR-07)
- Validation error → display Swedish error, allow retry or cancel
- Network error → display Swedish error message, return to Main Menu

**Output**:
```
Kontaktuppgifter uppdaterade.
```
or
```
Uppdatering avbruten.
```

---

### [2] Kontrollera data (Consistency Check)

**Input**: None.

**Behaviour**:
1. Fetch `club_info` record via `GET /api/records/v1/club_info/1`
2. Run all validation rules
3. Report each violation as: `[FÄLT] [värde] — [regel som bröts]`
4. If no violations: `Inga problem hittades.`

**Output** (example):
```
Kontroll klar — 2 problem hittades:

  [email] "" — E-postadress får inte vara tom
  [postal_code] "3423" — Postnummer måste ha formatet NNN NN

Gå tillbaka till menyn? (Enter)
```

---

### [3] Hjälp (Help)

Displays a description of each operation and step-by-step instructions. No network calls.

**Output** (abbreviated):
```
╔══════════ Hjälp ══════════════════╗

[1] Uppdatera kontaktuppgifter
    Redigera enskilda fält direkt i terminalen.
    Aktuella värden visas som standardsvar.
    ...

[2] Kontrollera data
    Hämtar det aktuella posten och kontrollerar alla fält
    mot valideringsreglerna. Visar eventuella problem.
    ...

Tryck Enter för att återgå till menyn.
```

---

### [4] Avsluta (Exit)

Prints `Hej då!` and exits with code `0`.

---

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Normal exit (user chose Avsluta, or `--version`/`--help`) |
| `1` | Fatal error (backend unreachable at startup, config write failure) |

---

## Config First-Run Wizard

Triggered when no config file exists or the existing config is missing required fields.

```
Välkommen till Alvesta Simsällskap Admin CLI!

Ange backend-URL (t.ex. https://alvestass-trailbase.fly.dev):
> _

Ange din e-postadress:
> _

Ange ditt lösenord:
> ****

Ansluter...  ✓

Konfiguration sparad. Du kan nu använda CLI:t.
```

Credentials (email/password) are used once to obtain a JWT; they are not stored.
