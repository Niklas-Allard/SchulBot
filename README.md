# SchulBot

Ein selbst-hostbarer Go-Dienst der ein IServ-Postfach überwacht und auf Befehls-Tags in E-Mails reagiert.

## Unterstützte Befehle

| Tag | Beschreibung |
|-----|-------------|
| `#ki` / `/ki` | KI-Frage per Gemini oder Cloudflare AI beantworten |
| `#tasks` / `/tasks` | Google Tasks verwalten |
| `#calendar` / `/calendar` | Google Kalender abfragen und Termine anlegen |

### Verwendung

Schreibe eine E-Mail an das Bot-Konto. Der Befehlstag kann im **Betreff** oder in der **ersten nicht-leeren Zeile** des E-Mail-Textes stehen:

```
Betreff: #ki Was ist Photosynthese?
```

oder

```
Betreff: Hallo Bot

/calendar heute
```

### #tasks Unterbefehle

```
#tasks liste                   → offene Aufgaben anzeigen
#tasks add Mathe-Hausaufgaben  → neue Aufgabe anlegen
#tasks fertig Mathe            → Aufgabe (nach Titel) als erledigt markieren
```

### #calendar Unterbefehle

```
#calendar heute                         → heutige Termine
#calendar woche                         → Termine dieser Woche
#calendar neu Sport morgen 14:00        → Termin anlegen
#calendar neu Elternabend 2026-05-20    → Termin mit ISO-Datum
```

---

## Setup

### 1. Konfiguration

```bash
cp .env.example .env
# .env mit echten Werten befüllen
```

Pflichtfelder:

| Variable | Beschreibung |
|----------|-------------|
| `IMAP_HOST` / `IMAP_USERNAME` / `IMAP_PASSWORD` | IServ-Zugangsdaten |
| `SMTP_HOST` / `SMTP_USERNAME` / `SMTP_PASSWORD` / `SMTP_FROM_ADDRESS` | SMTP-Zugangsdaten |
| `AI_API_KEY` | API-Schlüssel des gewählten KI-Anbieters |

### 2. AI-Provider wählen

**Google Gemini (empfohlen):**

```env
AI_PROVIDER=gemini
AI_API_KEY=AIza...
AI_MODEL=gemini-1.5-flash
```

Schlüssel: [Google AI Studio](https://aistudio.google.com/app/apikey)

**Cloudflare Workers AI:**

```env
AI_PROVIDER=cloudflare
AI_API_URL=https://api.cloudflare.com/client/v4/accounts/ACCOUNT_ID/ai/run
AI_API_KEY=cloudflare-api-token
AI_MODEL=@cf/meta/llama-3.1-8b-instruct
```

### 3. Google OAuth (für #tasks und #calendar)

1. Google Cloud Console → Projekt erstellen → APIs aktivieren:
   - Google Tasks API
   - Google Calendar API

2. OAuth 2.0-Client-ID erstellen (Typ: Desktop-Anwendung).

3. Einmalig ein Refresh-Token holen, z. B. über den [Google OAuth Playground](https://developers.google.com/oauthplayground):
   - Scopes: `https://www.googleapis.com/auth/tasks` und `https://www.googleapis.com/auth/calendar`

4. `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REFRESH_TOKEN` in `.env` eintragen.

---

## Lokal starten

```bash
go mod tidy
make run
```

## Docker

```bash
make docker-build
make docker-up
make docker-logs
```

## Tests ausführen

```bash
make test
```

---

## Architektur

```
cmd/server/main.go          → Einstiegspunkt, Dependency-Wiring
internal/config/            → Konfiguration aus Env-Variablen
internal/app/               → Poll-Schleife, Loop-Schutz, Orchestrierung
internal/mail/imap.go       → IMAP-Client (SSL/TLS + STARTTLS)
internal/mail/smtp.go       → SMTP-Client (SSL/TLS + STARTTLS)
internal/parser/            → Tag-Erkennung im Betreff / Body
internal/commands/          → Handler-Interface + Dispatcher
internal/handlers/ki/       → #ki → KI-Provider-Aufruf
internal/handlers/tasks/    → #tasks → Google Tasks API
internal/handlers/calendar/ → #calendar → Google Calendar API
internal/ai/                → AI-Provider-Abstraktion (Gemini, Cloudflare)
internal/store/             → BoltDB-basierte Duplikatserkennung
internal/model/             → Gemeinsame Datentypen
```

### Einen neuen Befehl hinzufügen

1. In `internal/parser/parser.go` → `knownTags` erweitern.
2. Neues Package `internal/handlers/meinbefehl/` mit `CanHandle` und `Handle` anlegen.
3. Handler in `cmd/server/main.go` im `NewDispatcher`-Aufruf eintragen.

---

## Sicherheit

- Kein Self-Reply: ausgehende Adresse wird gefiltert
- Auto-Reply-Erkennung (Out of Office, Abwesenheit, …)
- Duplikatserkennung via BoltDB (persistent über Neustarts)
- Payload-Größenbegrenzung (`MAX_PAYLOAD_CHARS`)
- Antwortlängenbegrenzung (`MAX_RESPONSE_CHARS`)
- Keine Credentials im Log
