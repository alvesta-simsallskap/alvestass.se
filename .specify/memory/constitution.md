<!--
SYNC IMPACT REPORT
==================
Version change: 1.1.0 → 1.2.0

Version bump rationale: MINOR — one new principle added (VII: GDPR & Data
Protection). No existing principles were removed or redefined. Principle VI
updated with a cross-reference to Principle VII to eliminate duplication.

Modified principles:
  - Principle VI: "Personal data (GDPR)" bullet replaced with a cross-reference
    to the new Principle VII so GDPR rules are defined in one place only.

Added sections:
  - Principle VII: GDPR Compliance & Data Protection

Removed sections: None

Templates requiring updates:
  ✅ .specify/templates/plan-template.md — Constitution Check is generic;
     no structural change needed. Plans for features involving personal data
     MUST include a "GDPR gate" note in the Constitution Check section.
  ✅ .specify/templates/spec-template.md — no structural change needed; the
     Key Entities section already calls for data-minimization decisions. Authors
     MUST now also identify the GDPR legal basis for each personal-data entity.
  ✅ .specify/templates/tasks-template.md — no structural change needed;
     Polish phase already includes "Security hardening". Authors should add a
     GDPR review task for any feature that introduces personal data.

Follow-up TODOs:
  - TODO(TEST_FRAMEWORK): Unit-test framework (e.g., Vitest) still not installed.
  - TODO(CI_PIPELINE): No CI/CD pipeline found.
  - TODO(TRAILBASE_DEPLOYMENT): Hosting target for Trailbase not yet decided.
  - TODO(IDROTTSARENAN_API): Idrottsarenan access mechanism not yet confirmed.
  - TODO(GDPR_REGISTER): A register of processing activities
    (behandlingsregister, GDPR Art. 30) covering all Trailbase tables and
    external integrations MUST be created before any personal data is stored.
  - TODO(GDPR_DPA): Data Processing Agreements with Cloudflare and the email
    provider MUST be confirmed and documented.
  - TODO(INTEGRITETSPOLICY): A Swedish-language privacy notice MUST be published
    at /integritetspolicy before any personal data is collected or displayed.
-->

# Alvesta Simsällskap Website — Project Constitution

## Preamble

The Alvesta Simsällskap website (alvestass.se) serves the members, caregivers,
swimmers, instructors, and board members of Alvesta Swim Club — founded in 1921.
It is also a place for the public to learn about the club. It is a content-driven
Astro site deployed on Cloudflare's edge network, combining static content with
lightweight interactive features, an internal time-reporting tool, and a
Trailbase-backed persistence layer.

The site integrates with **Idrottsarenan**, the digital platform operated by the
Swedish Sports Confederation (Riksidrottsförbundet) and used by all Swedish sport
clubs. Only the minimum data necessary to power displayed features is fetched from
and stored locally from that system.

All stored data and content displayed to users MUST be handled in accordance with
GDPR as far as the technical constraints and the voluntary nature of the club allow.

This constitution establishes the non-negotiable principles that govern all
development work on this project. Every feature, bugfix, and refactor MUST be
evaluated against these principles before being considered complete.

---

## Core Principles

### I. Code Quality

Every piece of code merged to `main` MUST be type-safe, clearly structured,
and free of dead code.

- TypeScript strict mode MUST be enabled and MUST produce zero errors.
  `astro check` MUST pass before any merge is considered complete.
- Explicit types MUST be used for all function parameters, return values, and
  Astro component props. The `any` type MUST NOT appear in production code.
- Astro components MUST follow single-responsibility: one component renders one
  distinct UI concern. Components serving multiple unrelated concerns MUST be
  split.
- Naming MUST be consistent: PascalCase for `.astro` components and TypeScript
  types; camelCase for variables and functions; kebab-case for file names and
  CSS class names.
- Dead code — unreachable branches, unused imports, commented-out blocks — MUST
  be removed before merging. Leaving dead code "just in case" is not permitted.

### II. Testing Standards

All business logic MUST be covered by automated tests. All UI changes MUST be
manually verified in a browser before merging.

- The full build pipeline (`pnpm build`, which runs `wrangler types &&
  astro check && astro build`) MUST succeed with zero errors on every branch
  before merge. A failing build is a blocking error.
- Business-logic modules — currently `src/lib/salary.ts`,
  `src/lib/timeReportValidation.ts`, and any future `src/lib/*.ts` — MUST have
  unit tests covering the primary path, boundary values, and known error
  conditions.
- UI changes (components, pages, styles) MUST be manually exercised using
  `pnpm dev` before a PR is submitted. The primary user flow AND at least one
  edge case MUST be verified in a real browser.
- No change that touches the time-reporting workflow may be marked done without
  an end-to-end manual test of the complete form submission flow, including the
  email delivery path.
- TODO(TEST_FRAMEWORK): A unit-test framework (e.g., Vitest) is not yet
  installed. It MUST be added before any new logic is introduced in `src/lib/`.
  Until then, logic MUST be kept simple enough to be verified by inspection.

### III. User Experience Consistency

Every user-visible element MUST use the established Bulma + Sass design system
and MUST behave consistently with existing interaction patterns.

- Bulma CSS MUST be the sole layout and component framework. Custom CSS MUST be
  written in Sass (`.scss`) and MUST extend or override Bulma's design tokens —
  raw hard-coded values that bypass the token system are not permitted.
- All client-side interactivity MUST use Alpine.js. No additional JavaScript
  framework or library may be introduced without an explicit architectural
  decision documented in the affected PR.
- All user-visible text MUST be written in Swedish. Exceptions require explicit
  justification (e.g., an explicitly international or English-language feature).
- Responsive design MUST be mobile-first. Every new component and page MUST be
  verified at mobile (≤ 768 px), tablet, and desktop breakpoints before merge.
- Editable content MUST be managed through Astro content collections
  (`src/content/`). Hard-coded content in `.astro` files is only permissible for
  truly static strings that will never be edited (e.g., the HTML `<title>`).
- Icons MUST use the Material Symbols Rounded font already included in the
  project. Adding a second icon library requires explicit justification.

### IV. Performance Requirements

Every page MUST achieve a Lighthouse Performance score of 90 or above on mobile.
JavaScript delivered to the browser MUST be kept to a minimum.

- Astro's static generation MUST be used for every page that does not require
  server-side data at request time. SSR via Cloudflare Workers is only permitted
  where static generation is technically impossible (e.g., the
  `/api/send-time-report` endpoint or Trailbase-backed API routes).
- Client-side JavaScript is limited to Alpine.js interactions already in the
  project. No new client-side bundles may be added. Third-party scripts MUST be
  loaded with `defer` and evaluated for payload size and runtime cost before
  adoption.
- Images MUST be served in modern formats (WebP or AVIF). Any image wider than
  400 px MUST include explicit `width` and `height` attributes to prevent
  Cumulative Layout Shift.
- Core Web Vitals targets (measured on mobile): LCP < 2.5 s, CLS < 0.1,
  INP < 200 ms. Lighthouse Performance ≥ 90. A change that regresses any of
  these below threshold MUST NOT be merged until the regression is resolved.
- Cloudflare cache headers MUST be set correctly for static assets. Setting
  `Cache-Control: no-store` on assets that are safe to cache is not permitted.

### V. Backend Architecture (Trailbase)

**Trailbase** is the sole backend for all server-side data persistence and
authenticated API access. No alternative database layer or custom server runtime
may be introduced without an explicit architectural decision documented in an
ADR committed to the repository.

- Trailbase MUST be the single source of truth for all mutable application data.
  No other database (PostgreSQL, PlanetScale, Supabase, etc.) may be added.
- All schema changes MUST be applied through Trailbase's migration system.
  Direct SQL mutations outside of migration files are not permitted in production.
- The Cloudflare Workers SSR layer MUST delegate all data reads and writes to
  Trailbase's REST API. Business logic that directly constructs raw SQL MUST NOT
  exist in `src/` or worker code.
- Trailbase's built-in authentication MUST be used for any feature requiring
  user identity. Rolling a custom auth scheme is not permitted.
- TODO(TRAILBASE_DEPLOYMENT): The hosting target for the Trailbase instance
  (self-hosted VPS, fly.io, etc.) MUST be decided and documented before the
  first persistence feature is implemented. Until then, no feature that depends
  on Trailbase MAY be merged to `main`.
- API keys and connection strings for Trailbase MUST be stored as Cloudflare
  Worker secrets (via `wrangler secret put`) and MUST NOT be committed to the
  repository.

### VI. External Data Integration (Idrottsarenan / Riksidrottsförbundet)

The site MAY integrate with **Idrottsarenan** — the Swedish Sports Confederation's
platform used by all sport clubs — to display member, competition, or activity
data. Such integration MUST follow a strict data-minimization discipline.

- **Minimum necessary data**: Only fields that are directly rendered on a
  user-visible page or required for a documented business rule MUST be fetched
  and stored. Fetching full member records "for future use" is not permitted.
- **No redundant local copies**: Data that can be fetched on-demand from
  Idrottsarenan at acceptable latency MUST NOT be stored locally in Trailbase.
  Local caching is only permitted where Idrottsarenan availability or rate limits
  make on-demand fetching impractical; in that case the cache TTL and eviction
  policy MUST be documented in the feature spec.
- **Purpose limitation**: Idrottsarenan data MUST only be used for the purpose
  explicitly stated in the feature specification that introduced the integration.
  Repurposing the data for new features requires a new documented decision.
- **Personal data**: Any Idrottsarenan field that constitutes personal data
  (names, birthdates, contact details, etc.) MUST be identified in the feature
  spec's Key Entities section and handled according to Principle VII.
- TODO(IDROTTSARENAN_API): The access mechanism (REST API, bulk export, webhook)
  provided by Idrottsarenan MUST be confirmed and documented before any
  integration feature is planned.
- All Idrottsarenan credentials and API tokens MUST be stored as Cloudflare
  Worker secrets and MUST NOT be committed to the repository.

### VII. GDPR Compliance & Data Protection

All personal data stored in Trailbase or displayed on any page MUST be handled
in compliance with GDPR (Regulation (EU) 2016/679) and the Swedish complementary
law (Lag (2018:218) med kompletterande bestämmelser till EU:s
dataskyddsförordning). The "as far as possible" qualifier in user impact applies
only to technical constraints outside the project's control (e.g., upstream
Idrottsarenan data structures); it does NOT relax rules the project can enforce
itself.

- **Legal basis (Art. 6 GDPR)**: Every Trailbase table that stores personal data
  MUST have its legal basis (e.g., contractual necessity for time reports;
  legitimate interest for published competition results; consent for photos)
  documented in the feature spec's Key Entities section before the table is
  created. Features whose legal basis is unclear MUST NOT be merged until the
  basis is confirmed.
- **Data minimization (Art. 5(1)(c))**: No personal data field may be stored in
  Trailbase unless it is directly used in a live, user-visible feature or a
  documented internal business rule. Storing data "for analytics" or "for the
  future" is not permitted.
- **Storage limitation (Art. 5(1)(e))**: Every Trailbase table containing
  personal data MUST define a retention period in its migration comment. Data
  exceeding its retention period MUST be deleted or anonymized. A retention
  schedule MUST be part of each feature spec that introduces personal data.
- **Access control**: Personal data MUST NOT be served by public (unauthenticated)
  API routes or rendered in statically generated pages. Any page or endpoint that
  exposes personal data MUST be gated behind Trailbase authentication.
- **No personal data in URLs, logs, or error messages**: Full names, emails,
  membership IDs, and similar identifiers MUST NOT appear in URL query strings,
  Cloudflare Worker log output, or client-visible error messages.
- **Photography and identifiable images**: Photographs of identifiable individuals
  MUST NOT be published without documented consent. Consent records MUST be
  retained for as long as the image is live plus two years. Images of minors
  require explicit parental consent.
- **Data subject rights (Arts. 15–22)**: Any personal data stored in Trailbase
  MUST be erasable on request within 30 days. Before implementing a new personal-
  data feature, the developer MUST confirm that a deletion path exists (e.g., an
  admin UI or a documented manual SQL procedure in `docs/gdpr-ops.md`).
- **Privacy notice**: A Swedish-language privacy policy (integritetspolicy) MUST
  be published at `/integritetspolicy` and linked in the site footer before any
  personal data is collected or displayed. It MUST be updated whenever a new
  category of personal data is introduced.
  TODO(INTEGRITETSPOLICY): Page not yet created; MUST be implemented before the
  first feature that handles personal data goes live.
- **Processing register (Art. 30)**: A record of processing activities
  (behandlingsregister) covering all Trailbase tables and external integrations
  MUST be maintained in `docs/gdpr-register.md`.
  TODO(GDPR_REGISTER): Document not yet created; MUST be in place before any
  personal data is stored.
- **Data processor agreements**: Cloudflare (edge/Workers), the email provider,
  and any other third-party processor that may handle personal data MUST have a
  valid Data Processing Agreement (DPA) in place. Their DPA status MUST be noted
  in `docs/gdpr-register.md`.
  TODO(GDPR_DPA): DPA status for Cloudflare and the email provider must be
  confirmed.

---

## Governance

### Amendment Procedure

1. Any contributor may propose an amendment by opening a PR that modifies this
   file.
2. The amendment MUST include an updated `Last Amended` date and a bumped
   version number following the policy below.
3. The PR description MUST explain the motivation and link to any relevant
   incidents or prior decisions.
4. At least one other contributor MUST review and approve before merging.

### Versioning Policy

`MAJOR.MINOR.PATCH` adapted for governance:

- **MAJOR**: Removal or fundamental redefinition of an existing principle.
- **MINOR**: Addition of a new principle or material expansion of an existing
  one's scope.
- **PATCH**: Clarifications, wording improvements, or non-semantic refinements
  that do not change intent.

### Compliance Review

- Each PR description MUST include a brief note confirming which principles
  were considered and whether any gates were triggered.
- The build gate (Principle II) SHOULD be enforced automatically via CI.
  All other principles rely on reviewer diligence until CI is configured.
- A full constitution review SHOULD be conducted at the start of each major
  development cycle, and at minimum once per calendar year.
- Any change introducing personal data processing MUST also trigger a review
  against Principle VII before the PR is approved.

---

**Version**: 1.2.0 | **Ratified**: 2026-04-15 | **Last Amended**: 2026-04-16
