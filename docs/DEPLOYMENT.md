# Deployment auf Google Cloud Run

Frontend und Backend laufen als zwei Cloud-Run-Services, die Datenbank in Cloud SQL (PostgreSQL). Die Beispiele nutzen die Namen und die Region des bisherigen OvervU-Deployments. Bitte an dein Projekt anpassen.

```bash
PROJECT=dein-gcp-projekt
REGION=europe-west3
INSTANCE=$PROJECT:$REGION:deine-cloudsql-instanz     # "Instance connection name"
BACKEND=overvu-backend                               # bestehender Service-Name
FRONTEND=overvu-frontend
```

## 1. Secrets (einmalig)

Das JWT-Secret **muss dasselbe wie beim alten Backend** sein, sonst werden alle Nutzer einmal ausgeloggt.

```bash
printf '%s' 'DAS-BISHERIGE-JWT-SECRET' | gcloud secrets create jwt-secret --data-file=- --project $PROJECT
printf '%s' 'DB-PASSWORT'              | gcloud secrets create db-password --data-file=- --project $PROJECT
```

Existieren die Secrets schon, stattdessen `gcloud secrets versions add …` verwenden. Das Service-Konto von Cloud Run braucht die Rolle `roles/secretmanager.secretAccessor`.

## 2. Backend

Cloud Run stellt Cloud SQL als Unix-Socket unter `/cloudsql/<instanz>` bereit. Den Pfad nutzt das Backend über `DB_HOST`.

```bash
gcloud run deploy $BACKEND \
  --project $PROJECT --region $REGION \
  --source backend \
  --add-cloudsql-instances $INSTANCE \
  --set-env-vars "DB_HOST=/cloudsql/$INSTANCE,DB_NAME=overvu,DB_USER=overvu,APP_TIMEZONE=Europe/Berlin,CORS_ALLOWED_ORIGINS=https://$FRONTEND-163257944718.$REGION.run.app" \
  --set-secrets "JWT_SECRET=jwt-secret:latest,DB_PASSWORD=db-password:latest" \
  --allow-unauthenticated
```

- `--set-env-vars` ersetzt alle bisherigen Variablen, also auch die alten `SPRING_DATASOURCE_*`.
- Die CORS-Origin muss exakt der Frontend-URL entsprechen (ohne Slash am Ende). Mehrere Origins trennst du mit Komma; dafür in `gcloud` die Syntax `^;^KEY=a,b;KEY2=c` verwenden.
- Prüfen: `curl https://<backend-url>/healthz` → `{"status":"UP"}`

## 3. Frontend

Die API-URL wird beim Build in das Frontend eingebacken (`VITE_API_BASE_URL`). Deshalb läuft der Build über Cloud Build mit `frontend/cloudbuild.yaml`:

```bash
API_URL=https://<backend-url>/api
IMAGE=$REGION-docker.pkg.dev/$PROJECT/cloud-run-source-deploy/$FRONTEND:$(git rev-parse --short HEAD)

gcloud builds submit frontend --project $PROJECT \
  --config frontend/cloudbuild.yaml \
  --substitutions "_API_URL=$API_URL,_IMAGE=$IMAGE"

gcloud run deploy $FRONTEND --project $PROJECT --region $REGION \
  --image $IMAGE --allow-unauthenticated
```

nginx lauscht auf dem von Cloud Run vorgegebenen `PORT`.

## 4. Datenbank-Migrationen

Die Produktionsdatenbank ist bereits auf V14, für den Umstieg ist also nichts zu tun. Für künftige Migrationen (`V15__…`) Flyway über den Cloud SQL Auth Proxy ausführen:

```bash
cloud-sql-proxy $INSTANCE --port 5433 &
docker run --rm --network host -v "$PWD/backend/migrations:/flyway/sql:ro" flyway/flyway:11 \
  -url=jdbc:postgresql://localhost:5433/overvu -user=overvu -password="$DB_PASSWORD" migrate
```

Erst migrieren, dann das Backend mit angehobenem `SCHEMA_VERSION` deployen.

## 5. Rollback

```bash
gcloud run revisions list --service $BACKEND --region $REGION
gcloud run services update-traffic $BACKEND --region $REGION --to-revisions=<alte-revision>=100
```

Da sich das Datenbankschema nicht ändert, kann jederzeit zwischen alter und neuer Revision gewechselt werden.
