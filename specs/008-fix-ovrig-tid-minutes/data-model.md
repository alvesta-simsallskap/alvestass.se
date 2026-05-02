# Data Model: Fix Övrig Tid Minutes Bug

**Date**: 2026-04-30

## Summary

No new data entities. No Trailbase schema changes. No migrations.

## Affected Type

The existing `ExtraTimeRow` type in `src/lib/types.ts` is unchanged. The `h` and `m` fields remain `string` (as they come from FormData). Empty strings are coerced to `"0"` in the parser before the row is stored.

```typescript
// Existing — no change needed
interface ExtraTimeRow {
  date: string;
  h: string;   // hours as string; empty string treated as "0"
  m: string;   // minutes as string; empty string treated as "0"
  desc: string;
}
```

The Alpine.js model for `extraTimes` rows will have `h: 0` and `m: 0` as defaults (numeric `0`, rendered as `"0"` by Alpine's `x-model` binding on a number input). This is compatible with the existing FormData parsing since `formData.get()` always returns a string.
