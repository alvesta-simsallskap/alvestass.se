# Deferred plan: fixed salary ("fast lön") with a time bank

> **Status: deferred — not yet implemented.** This was designed alongside the per-instructor
> monetary addon feature (now implemented) but intentionally postponed. The decisions below were
> confirmed with the user. Implement separately when prioritised.

## Goal

Some instructors are paid a fixed monthly salary, so an hourly estimate is meaningless. Instead,
the hours they report adjust a personal **time bank** (a stored balance, adjusted manually in the
DB at review time). The time-report preview and the emailed report should show *"Fast lön"* plus
the estimated time-bank change and the resulting balance — and must **not** compute a salary
figure for these instructors.

## Confirmed decisions

- **Time-bank change** = total reported time this month (all training groups + övrig tid,
  including automatic prep time), shown as a single positive figure. The reviewer applies the
  change manually in the DB. There is no "expected monthly hours" field.
- **Display** = current stored balance + estimated change + resulting new balance.
- The **monetary addon is hidden** for fixed-salary instructors (they show only "Fast lön" + the
  time-bank change, no kr lines).

## Schema

Add to the `instructors` table via a new migration (table-rename recreation pattern, as in
`trailbase/migrations/U1776686400__update_instructors.sql`; never edit an applied migration):

- `fixed_salary INTEGER NOT NULL DEFAULT 0` — boolean (0/1)
- `time_bank INTEGER NOT NULL DEFAULT 0` — signed balance in **minutes**

Keep the existing `CHECK(swim_school_rate IS NOT NULL OR coach_rate IS NOT NULL)` — the rates
still gate which form sections appear in the wizard, even for fixed-salary instructors. Include
the standard GDPR header comment block (employment terms = personal data).

## Code changes

- **`src/lib/types.ts`**: add `fixed_salary: boolean;` and `time_bank: number;` to `Instructor`.
- **`src/lib/salary.ts`**: add two pure helpers —
  - `calcWorkedMinutes(data, schedule, config, instructor)` → total reported minutes: sum of
    `hours*60 + minutes` from `calcSalary()` over the six groups + `extratid` (reuses the existing
    prep-time logic and the 10/15/20 flat-rate-sentinel exclusion).
  - `formatMinutes(min)` → `"h:mm"` string, handling negative balances (e.g. `"-2:30"`).
- **`src/lib/timeReportHtml.ts`**: in the `if (instructor)` block, when `instructor.fixed_salary`
  is true, skip the salary table entirely and render a "Tidbank" section instead — `Lön: Fast lön`
  plus a table with rows:
  - *Nuvarande saldo* — `formatMinutes(time_bank)`
  - *Rapporterad tid denna månad* — `+formatMinutes(worked)`
  - *Nytt saldo (uppskattat)* — `formatMinutes(time_bank + worked)`
  - plus a note that the time bank is adjusted manually at review.
  No addon row for fixed-salary instructors.
- No changes to the preview/send endpoints — both already fetch the full `Instructor` and pass it
  to `buildTimeReportHtml`.

## Tests

- Extend the `Instructor` fixtures in `tests/salary.test.ts` and `tests/trailbase.test.ts` with the
  two new fields.
- Add `calcWorkedMinutes` cases (sums groups + övrig tid) and `formatMinutes` cases (positive,
  zero, negative, minutes < 10 zero-padding).
- Render test: `buildTimeReportHtml` emits "Fast lön" + the time-bank rows for a fixed-salary
  instructor and no kr salary table.

## Verification

1. `pnpm test` and `pnpm build` (= `astro check && astro build`) pass.
2. `cd trailbase && fly deploy`; confirm the new columns in the Trailbase admin UI.
3. Set `fixed_salary = 1`, `time_bank = 90` (1:30) on a test instructor; run `pnpm dev`, go through
   `/tidrapport`, report some sessions, open the preview — verify it shows "Fast lön", saldo 1:30,
   the reported time as the change, the new estimated balance, and **no** kr salary table.
