# Implementation Plan: Frontend MPA Conversion

**Branch**: `004-frontend-mpa` | **Date**: 2026-04-23 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/004-frontend-mpa/spec.md`

---

## Summary

Split the current single-page layout (`index.astro`) into three dedicated static pages — `/simskola`, `/traning`, and `/foreningen` — backed by the existing Astro content collections. Update the navbar to use page-level navigation with active-state highlighting, and introduce a shared `Layout.astro` to eliminate per-page HTML boilerplate.

---

## Technical Context

**Language/Version**: TypeScript 5.9 (strict mode); Astro 5.17.1
**Primary Dependencies**: Astro content collections, Alpine.js 3.15.3, Bulma 1.0.4 + Sass
**Storage**: No new storage; all content from existing `src/content/` collections
**Testing**: Manual browser verification (mobile ≤ 768 px, tablet, desktop); `pnpm build` gate
**Target Platform**: Cloudflare Workers edge network (Cloudflare Pages static hosting)
**Project Type**: Astro static site (content-driven MPA)
**Performance Goals**: Lighthouse Performance ≥ 90 on mobile; LCP < 2.5 s; CLS < 0.1
**Constraints**: Zero TypeScript errors; `pnpm build` must pass; no new JS frameworks or libraries; Bulma + Sass only; all text in Swedish
**Scale/Scope**: 3 new static pages + 1 updated component + 1 new shared Layout component

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-checked after Phase 1 design.*

| Principle | Assessment | Gate Status |
|-----------|-----------|-------------|
| I — Code Quality | TypeScript strict; PascalCase components; no dead code | ✅ PASS |
| II — Testing | `pnpm build` must pass; manual browser test at all breakpoints required before PR | ⚠️ GATE: manual verification required |
| III — UX Consistency | Bulma + Sass only; Alpine.js only; Swedish text; mobile-first layout on all new pages | ✅ PASS |
| IV — Performance | All 3 new pages are static (no SSR needed); each must meet Lighthouse ≥ 90 | ✅ PASS |
| V — Trailbase | No Trailbase interaction; no schema changes | NOT TRIGGERED |
| VI — Idrottsarenan | No integration | NOT TRIGGERED |
| VII — GDPR | No personal data introduced | NOT TRIGGERED |

**Active gates**: Principle II — manual browser verification (mobile/tablet/desktop) is required before the PR is merged.

---

## Project Structure

### Documentation (this feature)

```text
specs/004-frontend-mpa/
├── plan.md              ← this file
├── research.md          ← Phase 0 output
├── data-model.md        ← Phase 1 output (no new entities; documents N/A rationale)
└── tasks.md             ← Phase 2 output (/speckit.tasks — NOT created by /speckit.plan)
```

### Source Code (changes and additions)

```text
src/
├── components/
│   ├── Layout.astro          NEW — shared page shell (html/head/body/Nav/Footer/MemberModal)
│   └── Nav.astro             UPDATED — page-level links + active-state highlighting
├── pages/
│   ├── index.astro           UPDATED — hero + section teasers linking to /simskola etc.
│   ├── simskola.astro        NEW — swim school groups page
│   ├── traning.astro         NEW — training groups page
│   └── foreningen.astro      NEW — club info page
│
│   (unchanged pages: kontakt.astro, tidrapport.astro, integritetspolicy.astro, tack.astro)
```

**Structure Decision**: Flat pages directory (Option 1 adapted). No subdirectories needed — all new pages are top-level routes. A `Layout.astro` component is introduced to satisfy Principle I (no duplicated HTML boilerplate) and Principle III (single-responsibility components).

---

## Complexity Tracking

No constitution violations requiring justification. All choices follow established project patterns.
