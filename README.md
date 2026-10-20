# RevU

Persönliches Life-Management-Tool: Ziele, Aufgaben, Habits, Fitness, Learning und Finanzen an einem Ort – mit Activity-Tracking, Dashboard und Dark Mode.

RevU ist die Neuauflage von **OvervU**: Das Backend ist von Java/Spring Boot nach **Go** portiert, das Frontend von React nach **Vue 3**. Datenbank, Flyway-Migrationen und REST-API sind unverändert, ein Wechsel ist ohne Datenmigration möglich (siehe [docs/MIGRATION.md](docs/MIGRATION.md)).

## Tech Stack

| | |
|---|---|
| **Backend** | Go 1.24, Gin, pgx v5, sqlc, JWT (golang-jwt), bcrypt |
| **Datenbank** | PostgreSQL, Schema über die unveränderten Flyway-Migrationen |
| **Frontend** | Vue 3 (`<script setup>`), TypeScript, Vite, Tailwind CSS 4, Vue Router, Pinia, Chart.js (vue-chartjs) |
| **Infrastruktur** | Docker, Docker Compose, GitHub Actions, Google Cloud Run + Cloud SQL |

## Features

- **Ziele** – langfristige Lebensziele verfolgen
- **Aufgaben** – To-dos mit Fälligkeitsdatum
- **Habits** – tägliche und wöchentliche Gewohnheiten (wöchentliche einmal pro Woche abhakbar)
- **Fitness** – Workouts mit Übungen und Sätzen
- **Learning** – Lernprojekte und Lernzeit
- **Finanzen** – Einnahmen, Ausgaben, Monatsbilanz
- **Dashboard** – Kennzahlen und Charts aller aktiven Bereiche
- **Profil** – Bereiche ein-/ausschalten (verschwinden dann aus Navigation *und* Dashboard), Dark Mode

## Projektstruktur

```
backend/
  cmd/server/            Einstiegspunkt
  internal/
    config/              Konfiguration über Umgebungsvariablen
    database/            Connection-Pool, Flyway-Versionscheck, Transaktionen
    auth/                JWT, bcrypt, Auth-Middleware
    httpx/               Fehlerformat, JSON-Binding, Validierung, Datumstypen
    decimal/             exakte Dezimalzahlen für Geld/Gewichte (statt BigDecimal)
    store/               von sqlc generierter Datenbankcode (+ queries/*.sql)
    user/ goal/ task/ activity/ habit/ fitness/ learning/ finance/
    server/              Router, CORS, Logging
  migrations/            Flyway-Skripte V1–V14 (byte-identisch zu OvervU)
  contracttest/          Black-Box-Tests für die komplette REST-API
frontend/
  src/
    api/ types/          API-Client und Typen
    stores/              Pinia: Auth/Feature-Toggles, Theme
    router/              Routen + Guards
    layouts/ pages/ components/
docs/                    Migration vom alten Stack, Deployment
```

## Lokale Entwicklung

**Voraussetzungen:** Go 1.24+, Node.js 22+ (CI nutzt 24), Docker

```bash
# 1. Datenbank starten und Migrationen anwenden (Flyway läuft im Container)
docker compose up -d postgres
docker compose run --rm flyway

# 2. Backend (http://localhost:8080)
cd backend
APP_ENV=dev go run ./cmd/server

# 3. Frontend (http://localhost:5173), separates Terminal
cd frontend
cp .env.example .env
npm install
npm run dev
```

Mit `APP_ENV=dev` verbindet sich das Backend ohne weitere Angaben mit `postgres://overvu:overvu@localhost:5432/overvu` und nutzt ein Entwicklungs-JWT-Secret.

### Komplett in Docker

```bash
docker compose up --build
```

Frontend auf http://localhost:3000, API auf http://localhost:8080/api. Der `flyway`-Service migriert die Datenbank, bevor das Backend startet.

## Konfiguration (Backend)

| Variable | Bedeutung | Standard |
|---|---|---|
| `DATABASE_URL` | Postgres-URL oder Key/Value-DSN | – |
| `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`, `DB_SSLMODE` | Alternative zu `DATABASE_URL`; `DB_HOST=/cloudsql/<instanz>` für Cloud SQL | – |
| `JWT_SECRET` | HMAC-Secret, mind. 32 Bytes. **Gleiches Secret wie bei OvervU ⇒ bestehende Logins bleiben gültig** | Pflicht (außer `APP_ENV=dev`) |
| `JWT_EXPIRATION_MS` | Token-Laufzeit | `86400000` (24 h) |
| `CORS_ALLOWED_ORIGINS` | Kommagetrennte Liste erlaubter Frontend-Origins | `http://localhost:5173` |
| `APP_TIMEZONE` | Zeitzone für „heute“ (Habits, Monatsgrenzen, Tagesstatistik) | `Europe/Berlin` |
| `PORT` | HTTP-Port (setzt Cloud Run automatisch) | `8080` |
| `SCHEMA_VERSION` / `SKIP_SCHEMA_CHECK` | Erwartete Flyway-Version / Check abschalten | `14` / `false` |
| `APP_ENV` | `dev` oder `prod` | `prod` |

Das Backend migriert **nicht** selbst. Beim Start prüft es nur, ob Flyway die erwartete Version erfolgreich eingespielt hat, und bricht sonst mit einer klaren Meldung ab.

## Tests

```bash
cd backend
go test ./...                       # Unit-Tests

# Komplette API-Vertragssuite gegen eine migrierte Datenbank (startet die API im Prozess)
TEST_DATABASE_URL="postgres://overvu:overvu@localhost:5432/overvu?sslmode=disable" go test ./...

# Dieselbe Suite gegen einen laufenden Server – z. B. das alte Spring-Backend
CONTRACT_BASE_URL=http://localhost:8080 CONTRACT_LEGACY=1 go test ./contracttest/ -count=1
```

`CONTRACT_LEGACY=1` überspringt die Tests für Verhalten, das in RevU bewusst korrigiert wurde.

```bash
cd frontend
npm run lint
npm run build        # inkl. Type-Check mit vue-tsc
```

### Datenbank-Code neu generieren

Neue oder geänderte Abfragen kommen nach `backend/internal/store/queries/*.sql`, danach:

```bash
cd backend && sqlc generate
```

Die CI prüft mit `sqlc diff`, dass der generierte Code aktuell ist.

### Neue Migrationen

Neue Flyway-Skripte (`V15__...sql`) kommen nach `backend/migrations/`. Danach `SCHEMA_VERSION` anheben und sqlc neu generieren. Angewendet werden sie mit Flyway (`docker compose run --rm flyway` lokal, für Cloud SQL siehe [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)).

## Deployment

Google Cloud Run (Frontend und Backend als eigene Services) mit Cloud SQL: [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md).

## Offene Ideen (aus OvervU übernommen)

- Mehrere Sätze auf einmal eintragen; Sätze/Übungen im UI wieder entfernen (die API kann es bereits)
- Aufgaben als Teilziele von Lebenszielen
- Farben im Dark Mode weiter verfeinern (Charts nutzen bereits hellere Töne)
