# Research: Fix Övrig Tid Minutes Bug

**Date**: 2026-04-30

## Summary

No external research required. The bug is fully understood from reading two source files.

## Root Cause

**File**: `src/lib/timeReportValidation.ts`, line 14

```typescript
if (date && h && m && desc) {
```

An empty string (`""`) is falsy in JavaScript/TypeScript. When the user leaves the minutes field blank, `m === ""`, so the entire condition evaluates to false and the row is not pushed to `extraRows`. The same issue affects `h` (hours left blank).

## Fix Decision

- **Decision**: Change the validity check to require only `date` and `desc` (the identifying fields). Coerce empty `h` and `m` to `"0"` before building the row.
- **Rationale**: A row with a date, a description, and zero time is a valid (if unusual) submission. Hours and minutes are purely numeric quantities with a natural default of 0.
- **Alternative considered**: Make the minutes input field `required` in HTML — rejected because it would break existing submissions mid-session and is an overcorrection (the field *should* be optional).

## UI Default Decision

- **Decision**: Change the Alpine.js initial model for new `extraTimes` rows from `{ h: '', m: '' }` to `{ h: 0, m: 0 }`.
- **Rationale**: Defaulting to `0` makes the intent clear to the user (field means "number of hours/minutes") and prevents the empty-string issue from arising at all, regardless of whether the server-side fix is in place.
- **Rationale for treating this as a complementary change**: The server-side fix alone is sufficient for correctness, but the UI default is a small, low-risk UX improvement that aligns with the behaviour of a numeric input field.
- **Alternative considered**: Leave UI unchanged and rely solely on server-side fix — technically correct but leaves the UX unchanged.
