# Decoupling Guide — API-backed reads

This document explains the feature flags and runtime behavior introduced to allow RentalCore to read products and customers from WarehouseCore via API rather than the shared database.

Feature flags
- `CABLE_SNAPSHOT_ENABLED` — existing flag for cable snapshot reads.
- `WAREHOUSE_PRODUCTS_ENABLED` — when `true`, `ProductRepository` will attempt to read products from WarehouseCore API first, falling back to the local DB on error.
- `WAREHOUSE_CUSTOMERS_ENABLED` — when `true`, `CustomerRepository` will attempt to read customers from WarehouseCore API first, falling back to the local DB on error. Writes remain local to RentalCore.

Environment variables
- `WAREHOUSECORE_BASE_URL` — base URL for WarehouseCore (e.g. `http://warehousecore:8082`).
- `WAREHOUSECORE_API_KEY` — API key to send with `X-API-Key` header for authenticated calls.
- `WAREHOUSE_PRODUCTS_ENABLED`, `WAREHOUSE_CUSTOMERS_ENABLED` — set to `true` to enable API-backed reads.

Behavioral notes
- Reads are best-effort: the code falls back to the local DB on any API error to avoid outages.
- Writes (create/update/delete) for customers remain local in RentalCore to avoid surprising cross-service side-effects.
- Product listing in API mode maps a minimal set of fields into the local `models.Product` structure. Complex filters should be handled by WarehouseCore when possible.

Testing
- Unit tests were added under `internal/repository` covering both DB and API-backed modes.

SSO and users/auth guidance
- Current approach: keep authentication and user records in RentalCore and expose a small SSO integration layer to WarehouseCore.
- Recommended SSO pattern:
  - Use JWTs signed by RentalCore with a shared HMAC or RSA key distributed via environment variables (e.g. `SSO_JWT_SECRET`; if unset, RentalCore falls back to `ENCRYPTION_KEY`).
  - WarehouseCore accepts and verifies JWTs on incoming requests and maps the token to a local session.
  - For delegated user queries (e.g., user details), WarehouseCore can call RentalCore's user API with `X-API-Key` or via mutual TLS.

Next steps
- Implement API-backed reads for users/auth (read-only) while keeping central auth in RentalCore. I can implement this next and add tests and docs for SSO.
