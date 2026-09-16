# Sign-Out API

Go backend for Sign-Out. Auth is backed by **MongoDB**.

## Layout

```
cmd/api/
internal/
  config/
  domain/
  handler/auth/
  middleware/
  repository/mongo/
  service/auth/
  server/
pkg/response/
pkg/validator/
```

## Quick start

```bash
cp .env.example .env
docker compose up -d
make run
```

API: `http://localhost:8080`

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
