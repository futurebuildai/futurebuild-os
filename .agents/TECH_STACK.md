# Tech Stack Configuration

This file is the single source of truth for the project's technology choices. All workflows and skills reference this file instead of hardcoding tech stacks.

**System:** FutureBuild Brain (System of Connection)

---

## Backend

- **Language:** Go 1.25
- **Framework:** chi router (go-chi/chi/v5)
- **ORM / Data Access:** pgx/v5 (jackc/pgx) — raw SQL with pgx types, no ORM

## Frontend

- **Framework:** Vite + Lit (vanilla Web Components)
- **Bundler:** Vite
- **Styling:** Vanilla CSS (CSS custom properties for design tokens)
- **Language:** TypeScript (strict mode)

## Mobile (if applicable)

- **Framework:** N/A — FB-Brain is web-only

## Database

- **Primary:** PostgreSQL 16+
- **Cache:** Redis 7
- **Search (if applicable):** None (deferred)
- **Message Queue (if applicable):** None (deferred)

## API

- **Style:** REST + MCP (Model Context Protocol)
- **Spec Format:** OpenAPI 3.1
- **Auth Model:** OIDC Provider — FB-Brain IS the identity provider for the FutureBuild ecosystem. Issues JWTs (access + refresh tokens) via standard OIDC flows. All downstream systems (including FB-OS) delegate authentication to FB-Brain.

## Infrastructure & Deployment

- **CI/CD:** GitHub Actions
- **Hosting:** TBD (to be determined during Architecture Spec stage)
- **Containerization:** Docker (multi-stage builds)
- **Orchestration:** Docker Compose (local dev), TBD for production

## Developer Tooling

- **IDE:** GoLand (JetBrains)
- **Package Manager:** Go Modules (backend), npm (frontend)
- **Linter:** golangci-lint (backend), ESLint (frontend)
- **Formatter:** gofmt (backend), Prettier (frontend)
- **Task Runner:** Makefile

## Testing

- **Unit Test Framework:** go test (backend), Vitest (frontend)
- **Integration Test Framework:** Testcontainers-go
- **E2E Test Framework:** Playwright
- **Coverage Tool:** go cover (backend), c8 (frontend)

## Observability

- **Logging:** structured JSON via slog (Go stdlib)
- **Tracing:** OpenTelemetry
- **Metrics:** Prometheus
- **Error Tracking:** Sentry

## AI Service Layer

- **Primary Provider:** Anthropic — Claude Opus 4.6 (reasoning, orchestration, tool use) and Claude Sonnet 4.5 (high-throughput, lower-latency tasks)
- **Maestro AI Co-pilot:** Built on Claude with native tool use for probabilistic intent parsing and A2A webhook emission
- **Embeddings:** Anthropic if available; fall back to open-source (e.g., nomic-embed, BGE) only if needed for pgvector ingestion
- **Open Source Policy:** Open-source models permitted ONLY for edge/niche use cases where Anthropic does not offer a suitable capability (e.g., on-device inference, domain-specific fine-tunes). All core intelligence routes through Anthropic.

## Constraints & Preferences

- **Currency (Composite Currency Pattern):** ALL monetary values stored as the **Composite Currency Pattern**: `amount_cents BIGINT` paired with `currency_code VARCHAR(3) DEFAULT 'USD'`. Supported currencies: USD (United States Dollar) and CAD (Canadian Dollar). No floating-point currency. Display formatting is a frontend concern only. Cross-currency arithmetic is **forbidden** — values with different `currency_code` MUST NOT be summed, compared, or subtracted. Aggregations must group by `currency_code`.
- **Currency Enforcement (CI Hard Gates):**
  - **SQL Migration Linter (CRITICAL — hard fail in GitHub Actions):** Script scans `migrations/*.sql` for: (1) forbidden types (`DECIMAL`, `NUMERIC`, `REAL`, `DOUBLE PRECISION`, `MONEY`, `FLOAT`) on columns matching monetary patterns (`cost`, `price`, `amount`, `total`, `budget`, `cents`, `fee`, `payment`, `invoice`, `balance`, `revenue`, `expense`) — any match fails CI; (2) any `amount_cents` column (or column ending in `_cents` matching monetary patterns) that lacks a corresponding `currency_code` column in the same CREATE TABLE statement — fails CI. No exemptions.
  - **Go Struct Naming Convention:** All monetary fields MUST end in `Cents` (e.g., `TotalActualCostCents`, `EstimatedPriceCents`) with a companion field ending in `CurrencyCode` (e.g., `TotalActualCostCurrencyCode`). `golangci-lint` custom rule flags `float32`/`float64` fields on structs containing monetary field names.
  - **TypeScript ESLint Rule:** Custom rule flags `number` type annotations on properties matching monetary name patterns unless the property name ends in `Cents`. Properties ending in `Cents` must have a sibling property ending in `CurrencyCode`. Enforced via `eslint-plugin-fb` in frontend lint config.
- **Numerical Typography:** JetBrains Mono for all numerical data fields in the UI.
- **AI-First Principle:** Anthropic Claude is the default AI provider across the ecosystem. Do not introduce Google Vertex, OpenAI, or other commercial LLM providers unless Anthropic cannot serve the use case. Open-source models are acceptable for edge cases only.
- **Identity Role:** FB-Brain owns user identity, org management, and authentication for the entire ecosystem. No other system stores credentials or manages sessions independently.
- **Polyrepo:** FB-Brain and FB-OS are separate repositories with separate deployment lifecycles.
- **API Contracts:** FB-Brain exposes OIDC-compliant endpoints (/.well-known/openid-configuration, /authorize, /token, /userinfo) consumed by FB-OS and future ecosystem services.
