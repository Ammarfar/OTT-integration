# Frontend

Vite + React frontend for the NETPLAY activation flow.

## Run

```bash
cd frontend
npm install
npm run dev
```

## Environment

- `VITE_API_BASE_URL` default `http://localhost:8080`

## Routes

- `/activation/:code` activation landing page
- `/` fallback landing page

## Notes

- Mobile-first UI with a dark hero and activation card.
- Calls backend `POST /api/activate` and `GET /api/subscription-status`.
