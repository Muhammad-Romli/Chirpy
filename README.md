# Chirpy

A microblogging backend, similar to Twitter, written in Go and backed by PostgreSQL.

Users can sign up, log in, post short messages called chirps, and read or delete them through a JSON HTTP API. The server also serves a static page and exposes simple admin endpoints.

> **Status: Completed.**

## Features

- User registration and login with hashed passwords
- JWT access tokens and revocable refresh tokens
- Create, list, fetch and delete chirps, with optional filtering by author
- Update the current user's email and password
- Toy webhook endpoint for premium upgrades
- Static file server for the front page and assets
- Request metrics and a reset endpoint for development
- Unit tests for the handlers and the auth package

## Requirements

- [Go](https://go.dev/dl/)
- [PostgreSQL](https://www.postgresql.org/download/)

For development only:

- [goose](https://github.com/pressly/goose) for database migrations
- [sqlc](https://sqlc.dev/) to generate Go code from SQL queries
- `psql` for inspecting the database

## Installation

```bash
git clone https://github.com/muhammad-romli/Chirpy.git
cd Chirpy
go build -o chirpy
```

## Setup

### 1. Create the database

```bash
psql -U postgres -c "CREATE DATABASE chirpy;"
```

Your connection string will look like this (adjust user, password, host and port):

```
postgres://USER:PASSWORD@localhost:5432/chirpy?sslmode=disable
```

### 2. Run the migrations

From the project root:

```bash
goose -dir sql/schema postgres "postgres://USER:PASSWORD@localhost:5432/chirpy?sslmode=disable" up
```

This creates the `users`, `chirps` and `refresh_tokens` tables.

### 3. Create the environment file

Create a `.env` file in the project root:

```
DB_URL="postgres://USER:PASSWORD@localhost:5432/chirpy?sslmode=disable"
PLATFORM=dev
JWT_SECRET=<generated secret>
POLKA_KEY=f271c81ff7084ee5b99a5091b42d486e
```

- `DB_URL`: the PostgreSQL connection string
- `PLATFORM`: set to `dev` to allow the reset endpoint
- `JWT_SECRET`: the secret used to sign access tokens
- `POLKA_KEY`: the API key for the toy Polka webhook. It is not a real service and the key is always the same, so use the value above as-is.

Generate `JWT_SECRET` with:

```bash
openssl rand -base64 64
```

Keep `.env` out of version control by adding it to `.gitignore`.

### 4. Run the server

```bash
./chirpy
```

## API

### Public

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/healthz` | Health check |
| `POST` | `/api/users` | Create a user |
| `GET` | `/api/users` | List all users |
| `GET` | `/api/users/{userID}` | Get one user |
| `POST` | `/api/login` | Log in and receive an access token and refresh token |
| `POST` | `/api/refresh` | Get a new access token using a refresh token |
| `POST` | `/api/revoke` | Revoke a refresh token |
| `GET` | `/api/chirps` | List all chirps (optional `?author_id=<userID>` to filter) |
| `GET` | `/api/chirps/{chirpID}` | Get one chirp |

### Authenticated (`Authorization: Bearer <token>`)

| Method | Endpoint | Description |
|---|---|---|
| `PUT` | `/api/users` | Update the current user's email and password |
| `POST` | `/api/chirps` | Create a chirp |
| `DELETE` | `/api/chirps/{chirpID}` | Delete one of your own chirps |

### Webhook

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/polka/webhooks` | Receive premium purchase events (`Authorization: ApiKey <POLKA_KEY>`) |

Polka is a toy payment service made for this project. It is not real, and its API key is always the same.

### Admin

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/admin/metrics` | Show how many times the front page was visited |
| `POST` | `/admin/reset` | Reset the visit counter and delete all users (`PLATFORM=dev` only) |

The front page and assets are served under `/app/`.

> Warning: `/admin/reset` wipes data. Don't enable it against a database you care about.

### Example

```bash
# Create a user
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"email": "alice@example.com", "password": "secret"}'

# Log in
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email": "alice@example.com", "password": "secret"}'

# Post a chirp
curl -X POST http://localhost:8080/api/chirps \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{"body": "Hello, Chirpy!"}'

# List chirps
curl http://localhost:8080/api/chirps
```

## Development

Project layout:

```
.
├── main.go              # entry point, routes and server setup
├── handlers.go          # HTTP handlers
├── handlers_test.go     # handler tests
├── index.html           # static front page
├── assets/              # static assets
├── internal/
│   ├── auth/            # password hashing, JWT and token helpers
│   └── database/        # sqlc-generated code (do not edit by hand)
├── sql/
│   ├── schema/          # goose migrations
│   └── queries/         # sqlc query definitions
└── sqlc.yaml
```

After changing anything in `sql/queries/` or `sql/schema/`, regenerate the database code:

```bash
sqlc generate
```

Run the tests:

```bash
go test ./...
```

## Built with

- [lib/pq](https://github.com/lib/pq), PostgreSQL driver
- [google/uuid](https://github.com/google/uuid), UUID generation
- [godotenv](https://github.com/joho/godotenv), `.env` loading
- [golang-jwt](https://github.com/golang-jwt/jwt), JWT handling
- [sqlc](https://sqlc.dev/) and [goose](https://github.com/pressly/goose)

## License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.
