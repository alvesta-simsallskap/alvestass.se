# Register över behandlingsaktiviteter (Art. 30 GDPR)

**Personuppgiftsansvarig**: Alvesta Simsällskap, 826001-1930  
**Kontakt**: kansli@alvestass.se  
**Senast uppdaterad**: 2026-05-19

---

## 1. Instruktörsuppgifter (`instructors`-tabellen)

| Fält | Innehåll |
|------|----------|
| **Ändamål** | Beräkna preliminär löneuppskattning i tidrapporter |
| **Datakategorier** | E-postadress (identifierare), timlön simskola (SEK/h), timlön tävlingsträning (SEK/h, kan saknas) |
| **Registrerade** | Instruktörer anställda av Alvesta Simsällskap |
| **Rättslig grund** | Art. 6.1 b – avtalsnödvändighet (anställningsavtal) |
| **Lagringstid** | Till och med anställningens upphörande + 1 år (bokföringsändamål) |
| **Lagringsplats** | Trailbase SQLite-databas på fly.io (region: arn / Stockholm, EU) |
| **Personuppgiftsbiträde** | fly.io, Inc. – DPA tecknas via fly.io:s standardavtal |
| **Raderingsväg** | Administratör raderar rad via Trailbase admin-gränssnitt |
| **Tillgångsbegränsning** | Läsåtkomst kräver giltig service-user-token; ingen publik åtkomst |

---

## 2. Cloudflare Workers (databehandlare)

| Fält | Innehåll |
|------|----------|
| **Roll** | Personuppgiftsbiträde — behandlar e-postadresser i samband med tidrapportering |
| **Ändamål** | Vidarebefordran av tidrapporter via e-post; uppslagnig av instruktörsuppgifter |
| **DPA-status** | Cloudflare Data Processing Addendum (DPA) ingår i Cloudflare-avtalet (automatiskt för Workers-kunder) |
| **Personuppgifter som berörs** | E-postadress (instruktör) — används som uppslagsnyckeln; skickas vidare till Mailjet för e-postleverans |
| **Lagringstid** | Inte lagrade i Cloudflare Workers (stateless) — transitbehandling enbart |

---

## 3. Mailjet (databehandlare, e-postleverans)

| Fält | Innehåll |
|------|----------|
| **Roll** | Personuppgiftsbiträde — levererar tidrapport-e-post till lönehanteringens inkorg |
| **Ändamål** | Leverans av tidrapport-e-post (instruktörens namn, e-post, valda pass ingår i e-postinnehållet) |
| **DPA-status** | Mailjet Data Processing Agreement (DPA) signeras separat via Mailjet-kontot |
| **Personuppgifter som berörs** | Namn och e-postadress (instruktör), tidsredovisningsdata per rapport |
| **Lagringstid** | Mailjet behåller loggar i 60 dagar (Mailjet-standardpolicy); e-posten i sig levereras till mottagarens inkorg |
| **Raderingsväg** | Kontakta Mailjet-support eller administratören av lönehanteringens inkorg |

---

---

## 4. Memberregister — `members`-tabellen (feature 010-member-import)

| Fält | Innehåll |
|------|----------|
| **Ändamål** | Administrera aktiva medlemmar (simmare, styrelseledamöter) för framtida inloggning, närvaro och kommunikation |
| **Datakategorier** | Förnamn, efternamn, kön, födelsedag, postort, e-postadress, telefonnummer, IdrottsID (IID), inträdesår, rollflaggor |
| **Registrerade** | Aktiva simmare (Deltagare i WeUnite) och formella styrelseledamöter (IdrottOnline) |
| **Rättslig grund** | Art. 6.1 b – avtalsnödvändighet (medlemsavtal) |
| **Lagringstid** | Till och med att medlemskapet upphör + 2 år; granskas årligen |
| **Lagringsplats** | Trailbase SQLite-databas på fly.io (region: arn / Stockholm, EU) |
| **Personuppgiftsbiträde** | fly.io, Inc. – DPA via fly.io:s standardavtal |
| **Raderingsväg** | `alvestass-admin delete-member --iid <IID>` (kaskaderar till guardians, member_training_groups, family_members) |
| **Tillgångsbegränsning** | Autentiserad åtkomst krävs — ingen publik API-åtkomst; tabellen konfigureras som authenticated-only i Trailbase admin |

---

## 5. Vårdnadshavare — `guardians`-tabellen (feature 010-member-import)

| Fält | Innehåll |
|------|----------|
| **Ändamål** | Lagra kontaktuppgifter till vårdnadshavare för aktiva minderåriga simmare |
| **Datakategorier** | Förnamn, efternamn, telefonnummer, e-postadress; referens till simmarens IID |
| **Registrerade** | Juridiska vårdnadshavare till aktiva minderåriga simmare (< 18 år) |
| **Rättslig grund** | Art. 6.1 f – berättigat intresse (föreningens säkerhetsansvar för minderåriga) |
| **Lagringstid** | Så länge länkad simmare är aktiv medlem + 2 år |
| **Lagringsplats** | Trailbase SQLite-databas på fly.io (region: arn / Stockholm, EU) |
| **Personuppgiftsbiträde** | fly.io, Inc. – DPA via fly.io:s standardavtal |
| **Raderingsväg** | Kaskaderas vid radering av länkat `members`-record; eller direkt radradering via Trailbase admin-gränssnitt |
| **Tillgångsbegränsning** | Autentiserad åtkomst krävs — ingen publik API-åtkomst |

---

## 6. Träningsgrupper — `training_groups`-tabellen (feature 010-member-import)

| Fält | Innehåll |
|------|----------|
| **Ändamål** | Namngiven lista över föreningens träningsgrupper (utan tidsluckor) |
| **Datakategorier** | Gruppnamn, kategori — inga personuppgifter |
| **Rättslig grund** | Ej tillämplig (inga personuppgifter) |
| **Lagringstid** | Bevaras löpande (operationsdata) |
| **Lagringsplats** | Trailbase SQLite-databas på fly.io (region: arn / Stockholm, EU) |
| **Tillgångsbegränsning** | Läsning utan autentisering acceptabel (inga personuppgifter); skrivning kräver admin-CLI |

---

## 7. Gruppmedlemskap — `member_training_groups`-tabellen (feature 010-member-import)

| Fält | Innehåll |
|------|----------|
| **Ändamål** | Kopplar `members`-poster till deras träningsgrupper med roll (deltagare) |
| **Datakategorier** | Referens till member.iid och training_group.id, roll — inga egna personuppgifter utöver FK |
| **Rättslig grund** | Samma som `members` (Art. 6.1 b) — behandlingens syfte är att administrera gruppmedlemskap |
| **Lagringstid** | Bevaras så länge `members`-posten finns; kaskaderas vid radering |
| **Lagringsplats** | Trailbase SQLite-databas på fly.io (region: arn / Stockholm, EU) |
| **Tillgångsbegränsning** | Autentiserad åtkomst krävs (avslöjar vem som finns i en grupp) |

---

## 8. Familjeenheter — `families`-tabellen (feature 010-member-import)

| Fält | Innehåll |
|------|----------|
| **Ändamål** | Representerar en familjeenhet för identifiering av syskon och samlad kommunikation |
| **Datakategorier** | Källetikett från IdrottOnline (familjeidentifierare) — inga direkta personuppgifter |
| **Rättslig grund** | Art. 6.1 f – berättigat intresse (familjerelaterade rabatter och kommunikation) |
| **Lagringstid** | Bevaras så länge länkade `members`-poster finns; granskas vid uppdatering av register |
| **Lagringsplats** | Trailbase SQLite-databas på fly.io (region: arn / Stockholm, EU) |
| **Tillgångsbegränsning** | Autentiserad åtkomst krävs |

---

## 9. Familjemedlemmar — `family_members`-tabellen (feature 010-member-import)

| Fält | Innehåll |
|------|----------|
| **Ändamål** | Kopplar `members`-poster till deras familjeenhet |
| **Datakategorier** | Referens till family.id och member.iid — inga egna personuppgifter utöver FK |
| **Rättslig grund** | Samma som `families` (Art. 6.1 f) |
| **Lagringstid** | Kaskaderas vid radering av länkad `members`- eller `families`-post |
| **Lagringsplats** | Trailbase SQLite-databas på fly.io (region: arn / Stockholm, EU) |
| **Tillgångsbegränsning** | Autentiserad åtkomst krävs |

---

## Anteckningar

- Tidrapporter lagras **inte** i Trailbase-databasen. E-post är det enda registret för varje inlämning (FR-007).
- Löneuppskattningar i e-post är preliminära; definitiv lön fastställs separat av lönehanteringen.
- `src/content/club/04-styrelse.md` innehåller styrelsens personuppgifter (namn, roll). Dessa är ännu inte migrerade till Trailbase och kräver separat GDPR-hantering (TODO: framtida feature).
