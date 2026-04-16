# Register över behandlingsaktiviteter (Art. 30 GDPR)

**Personuppgiftsansvarig**: Alvesta Simsällskap, 826001-1930  
**Kontakt**: kansli@alvestass.se  
**Senast uppdaterad**: 2026-04-16

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

## Anteckningar

- Tidrapporter lagras **inte** i Trailbase-databasen. E-post är det enda registret för varje inlämning (FR-007).
- Löneuppskattningar i e-post är preliminära; definitiv lön fastställs separat av lönehanteringen.
- `src/content/club/04-styrelse.md` innehåller styrelsens personuppgifter (namn, roll). Dessa är ännu inte migrerade till Trailbase och kräver separat GDPR-hantering (TODO: framtida feature).
