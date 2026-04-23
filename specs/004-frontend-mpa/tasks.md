# Tasks: Frontend MPA Conversion

**Input**: Design documents from `specs/004-frontend-mpa/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅

**Organization**: Tasks are grouped by user story to enable independent implementation and verification.

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)

---

## Phase 1: Setup — Shared Layout Component

**Purpose**: Create the shared page shell that all new pages will use. This is the foundation; no page work can begin without it.

- [ ] T001 Create `src/components/Layout.astro` — accept `title: string` and `description?: string` props; render `<html class="has-navbar-fixed-top">`, full `<head>` (charset, viewport, title as `{title} — Alvesta Simsällskap`, optional description meta, Bulma CSS import, `_global.scss` import, Figtree font import, Material Symbols CSS import), `<body x-data="{ atTop: true, open: false, modal: false }" x-on:scroll.window="...">`, `<Nav />`, `<slot />`, `<Footer />`, `<MemberModal :class="{ 'is-active': modal }" />`; do NOT include the `alpinejs-scroll-to` CDN script

**Checkpoint**: Layout.astro exists and compiles without TypeScript errors (`pnpm build`)

---

## Phase 2: Foundational — Navigation Update

**Purpose**: Replace anchor-scroll navigation with page-level links. This must be done before any section page can be verified, because the Nav is included in Layout.

**⚠️ CRITICAL**: No user story verification can be done until this phase is complete.

- [ ] T002 Update `src/components/Nav.astro` — replace each `href="#…" @click="$refs.….scrollIntoView()"` with plain page `<a>` links (`href="/"` for logo, `href="/simskola"`, `href="/traning"`, `href="/foreningen"`); derive the current pathname via `const { pathname } = Astro.url;` and add `class:list={['navbar-item', { 'is-active': pathname === '/simskola' }]}` (and equivalently for each link) so the active page is highlighted; remove the `x-scroll-to-header` directive from the `<nav>` element

**Checkpoint**: Nav renders page links; `pnpm build` passes with zero TypeScript errors

---

## Phase 3: User Story 1 — Section Pages (Priority: P1) 🎯 MVP

**Goal**: `/simskola`, `/traning`, and `/foreningen` exist as fully functional, standalone static pages. Visitors can navigate directly to any section from an external link.

**Independent Test**: Open `http://localhost:4321/simskola`, `/traning`, and `/foreningen` directly in a browser. Each page must load without errors, display its content, and show the correct nav item highlighted as active.

- [ ] T003 [P] [US1] Create `src/pages/simskola.astro` — import Layout; fetch `swimSchool` collection sorted by `order`; render `<Layout title="Simskola">` wrapping a `<Section id="simskola" title="Simskola">` with a `<Grid>` of `<SwimSchoolGroup>` cards; page is static (no `prerender = false`)
- [ ] T004 [P] [US1] Create `src/pages/traning.astro` — import Layout; fetch `trainingGroups` collection sorted by `order`; render `<Layout title="Träningsgrupper">` wrapping a `<Section id="traning" title="Träningsgrupper">` with a `<Grid>` of `<TrainingGroup>` cards; page is static
- [ ] T005 [P] [US1] Create `src/pages/foreningen.astro` — import Layout; fetch `clubInfo` collection sorted by `order`; render `<Layout title="Föreningen">` wrapping a `<Section id="foreningen" title="Om föreningen">` with a `<Grid>` of `<ClubInfo>` cards; page is static

**Checkpoint**: All three section pages load correctly at their respective URLs; navigation active state highlights the current page; `pnpm build` passes

---

## Phase 4: User Story 2 — Home Page Update (Priority: P2)

**Goal**: `index.astro` becomes a clean landing page: the Hero is retained, the three content sections are replaced with teaser cards that link to the section pages.

**Independent Test**: Open `http://localhost:4321/`. The hero displays. Below it, three teaser cards appear for Simskola, Träningsgrupper, and Om föreningen, each with a link to its section page. No section content is rendered inline. Nav shows no active item (root path has no match).

- [ ] T006 [US2] Update `src/pages/index.astro` — wrap with `<Layout title="Startsidan">`; remove the three `<Section>` blocks and their collection fetches; keep `<Hero />`; add a Bulma columns/cards teaser section below the hero with three cards: "Simskola" → `/simskola`, "Träningsgrupper" → `/traning`, "Om föreningen" → `/foreningen`; remove the `alpinejs-scroll-to` CDN script tag (now absent from Layout); remove unused collection imports

**Checkpoint**: Home page shows hero + three teaser cards; no inline section content; `pnpm build` passes; no unused imports remain

---

## Phase 5: User Story 3 — 404 Page (Priority: P3)

**Goal**: Visitors who navigate to a non-existent URL receive a friendly Swedish-language "page not found" response that still includes full site navigation so they can find what they need.

**Independent Test**: Navigate to `http://localhost:4321/finns-inte`. The page renders with the site nav and footer, displays a clear "Sidan hittades inte" heading, and provides a link back to the home page.

- [ ] T007 [US3] Create `src/pages/404.astro` — use `<Layout title="Sidan hittades inte">`; render a centered Bulma `section` with a `title` heading "Sidan hittades inte" and a subtitle, plus a `<a href="/" class="button is-primary">Gå till startsidan</a>`; page is static

**Checkpoint**: `http://localhost:4321/finns-inte` returns the 404 page with nav and a back-link; `pnpm build` passes

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final quality pass — clean up leftover Alpine scroll artefacts, verify meta descriptions, confirm build and browser compliance.

- [ ] T008 [P] Remove unused `x-ref` prop from `src/components/Section.astro` — replace `x-ref={id}` with nothing (the `id` attribute is kept for anchor compatibility); update the destructure to remove the unused `x-ref` binding
- [ ] T009 [P] Add `description` prop to each new page's `<Layout>` call — `simskola.astro`: "Simskola för barn i Alvesta Simsällskap — se alla grupper och nivåer."; `traning.astro`: "Träningsgrupper för simmare på alla nivåer i Alvesta Simsällskap."; `foreningen.astro`: "Om Alvesta Simsällskap — historia, styrelse och kontaktuppgifter."
- [ ] T010 Run `pnpm build` and confirm zero TypeScript errors and zero `astro check` warnings
- [ ] T011 Manual browser verification — open each page (`/`, `/simskola`, `/traning`, `/foreningen`, `/kontakt`, `/tidrapport`, `/404`) at mobile (≤ 768 px), tablet, and desktop widths; confirm layout, nav active state, and teaser cards render correctly at all breakpoints

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Layout)**: No dependencies — start immediately
- **Phase 2 (Nav)**: Depends on Phase 1 (Nav is rendered by Layout)
- **Phase 3 (Section pages)**: Depends on Phase 1 + Phase 2 — can then run in parallel (T003, T004, T005)
- **Phase 4 (Home page)**: Depends on Phase 1 + Phase 2; independent of Phase 3
- **Phase 5 (404)**: Depends on Phase 1 only; independent of Phase 3 and Phase 4
- **Phase 6 (Polish)**: Depends on all phases complete

### User Story Dependencies

- **US1 (Section pages)**: Requires Layout (Phase 1) + Nav (Phase 2) — the three pages can then be written in parallel
- **US2 (Home page)**: Requires Layout (Phase 1) + Nav (Phase 2) — independent of US1
- **US3 (404 page)**: Requires Layout (Phase 1) only — independent of US1 and US2

### Parallel Opportunities

- T003, T004, T005 — three independent page files, no shared state
- T008, T009 — different components, no dependencies between them
- US2 and US3 can be worked on simultaneously once Phase 2 is done

---

## Parallel Example: User Story 1

```text
After Phase 2 completes, launch all three section pages together:

Task A: "Create src/pages/simskola.astro" (T003)
Task B: "Create src/pages/traning.astro"  (T004)
Task C: "Create src/pages/foreningen.astro" (T005)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Create Layout.astro
2. Complete Phase 2: Update Nav.astro
3. Complete Phase 3: Create the three section pages
4. **STOP and VALIDATE**: Open each section page directly, confirm content and nav
5. The site is navigable to all sections — core MPA goal achieved

### Incremental Delivery

1. Phase 1 + Phase 2 → Foundation ready
2. Phase 3 → Section pages live → **MVP** (direct linking works)
3. Phase 4 → Home page teased → Improved landing experience
4. Phase 5 → 404 page → Edge case handled
5. Phase 6 → Polish pass → Ready to merge

---

## Notes

- `pnpm build` (`wrangler types && astro check && astro build`) is the merge gate — run it after each phase
- All new pages are static; do NOT add `export const prerender = false`
- All user-visible text must be in Swedish (Principle III)
- Verify layout at mobile (≤ 768 px), tablet, and desktop before marking Phase 6 complete
- The MemberModal is included in Layout.astro — no per-page changes needed for the modal
