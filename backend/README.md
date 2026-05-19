# Backend

Go + Gin backend for the INDICO OTT Integration take-home.

## Run

```bash
cd backend
go test ./...
go run ./cmd/api
```

## Environment Variables

- `PORT` default `8080`
- `FRONTEND_BASE_URL` default `http://localhost:5173`
- `NETPLAY_BASE_URL` default `https://ctazh5lrhe.execute-api.ap-southeast-3.amazonaws.com/dev/api`
- `STORE_PATH` default `data/activations.json`
- `HTTP_TIMEOUT_SECONDS` default `10`

## Persistence Choice

JSON file storage. It is simple, survives restart, and fits the assignment without a database.

## Assumptions

- `NETPLAY` is the real provider integration; `NETFLIX` is included as a demo provider.
- Frontend activation route uses `/activation/{code}`.
- Local activation code is the frontend-facing token.

## Known Limitations

- No retry policy for provider calls.
- One provider is implemented, but the provider registry is generic.

## AI Usage

AI was used to accelerate scaffolding, refactoring, and test drafting. Final code was reviewed and adjusted manually.
