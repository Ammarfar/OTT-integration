# Architecture Note

## Structure

- `internal/domain`: business entities and state
- `internal/dto/http`: HTTP request/response shapes
- `internal/dto/provider`: NETPLAY provider payloads
- `internal/provider`: provider interface and registry
- `internal/provider/netplay`: NETPLAY HTTP client
- `internal/service`: subscribe, activate, status business logic
- `internal/store`: JSON file persistence
- `internal/handler`: Gin routes and DTO mapping

## Flow

1. `POST /api/subscribe` validates input and calls provider subscribe.
2. Service stores provider token and local activation record.
3. Backend returns activation link and SMS-style message.
4. `POST /api/activate` resolves the stored record and calls provider activate.
5. `GET /api/subscription-status` refreshes status from provider.
6. `GET /api/providers` returns registered providers.

## Trade-offs

- JSON file instead of database to keep the solution lightweight.
- Thin layered architecture instead of full DDD to keep scope aligned with the assignment.
- Provider registry is small and generic, enough for future providers without extra framework code.
