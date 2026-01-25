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

## Agent Workflow (Prism Protocol)

This repository uses the **Prism Protocol** with a hybrid Antigravity + Claude Code workflow.

### Setup
1. **Install Claude Code**: `npm install -g @anthropic-ai/claude-code`
2. **Authenticate**: Run `claude login` in your terminal
3. **Configure Project**: Ensure a `CLAUDE.md` exists in your root

### The Execution Loop
1. **Plan (Antigravity)**: Ask DevTeam to "Prepare Step X". It generates a **Context Prompt**.
2. **Execute (Terminal)**: Copy the Context Prompt, run `claude -p "[Paste Context Prompt]"`
3. **Audit (Antigravity)**: Paste terminal output back, run `/CTO` for Triple Review

### Available Skills
| Skill | Purpose |
|-------|---------|
| `/product` | Discovery & Definition - Creates PRDs from ideas |
| `/devteam` | Engineering Orchestrator - Implements from specs |
| `/ops` | Operations - Reliability, security, incident response |
| `/software_engineer` | Context Prompt generation for Claude Code |

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

- [Backend Scope](specs/BACKEND_SCOPE.md)
- [Frontend Scope](specs/FRONTEND_SCOPE.md)
- [Roadmap](planning/ROADMAP.md)
- [System Prompt](agent/SYSTEM_PROMPT.md)

## License

Proprietary - All rights reserved.
