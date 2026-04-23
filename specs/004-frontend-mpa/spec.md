# Feature Specification: Frontend MPA Conversion

**Feature ID**: 004-frontend-mpa
**Status**: Draft
**Created**: 2026-04-23

---

## Overview

Convert the website frontend from a single-page layout — where all content sections live on one page with anchor-based navigation — into a proper Multi-Page Application (MPA), where each major content area is served as its own distinct page.

**Problem**: All content (swim school, training groups, club information) is crammed onto a single page. Visitors cannot link directly to a specific section, browser history does not reflect navigation, and search engines cannot index sections individually.

**Goal**: Give each major section its own URL, enabling direct linking, proper browser history, per-page titles and metadata, and improved discoverability.

---

## User Scenarios & Testing

### Primary Flow — Visiting the Home Page

1. Visitor lands on the home page (`/`).
2. The home page displays a hero section and short descriptions or entry points for each major area.
3. Visitor clicks "Simskola" in the navigation or on the home page.
4. Browser navigates to `/simskola` — a dedicated page listing swim school groups.
5. Visitor uses the browser back button; they return to the home page.

**Pass**: Each navigation action results in a full URL change and distinct page load.

### Secondary Flow — Direct Deep Link

1. A parent shares a link to `/simskola` with another parent.
2. Recipient opens the link and lands directly on the swim school page.
3. The page displays correctly without requiring a visit to the home page first.

**Pass**: `/simskola`, `/traning`, and `/foreningen` each work as standalone entry points.

### Edge Case — Unknown Page

1. Visitor navigates to a non-existent URL (e.g., `/om-oss`).
2. They receive a clear "page not found" response.
3. Navigation remains accessible so they can find the correct section.

**Pass**: 404 responses include site navigation; visitors are not stranded.

---

## Functional Requirements

### Navigation

1. The site navigation links to dedicated pages (`/simskola`, `/traning`, `/foreningen`) rather than in-page anchors.
2. The active page is visually indicated in the navigation bar.
3. The navigation is present on every page.

### Pages

4. A **home page** (`/`) exists and serves as the primary entry point; it may include a hero and short teasers or direct links to each section.
5. A **swim school page** (`/simskola`) displays all swim school groups, equivalent to the current `#swimschool` section.
6. A **training groups page** (`/traning`) displays all training groups, equivalent to the current `#training` section.
7. A **club page** (`/foreningen`) displays club information cards, equivalent to the current `#club` section.
8. Each page has a descriptive `<title>` that includes both the section name and the club name.
9. Each page has a relevant `<meta name="description">` appropriate to its content.

### Existing Pages

10. The existing pages (`/kontakt`, `/tidrapport`, `/integritetspolicy`, `/tack`) continue to function unchanged.

### Content

11. No content is removed during the conversion; all existing swim school groups, training groups, and club info cards remain visible on their respective pages.

---

## Success Criteria

1. Every major section is reachable via a distinct, bookmarkable URL — no section relies solely on anchor navigation.
2. Visitors can navigate directly to any section from an external link without loading the home page first.
3. The browser's back/forward buttons work correctly across all page transitions.
4. Each page displays a unique title and description in the browser tab and when shared on social media.
5. No existing content is lost compared to the current single-page layout.
6. All existing secondary pages (`/kontakt`, `/tidrapport`, etc.) continue to work without regression.

---

## Assumptions

- The URL slugs `/simskola`, `/traning`, and `/foreningen` are acceptable. (Swedish, no diacritics in URLs for simplicity.)
- The home page does not need to reproduce all section content in full — teasers or simple links to the section pages are sufficient.
- SEO metadata per page is in scope; structured data markup (JSON-LD) is out of scope for this feature.
- The member modal (login) remains globally accessible and does not change.
- Mobile/responsive behavior of each page follows the existing layout patterns.

---

## Out of Scope

- Adding new content to any section.
- Changing the visual design or component styles.
- Per-group detail pages (e.g., `/simskola/ankungen`).
- Search functionality.
- Breadcrumb navigation.
