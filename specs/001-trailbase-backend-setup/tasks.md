# Tasks: Trailbase Backend Setup (Minimal Starter)

**Input**: Design documents from `specs/001-trailbase-backend-setup/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/club-info-api.md, quickstart.md

**Tests**: No automated tests requested. Testing is manual browser verification per the spec and constitution (Principle II).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Trailbase Backend Files)

**Purpose**: Create the Trailbase backend service configuration files in the repository.

- [x] T001 Create `trailbase/Dockerfile` — use `FROM ghcr.io/trailbase-core/trail:latest` per research.md §4
- [x] T002 [P] Create `trailbase/fly.toml` — app `alvestass-trailbase`, `primary_region = "arn"`, mount volume `trailbase_data` at `/app/data`, internal port `4000`, HTTPS service per plan.md fly.toml section
- [x] T003 [P] Create `trailbase/migrations/0001_initial.sql` — `club_info` table schema with CHECK constraints and seed row per data-model.md SQL Migration section

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Prepare the Astro codebase for SSR and Trailbase integration AND deploy the backend. Two parallel tracks: **Astro code** (T004–T006) and **Trailbase deployment** (T007–T010). Both MUST complete before user story work begins.

### Track A: Astro Codebase Preparation

- [x] T004 Update `astro.config.mjs` — Astro 5 removed `output: 'hybrid'`; the default `output: 'static'` already supports per-page SSR via `prerender = false`. No config change needed.
- [x] T005 [P] Update `src/env.d.ts` — add `TRAILBASE_URL: string` to the `env` interface per quickstart.md env.d.ts section
- [x] T006 [P] Create `src/lib/trailbase.ts` — export `ClubInfo` interface (per data-model.md TypeScript section) and `fetchClubInfo(baseUrl: string): Promise<ClubInfo | null>` function that GETs from the Trailbase Records API (per contracts/club-info-api.md), returns the first record or `null` if empty, and throws on network errors

### Track B: Trailbase Deployment (Manual / Operational)

- [x] T007 Deploy Trailbase to fly.io — from `trailbase/` run `fly launch --no-deploy --region arn --name alvestass-trailbase`, `fly volumes create trailbase_data --region arn --size 1`, `fly deploy` per quickstart.md step 1
- [x] T008 Retrieve admin credentials — run `fly logs --app alvestass-trailbase`, find the "Created admin user" line with generated password, then log in at `https://alvestass-trailbase.fly.dev/_/admin/` and verify `club_info` table and seed row exist per quickstart.md step 2
- [x] T009 Configure `club_info` access rules — set read access to `Everyone` (public, unauthenticated) and write to `Admin only` per quickstart.md step 3 and research.md §1
- [x] T010 Verify REST API — run `curl "https://alvestass-trailbase.fly.dev/api/records/v1/club_info?limit=1"`, confirm JSON response with a `records` array per quickstart.md step 4

**Checkpoint**: Trailbase running in `arn`, API reachable and returning JSON, Astro configured for hybrid SSR, typed client ready.

---

## Phase 3: User Story 1 — Public Visitor Sees Live Club Info (Priority: P1) 🎯 MVP

**Goal**: A visitor to alvestass.se sees up-to-date club contact info on `/kontakt`, served from Trailbase through Cloudflare edge cache.

**Independent Test**: Open `http://localhost:4321/kontakt` during `pnpm dev` and confirm club name, address, phone, email, and short description appear, all sourced from Trailbase.

### Implementation for User Story 1

- [x] T011 [US1] Create `src/pages/kontakt.astro` — SSR page (`export const prerender = false`) that imports `fetchClubInfo` from `src/lib/trailbase.ts`, calls it with `Astro.locals.runtime.env.TRAILBASE_URL` in the frontmatter, and renders a Bulma card displaying: name, tagline, founding year, short description, address + city + postal code, phone (as `tel:` link), email (as `mailto:` link). All visible text in Swedish. Mobile-first layout per Principle III.
- [x] T012 [US1] Set cache headers in `kontakt.astro` — `Astro.response.headers.set('Cache-Control', 'public, max-age=300, stale-if-error=86400')` per research.md §2. This gives ≤5-min freshness (SC-002) and 24h fallback when backend is unreachable.
- [x] T013 [US1] Handle edge cases in `kontakt.astro` — wrap `fetchClubInfo()` in try/catch; if Trailbase is unreachable (catch block), `clubInfo` remains `null`. If `clubInfo` is `null`, render a Swedish placeholder: "Kontaktuppgifter laddas inte just nu. Försök igen senare." per spec Edge Cases section.
- [x] T014 [US1] Run `pnpm build` — verify zero TypeScript errors, zero `astro check` errors, successful build output in `dist/`
- [x] T015 [US1] Manual browser test — run `pnpm dev`, open `/kontakt` in a browser (mobile + desktop widths), verify all ClubInfo fields render correctly and match the seed data in Trailbase admin UI (SC-001)

**Checkpoint**: User Story 1 is fully functional and independently testable. Visitor can see live backend data on `/kontakt`.

---

## Phase 4: User Story 2 — Admin Updates Club Info Without a Deployment (Priority: P2)

**Goal**: A developer/admin edits any club info field in the Trailbase admin UI and the change appears on `/kontakt` within 5 minutes — no code deployment required.

**Independent Test**: Change a field in the Trailbase admin, wait ≤5 min, hard-refresh `/kontakt`, confirm the new value.

### Implementation for User Story 2

- [x] T016 [US2] Fill in remaining seed data — in Trailbase admin UI, populate the `address`, `tagline`, and `short_description` fields for the club_info record with real Alvesta Simsällskap content
- [x] T017 [US2] Set `TRAILBASE_URL` Cloudflare Worker secret — run `wrangler secret put TRAILBASE_URL` (enter `https://alvestass-trailbase.fly.dev`) per quickstart.md step 4
- [x] T018 [US2] Manual test: update flow — change the tagline in the Trailbase admin UI, wait up to 5 minutes for Cloudflare edge cache to expire, hard-refresh `/kontakt`, confirm the new tagline appears (SC-002)
- [x] T019 [US2] Manual test: validation rejection — attempt to save an empty `name` field in the Trailbase admin UI, confirm the save is rejected with a validation error and `/kontakt` continues to show the previous valid data (FR-005, spec Acceptance Scenario 2)

**Checkpoint**: Both user stories independently verified. Admin can update content without a deployment.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Final verification that no existing functionality has regressed and developer onboarding is smooth.

- [x] T020 [P] Create `.dev.vars.example` at repository root — include `TRAILBASE_URL=https://alvestass-trailbase.fly.dev` as a template for local development setup
- [x] T021 Run `pnpm build` final gate — zero errors, successful build (Principle II)
- [x] T022 Manual regression check — run `pnpm dev` and verify existing pages (`/`, `/tidrapport`, `/tack`) render correctly, are still statically generated, and show no regressions from the `output: 'hybrid'` change
- [x] T023 Manual test: fallback behaviour — stop fly.io machine (`fly scale count 0 --app alvestass-trailbase`), visit `/kontakt` on the deployed site, verify it serves last cached version (not error or blank), then restore (`fly scale count 1 --app alvestass-trailbase`) per quickstart.md step 5

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS all user stories
  - Track A (T004–T006) and Track B (T007–T010) can proceed in parallel
- **User Story 1 (Phase 3)**: Depends on Foundational phase completion (both tracks)
- **User Story 2 (Phase 4)**: Depends on User Story 1 completion (T015 confirms `/kontakt` works)
- **Polish (Phase 5)**: Depends on both user stories complete

### Task Dependencies Within Phases

- **Phase 1**: T001 creates the Dockerfile; T002 and T003 are parallel (different files)
- **Phase 2 Track A**: T004 is independent; T005 and T006 are parallel (different files)
- **Phase 2 Track B**: T007 → T008 → T009 → T010 (sequential deployment chain)
- **Phase 3**: T011 → T012 → T013 → T014 → T015 (build on the same page file)
- **Phase 4**: T016 → T017 → T018 → T019 (sequential operational + test chain)
- **Phase 5**: T020 is parallel with T021; T022 and T023 depend on T021

### Parallel Opportunities

```text
# Phase 1 — all three files in parallel:
T001: Create trailbase/Dockerfile
T002: Create trailbase/fly.toml
T003: Create trailbase/migrations/0001_initial.sql

# Phase 2 — two independent tracks:
Track A: T004, T005 (parallel), T006 (parallel)
Track B: T007 → T008 → T009 → T010

# Phase 5 — two tasks in parallel:
T020: Create .dev.vars.example
T021: Final build gate
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (create backend files)
2. Complete Phase 2: Foundational (deploy Trailbase + prepare Astro)
3. Complete Phase 3: User Story 1 (`/kontakt` page live with backend data)
4. **STOP and VALIDATE**: Visit `/kontakt` — data from Trailbase? ✓ MVP done.
5. Proceed to Phase 4 for the admin update flow.

### Incremental Delivery

1. Setup + Foundational → Backend running, Astro ready
2. User Story 1 → `/kontakt` live with cached backend data (MVP)
3. User Story 2 → Admin update flow verified end-to-end
4. Polish → Regression check, onboarding file, fallback test
