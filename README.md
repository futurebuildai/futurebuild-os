# FutureBuild

FutureBuild is an AI-powered construction project management platform that helps builders, general contractors, and project managers streamline their workflows.

## Features

- **AI Chat Interface**: Converse with an intelligent agent to manage schedules, invoices, and tasks.
- **Project Management**: Track schedules, budgets, and timelines with real-time updates.
- **Invoice Processing**: Upload and analyze invoices with AI-powered extraction.
- **Real-Time Collaboration**: Multi-user support with role-based access control.

## Getting Started

### Prerequisites

- Go 1.24+
- Node.js 20+
- PostgreSQL 15+
- Redis 7+

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/colton/futurebuild.git
   cd futurebuild
   ```

2. Install dependencies:
   ```bash
   go mod download
   npm --prefix frontend install
   ```

3. Set up environment variables:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. Run database migrations:
   ```bash
   make migrate-up
   ```

5. Start the application:
   ```bash
   # Backend
   go run cmd/api/main.go

   # Frontend
   npm --prefix frontend run dev
   ```

## Development

### Running Tests

```bash
# Unit tests (excludes integration tests)
make test

# Integration tests (requires running database)
make test-integration

# Frontend tests
npm --prefix frontend test
```

### Code Audit

```bash
make audit
```

## Architecture

- **Frontend**: Lit + TypeScript + Signals (Reactive State)
- **Backend**: Go 1.24 + chi router + pgx/v5
- **Database**: PostgreSQL + pgvector
- **AI**: Vertex AI / Gemini

## Documentation

- [Backend Scope](docs/BACKEND_SCOPE.md)
- [Frontend Scope](docs/FRONTEND_SCOPE.md)
- [Production Plan](docs/PRODUCTION_PLAN.md)

## License

Proprietary - All rights reserved.
