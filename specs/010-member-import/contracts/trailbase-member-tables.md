# Contract: Trailbase Member Table Access Patterns

**Branch**: `010-member-import` | **Date**: 2026-05-19

These tables are created by migration `U1779235200__create_member_register.sql`. This document defines the intended read/write access patterns for future features that consume the member register.

## Access Control Principle

**All six tables MUST be configured as authenticated-only in Trailbase's access control settings.** No table may be exposed via a public (unauthenticated) Trailbase API route. This is a hard GDPR requirement (Principle VII).

Until a Trailbase authentication feature is implemented, the tables should have **no API access rules** (accessible only via the Trailbase admin UI and the admin CLI).

### Applying Access Control (T003b — required before production import)

After applying the migration, configure access control in the Trailbase admin UI:

1. For each of the six tables (`members`, `guardians`, `training_groups`, `member_training_groups`, `families`, `family_members`): open **Table Settings → Access Rules** and ensure no public (unauthenticated) read or write rules are defined.
2. The admin CLI uses a service-user token (stored in `config.json`) which bypasses table-level access rules.
3. Record the completion of this step with a ✅ in the GDPR gate table in `specs/010-member-import/plan.md` before running the production import.

## Table Access Patterns

### `members`

| Operation | Caller | Notes |
|-----------|--------|-------|
| Read all | Admin CLI | Authenticated service token |
| Read by IID | Future login/auth feature | Look up by `iid` after auth |
| Read by group | Future attendance feature | Join with `member_training_groups` |
| Write (upsert) | Admin CLI import | Import command only |
| Delete | Admin CLI delete-member | Cascades to child tables |

### `guardians`

| Operation | Caller | Notes |
|-----------|--------|-------|
| Read by member_iid | Future parent communication feature | Authenticated |
| Write | Admin CLI import | Import command only |
| Delete | CASCADE on member delete | Or direct admin UI row delete |

### `training_groups`

| Operation | Caller | Notes |
|-----------|--------|-------|
| Read all | Future scheduling, attendance, and public group listing features | Unauthenticated read is acceptable for group names/categories (no personal data) |
| Write | Admin CLI import | Import command only; future admin UI |

### `member_training_groups`

| Operation | Caller | Notes |
|-----------|--------|-------|
| Read by group_id | Future attendance feature | Authenticated — reveals who is in a group |
| Read by member_iid | Future member profile feature | Authenticated |
| Write | Admin CLI import | Import command only |

### `families` and `family_members`

| Operation | Caller | Notes |
|-----------|--------|-------|
| Read | Future family discount / sibling lookup feature | Authenticated |
| Write | Admin CLI import | Import command only |

## Go SDK Usage Pattern

Future Go code accessing these tables should follow the pattern established by `internal/trailbase/sessions.go`:

```go
// Example: list members in a training group
type Member struct {
    IID         string `json:"iid"`
    FirstName   string `json:"first_name"`
    LastName    string `json:"last_name"`
    // ... other fields
}

api := tb.NewRecordApi[Member](client.sdk, "members")
resp, err := api.List(&tb.ListArguments{
    Filters: []tb.Filter{
        // join via member_training_groups is not directly supported by SDK;
        // use a Trailbase view or fetch group members separately
    },
})
```

## Trailbase View (Recommended for Attendance Feature)

To avoid N+1 queries when listing group members with names, a Trailbase SQL view is recommended for future features:

```sql
CREATE VIEW IF NOT EXISTS group_members_view AS
SELECT
    mtg.group_id,
    tg.name        AS group_name,
    tg.category    AS group_category,
    m.iid,
    m.first_name,
    m.last_name,
    mtg.role
FROM member_training_groups mtg
JOIN members m ON m.iid = mtg.member_iid
JOIN training_groups tg ON tg.id = mtg.group_id;
```

This view can be added in a future migration when the attendance feature is planned.
