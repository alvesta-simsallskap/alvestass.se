# Research: Member Register – Data Model and Initial Import

**Branch**: `010-member-import` | **Date**: 2026-05-19

## 1. Defining "Active" Members

**Decision**: A person is considered active if they appear in the WeUnite Grupplista export as a Deltagare, Ledare, or Huvudledare in any group, OR if they hold a board role (Styrelseledamot, Ordförande, Kassör, etc.) in the IdrottOnline export without an end date.

**Rationale**: The WeUnite Grupplista was exported on 2026-05-19 and contains only currently enrolled group members — it acts as the definitive active-membership roster. Board members who may not be enrolled in a training group are identified via IdrottOnline role fields.

**Alternatives considered**:
- Use IdrottOnline `Medlem t.o.m.` (membership expiry) as the active filter: Rejected — the field is often blank for life members and unreliable for current-season filtering. WeUnite group enrolment is a more reliable signal.
- Import all current-year members: Rejected — spec requires only actively engaged persons; historical/lapsed members must be excluded.

**Implementation note**: The WeUnite file has `Start` and `Slut` date columns per group booking. A booking is current if `Slut` ≥ today (2026-05-19). Rows with a past `Slut` date are excluded.

---

## 2. Linking Source Files (Personnummer as Transient Key)

**Decision**: Use the personnummer (Swedish national ID, format `YYYYMMDD-XXXX`) as a transient join key to match WeUnite rows (column `Personnummer`) to IdrottOnline rows (column `Födelsedat./Personnr.`). The personnummer is used only in memory during the import process and is **never written to Trailbase**.

**Rationale**: The WeUnite export does not include IdrottsID (IID). The only shared identifier between the two source files is the personnummer. Once the IID is looked up from IdrottOnline, all Trailbase writes use IID as the identifier.

**Alternatives considered**:
- Match on `first_name + last_name + date_of_birth`: Rejected — name collisions exist (common Swedish names); birth date alone is not unique enough.
- Use only the WeUnite export (skip IdrottOnline): Rejected — WeUnite does not provide IdrottsID, which is the required primary key per spec FR-001.

**Privacy note**: The personnummer lookup map is built in memory and discarded at the end of the import run. Go GC guarantees the map is not persisted. No personnummer appears in logs or CLI output.

---

## 3. Group Name Normalisation

**Decision**: Strip time-slot suffixes from swim school group names using the regular expression `\s+\d{1,2}[.:]\d{2}\s*[-–]\s*\d{1,2}[.:]\d{2}$`. This converts e.g. `"Baddaren 12.55-13.40"` → `"Baddaren"` and `"Guldpingvinen 13.05–13.50"` → `"Guldpingvinen"`.

**Rationale**: Multiple WeUnite rows share the same base group name with different time slots (e.g. six rows for `Baddaren`). After stripping the time slot, all rows for the same level merge into a single `training_groups` record. Time-of-practice data belongs in a future scheduling feature per spec FR-005.

**Alternatives considered**:
- Keep time slots and deduplicate later: Rejected — results in six separate `Baddaren` records which is incorrect.
- Manual name map (hardcode each group): Rejected — fragile; the regex is more robust to future group additions.

---

## 4. Training Group Categories

**Decision**: Map source group names to categories as follows:

| Source (after normalisation) | Category |
|------------------------------|----------|
| Baddaren, Guldfisken, Guldhajen, Guldpingvinen, Silverfisken, Silverhajen, Silverpingvinen | `swim_school` |
| A-gruppen, B-gruppen | `competitive` |
| Teknikgruppen | `technique` |
| Masters | `masters` |
| Vuxencrawl | `adult` |

**Rationale**: Observed directly from the WeUnite Grupplista export. The `Sektion` and `Nivå` columns confirm these groupings (Simskola, Tävlingssim, Vuxenverksamhet).

---

## 5. Guardian Handling

**Decision**: Import guardians from the WeUnite Grupplista, which carries up to three guardian slots per row (Målsman 1/2/3 with Förnamn, Efternamn, Telefon, E-post). Each non-empty guardian slot becomes a `guardians` row linked to the swimmer's IID. If the same guardian appears for multiple siblings (same first name + last name + phone), a single `guardians` row is created and linked to each sibling's member IID separately.

**Rationale**: WeUnite is the most complete source for guardian contact data. IdrottOnline also has a `Målsman` field, but the WeUnite structured columns (up to 3 guardians with separate fields) are easier to parse reliably.

**Guardian IID**: No attempt is made to automatically match a guardian to an existing `members` IID — the `member_iid_self` column defaults to NULL for all imported guardians. A future admin workflow can link guardians who are also members.

---

## 6. Family Linking

**Decision**: Use the `Familj` column (index 36) from the IdrottOnline export. IdrottOnline groups members into families with a shared string label (typically a surname or a generated ID like `"22011829"`). Members sharing the same non-empty `Familj` value are placed in a single `families` record with each member linked via `family_members`.

**Rationale**: This is the most direct source for family information and avoids complex inference logic.

**Limitations**: Only members who appear in both the WeUnite export (active) and the IdrottOnline export (has an IID) are linked to families. Guardians are not added to the `families` table — that relationship is captured via the `guardians` table.

---

## 7. Idempotency (Re-run Safety)

**Decision**: Use `INSERT OR REPLACE` (SQLite upsert) for all `members` rows keyed on `iid`. For `training_groups`, upsert on `name`. For `guardians` and `member_training_groups`, delete-and-reinsert per member IID before inserting. For `families` and `family_members`, delete all and reinsert on each run.

**Rationale**: The import may be re-run after correcting source files (SC-005). The IID is stable and guaranteed unique per person across runs.

---

## 8. GDPR Legal Basis

**Decision**: Legal basis for storing personal data in the new tables is:

- **Art. 6(1)(b) — contractual necessity**: Member data is necessary to administer the membership agreement (access to training groups, swim school enrolment, communication about training).
- **Art. 6(1)(f) — legitimate interest**: Family linking and guardian records are necessary to operate a sports club with minor members (emergency contacts, consent management) where the legitimate interest of the club and the members' safety outweighs the data subjects' interests.

**Rationale**: The Swedish Sports Confederation (RF) requires member registers. Swim school participants are minors; guardian contact data is a safety necessity.

**Retention**: Active membership + 2 years after the member's last recorded activity. Review annually. Personal data of guardians is retained as long as the linked swimmer is an active member.

---

## 9. CSV Encoding

**Decision**: Both source files are UTF-8 encoded with semicolon delimiters. The Go `encoding/csv` reader must be configured with `Reader.Comma = ';'`. No BOM stripping is needed (confirmed by file inspection — no `\xef\xbb\xbf` prefix).

**Evidence**: File headers parsed correctly as UTF-8 in shell inspection. Swedish characters (å, ä, ö) display correctly without re-encoding.
