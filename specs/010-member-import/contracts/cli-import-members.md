# Contract: CLI — Import Members

**Branch**: `010-member-import` | **Date**: 2026-05-19  
**Command**: `alvestass-admin` → menu item `Importera memberregister`

## Menu Integration

The existing TUI menu gains a new item between "Importera tidrapportsessioner" and "Hjälp":

```
Importera memberregister   (MenuImportMembers)
```

## User Flow

```
1. User selects "Importera memberregister" from main menu
2. CLI prompts for path to IdrottOnline export CSV
3. CLI prompts for path to WeUnite Grupplista CSV
4. CLI parses both files; reports parse errors (if any) and aborts
5. CLI displays preview:
     Hittade X aktiva simmare, Y ledare, Z styrelseledamöter
     Hittade W målsmän, V träningsgrupper, U familjekonstellationer
     XX poster saknar IID och hoppas över
6. CLI asks: "Importera till Trailbase? (j/n)"
7. On confirm: runs upsert import; shows progress
8. Displays result summary:
     Importerade: X members, Y guardians, Z groups, W family links
     Hoppade över: N poster (visas lista med skäl)
9. Returns to main menu
```

## Import Result Type

```go
type MemberImportResult struct {
    MembersImported    int
    GuardiansImported  int
    GroupsImported     int
    FamilyLinksImported int
    Skipped            []SkippedRecord
}

type SkippedRecord struct {
    SourceFile string // "idrottonline" | "weunite"
    Line       int
    Reason     string // human-readable Swedish description
}
```

## Error Handling

| Condition | Behaviour |
|-----------|-----------|
| File not found or unreadable | Error message; return to menu |
| CSV parse error (malformed row) | List all errors; abort before any Trailbase write |
| Record missing IID | Skip record; add to `Skipped` list |
| Trailbase write failure | Abort; show error; Trailbase state may be partial |
| User declines confirmation | Return to menu without writing anything |

## Deletion Contract

A future `delete-member` sub-flow in the admin CLI must support:

```
Input:  IID string (e.g. "IID12345678")
Action: DELETE FROM members WHERE iid = ?
Effect: CASCADE deletes guardians, member_training_groups, family_members rows
Output: Confirmation message with counts of cascaded deletions
```

This deletion path satisfies the GDPR data-subject erasure requirement (Art. 17) and must be implemented before the member register feature is considered production-ready.
