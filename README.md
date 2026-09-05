# Sign-Out API

Go backend for Sign-Out. This scaffold starts with **authentication**.

## Layout

```
cmd/api/                 # process entrypoint
internal/
  config/                # env-based configuration
  domain/                # entities + domain errors
  handler/auth/          # HTTP handlers
  middleware/            # auth, recover, request id
  repository/postgres/   # Postgres persistence
  service/auth/          # business logic + JWT
  server/                # router wiring
migrations/              # SQL migrations
pkg/response/            # JSON envelope helpers
pkg/validator/           # request validation helpers
```

## Quick start

```bash
cp .env.example .env
docker compose up -d
make run
```

API listens on `http://localhost:8080`.

## Auth endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/v1/auth/register` | — | Create account |
| `POST` | `/v1/auth/login` | — | Login |
| `POST` | `/v1/auth/refresh` | — | Rotate refresh token |
| `POST` | `/v1/auth/logout` | — | Revoke refresh token |
| `GET`  | `/v1/auth/me` | Bearer | Current user |
| `GET`  | `/healthz` | — | Health check |

### Register

```bash
curl -s http://localhost:8080/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@school.edu.ng","password":"password123","full_name":"Laetitia"}'
```

### Login

```bash
curl -s http://localhost:8080/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@school.edu.ng","password":"password123"}'
```

### Me

```bash
curl -s http://localhost:8080/v1/auth/me \
  -H "Authorization: Bearer <access_token>"
```

Responses use a standard envelope:

```json
{ "success": true, "data": { ... } }
```
