# Data Model: Frontend MPA Conversion

**Feature**: 004-frontend-mpa
**Date**: 2026-04-23

---

## No New Data Entities

This feature introduces no new data entities, database tables, or schema changes.

All content displayed on the new pages (`/simskola`, `/traning`, `/foreningen`) is sourced from the **existing Astro content collections**:

| Collection | Directory | Used on page |
|-----------|-----------|-------------|
| `swimSchool` | `src/content/swim-school/` | `/simskola` |
| `trainingGroups` | `src/content/training-groups/` | `/traning` |
| `clubInfo` | `src/content/club/` | `/foreningen` |

These collections are already typed and validated by Astro's content schema system. No modifications to collection schemas are needed.

## Trailbase

No Trailbase tables are added, modified, or queried by this feature. Principle V (Trailbase as sole backend) is not triggered.

## GDPR

No personal data is introduced. Principle VII is not triggered.
