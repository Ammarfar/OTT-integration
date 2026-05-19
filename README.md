# OTT Integration Demo

Simple INDICO OTT integration take-home with:

- `backend/` Go + Gin API
- `frontend/` Vite + React activation page

## Quick Start

Backend:

```bash
cd backend
go test ./...
go run ./cmd/api
```

Frontend:

```bash
cd frontend
npm install
npm run dev
```

## Main Env Vars

- Backend `PORT`
- Backend `FRONTEND_BASE_URL`
- Backend `NETPLAY_BASE_URL`
- Backend `STORE_PATH`
- Backend `HTTP_TIMEOUT_SECONDS`
- Frontend `VITE_API_BASE_URL`

## Notes

- `NETPLAY` is the real provider integration.
- `NETFLIX` is included as a demo provider.
- Local activation code is the frontend-facing token.
- State is stored in a JSON file.

## Docs

- Backend architecture: [`backend/ARCHITECTURE.md`](./backend/ARCHITECTURE.md)
- Backend setup: [`backend/README.md`](./backend/README.md)
- Frontend setup: [`frontend/README.md`](./frontend/README.md)
- AI usage: [`backend/AI-USAGE.md`](./backend/AI-USAGE.md)
