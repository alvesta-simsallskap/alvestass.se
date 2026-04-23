# Research: Frontend MPA Conversion

**Feature**: 004-frontend-mpa
**Date**: 2026-04-23

---

## Decision 1 — URL Slugs

**Decision**: `/simskola`, `/traning`, `/foreningen`

**Rationale**: Swedish words without diacritics, matching the section labels already used in the existing anchor IDs (`#swimschool`, `#training`, `#club`). Avoids `%C3%A4` percent-encoding in shared URLs. All three slugs are immediately recognisable to Swedish users and consistent with the existing `/kontakt` page naming convention.

**Alternatives considered**:
- `/swim-school`, `/training`, `/about` — English slugs; rejected because all user-visible text is in Swedish (Principle III).
- `/simskola`, `/träning`, `/föreningen` — Swedish with diacritics; technically valid in modern browsers but complicates copy-paste sharing and server log readability.

---

## Decision 2 — Shared Layout Component

**Decision**: Introduce `src/components/Layout.astro` to hold the page shell (`<html>`, `<head>`, `<body x-data="…">`, `<Nav />`, `<Footer />`, `<MemberModal />`).

**Rationale**: Without a Layout component, creating 3 new pages would duplicate 20+ lines of identical boilerplate (charset, viewport, font imports, Alpine.js body init, Nav, Footer, MemberModal). Principle I prohibits dead code; repeated identical code across files is structural dead code. A single-responsibility Layout component is the canonical Astro pattern and explicitly supported by the constitution.

**Props**:
- `title: string` — injected into `<title>` as `{title} — Alvesta Simsällskap`
- `description?: string` — injected into `<meta name="description">`

**Alternatives considered**:
- Keep boilerplate per-page — rejected; violates Principle I (dead code) and makes future header changes require edits in every file.
- Astro base layout via `layouts/` directory — functionally identical; kept in `components/` to match the existing project structure (all reusable `.astro` files live in `src/components/`).

---

## Decision 3 — Nav Active-State Highlighting

**Decision**: Use `Astro.url.pathname` in `Nav.astro` to compare against each link's `href`. Apply Bulma's `is-active` modifier class when they match.

**Rationale**: Astro exposes `Astro.url` at build time for static pages and at request time for SSR pages — both cases work. This requires zero JavaScript and no additional dependencies.

**Implementation note**: Pass `pathname` from the parent layout via a prop so `Nav.astro` stays a pure presentational component with no direct coupling to the Astro request context. Alternatively, `Astro.url` can be accessed directly inside `Nav.astro` since it is available in all Astro component server scripts.

**Alternatives considered**:
- Alpine.js `window.location.pathname` check — requires JavaScript; violates Principle IV (minimise client-side JS) when a static solution exists.
- Hard-coding active state per page — fragile; rejected.

---

## Decision 4 — Alpine.js Scroll-to Behaviour

**Decision**: Remove `alpinejs-scroll-to` CDN script and `x-ref`-based `scrollIntoView()` calls from the Nav. Replace with standard `<a href="/simskola">` etc. Keep the logo link pointing to `/` (standard page link, no scroll).

**Rationale**: The scroll-to library and Alpine `x-ref` bindings were only needed for in-page anchor navigation. In a true MPA, browser-native page navigation replaces them entirely. The logo still links to the home page — no special scroll behaviour is needed.

**What changes in Nav.astro**:
- `href="#swimschool" @click="$refs.swimschool.scrollIntoView()"` → `href="/simskola"`
- `href="#training" @click="$refs.training.scrollIntoView()"` → `href="/traning"`
- `href="#club" @click="$refs.club.scrollIntoView()"` → `href="/foreningen"`
- `href="#start" @click="$refs.start.scrollIntoView()"` → `href="/"`
- Remove `<script is:inline defer src="…alpinejs-scroll-to…">` from Layout

**Alternatives considered**:
- Keeping the CDN script for potential future use — rejected; Principle I prohibits unused code.

---

## Decision 5 — Home Page Content After Section Extraction

**Decision**: `index.astro` retains the full `<Hero />` component (which already contains news boxes). Add a simple three-column Bulma card row below the hero that teases each section with a title, brief description, and a "Läs mer →" link.

**Rationale**: The Hero already provides the primary landing experience. A short teaser row gives first-time visitors clear navigation paths without duplicating the full section content. This also satisfies the spec requirement that the home page provides "entry points for each major area."

**Teasers**:
| Section | Heading | Link |
|---------|---------|------|
| Simskola | "Simskola" | `/simskola` |
| Träningsgrupper | "Träningsgrupper" | `/traning` |
| Föreningen | "Om föreningen" | `/foreningen` |

**Alternatives considered**:
- Hero only, no teasers — acceptable but leaves the home page without any navigation to sections for users who scroll past the nav.
- Reproducing all section cards on the home page — duplicates content and was explicitly out of scope.
