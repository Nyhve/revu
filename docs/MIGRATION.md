# Migration von OvervU (Spring Boot + React) zu RevU (Go + Vue)

## Was gleich bleibt

- **Datenbank:** dieselben 14 Tabellen, die Flyway-Skripte V1–V14 sind byte-identisch übernommen. Es gibt keine Datenmigration und keine Schemaänderung; die vorhandene `flyway_schema_history` bleibt gültig.
- **REST-API:** alle 46 Endpunkte mit denselben Pfaden, Methoden, Statuscodes, JSON-Feldnamen und -Formaten:
  - Datumswerte als `"2026-10-20"`, Zeitstempel ohne Zeitzone wie `"2026-10-20T07:04:00.123456"` (UTC)
  - Geldbeträge und Gewichte als JSON-Zahlen (`12.50`)
  - Enums als Strings (`"ACTIVE"`, `"WEEKLY"`, …)
  - leere Listen als `[]`
  - Fehler als `{"timestamp": "...", "status": 404, "message": "Goal not found with id: 5"}`
  - Validierungsfehler als `"title: Title must not be blank, status: Status must not be null"`
- **Passwörter:** Spring hat BCrypt-Hashes (`$2a$10$…`) gespeichert; Go prüft genau dieses Format. Alle Konten funktionieren weiter.
- **JWT:** gleicher Aufbau (`sub` = E-Mail, `userId`, `iat`, `exp`, 24 h). Der Algorithmus richtet sich wie bei jjwt nach der Secret-Länge (32–47 Bytes → HS256, 48–63 → HS384, ab 64 → HS512). **Mit demselben `JWT_SECRET` bleiben eingeloggte Nutzer beim Umstieg eingeloggt.**

Die Black-Box-Suite in `backend/contracttest` prüft das über HTTP und lässt sich gegen beide Backends ausführen:

```bash
CONTRACT_BASE_URL=http://localhost:8080 CONTRACT_LEGACY=1 go test ./contracttest/ -count=1   # altes Spring-Backend
CONTRACT_BASE_URL=http://localhost:8080 go test ./contracttest/ -count=1                     # RevU
```

## Bewusste Änderungen

| Bereich | Vorher (OvervU) | Jetzt (RevU) |
|---|---|---|
| Nicht eingeloggt / Token abgelaufen | `403` ohne Body, Frontend zeigte endlos „An error occurred“ | `401` mit Fehler-Body; das Frontend meldet ab und leitet zum Login |
| Wöchentliche Habits | ließen sich täglich abhaken (Konsistenz dadurch schnell 100 %) | einmal pro Kalenderwoche (Mo–So), sonst `409 "Habit already completed this week"` |
| „Heute“ | Server-Zeitzone (Cloud Run = UTC): zwischen 0 und 2 Uhr landete ein Haken auf dem Vortag | `APP_TIMEZONE` (Standard `Europe/Berlin`) für Habits, Tagesstatistik und Monatsgrenzen; das Frontend nutzt lokale statt UTC-Daten |
| Ungültige Eingaben ohne passenden Handler | `500` (z. B. Monat 13, Wert zu groß für die Spalte) | `400` mit Meldung |
| Unbekannte Routen | `403` | `404` |

## Behobene Fehler

1. **Fitness: Übungen erschienen mehrfach.** Die JPA-Abfrage lud Übungen und Sätze in einem Join; jede Übung tauchte so oft auf, wie sie Sätze hatte, immer mit derselben ID. Deshalb änderte sich im UI „alles gleichzeitig“ (TODO aus der alten README). Jetzt erscheint jede Übung genau einmal.
2. **Sicherheitslücke:** `DELETE /api/habits/{id}/entries/{entryId}` und `DELETE /api/learning/{id}/sessions/{sessionId}` prüften nicht, ob der Eintrag zum Habit bzw. Projekt in der URL gehört. Mit einem eigenen Habit ließen sich so Einträge fremder Nutzer löschen. Das wird jetzt geprüft und mit `404` abgelehnt.
3. **Satznummern:** Nach dem Löschen eines Satzes wurden Nummern doppelt vergeben. Jetzt wird neu nummeriert (1..n), ebenso die Reihenfolge der Übungen.
4. **Dark Mode** wurde nur angewendet, solange die Profilseite offen war; nach einem Reload war die App hell. Jetzt wird das Theme vor dem ersten Rendern gesetzt, ohne Flackern.
5. **Feature-Toggles:** Abgeschaltete Bereiche verschwinden jetzt auch aus dem Dashboard, wie es die Profilseite verspricht, und sind nicht mehr per URL erreichbar.
6. **Volumen-Chart (Fitness)** zeigte die ältesten statt der letzten 10 Trainings.
7. **Learning:** Eine neue Session in einem noch nicht aufgeklappten Projekt führte dazu, dass beim Aufklappen nur diese eine Session angezeigt wurde.
8. **Atomarität:** Änderung und zugehöriger Activity-Eintrag werden in einer Transaktion geschrieben. Workout-Änderungen sperren das Workout gegen parallele Änderungen.
9. **Performance:** keine N+1-Abfragen mehr (Lernzeit pro Projekt, Habit-Konsistenz, 6-Monats-Übersicht, Workout-Liste).
10. **Timing:** Logins mit unbekannter E-Mail dauern so lange wie Logins mit falschem Passwort, damit sich nicht erraten lässt, welche E-Mails registriert sind.

## Repo-Hygiene

- Keine `.env`-Dateien mit Secrets mehr im Repo (es gibt nur `.env.example`), kein `run-log.txt`.
- `.gitattributes` erzwingt LF-Zeilenenden. Das alte Repo hatte gemischte CRLF/LF-Dateien.
- Der Compose-Stack funktioniert wieder: Vorher lauschte nginx auf 8080, Compose leitete aber auf 80 weiter, und `/api` wurde nicht durchgereicht.
- Die nie eingebundene `AnalyticsPage` ist entfallen; ihre Charts sind im Dashboard.

## Umstieg in Produktion (Checkliste)

1. `JWT_SECRET` des alten Backends aus Cloud Run bzw. Secret Manager übernehmen.
2. RevU-Backend als **neue Revision des bestehenden Backend-Service** deployen (siehe [DEPLOYMENT.md](DEPLOYMENT.md)). Dabei die alten `SPRING_*`-Variablen durch `DB_*` ersetzen. Das Backend startet nur, wenn die Datenbank auf Flyway-Stand V14 ist; das ist sie bereits.
3. Smoke-Test: `GET /healthz`, Login mit einem bestehenden Konto, Dashboard laden.
4. RevU-Frontend als neue Revision des bestehenden Frontend-Service deployen. Die URLs bleiben gleich, CORS passt also.
5. **Rollback** jederzeit möglich, da sich die Datenbank nicht ändert:
   `gcloud run services update-traffic <service> --to-revisions=<alte-revision>=100`
