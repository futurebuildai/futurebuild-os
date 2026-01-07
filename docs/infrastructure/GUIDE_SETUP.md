# Local Development Setup Guide

This guide will help you set up the infrastructure required to run FutureBuild locally.

## 1. Install Dependencies

You need to install Docker and a few CLI tools:

### Docker & Docker Compose
- **Ubuntu/Debian:**
  ```bash
  sudo apt update
  sudo apt install docker.io docker-compose -y
  sudo usermod -aG docker $USER
  # Log out and back in for group changes to take effect
  ```
- **macOS/Windows:**
  Install [Docker Desktop](https://www.docker.com/products/docker-desktop/).

### Golang Migrate (For DB Schema)
- **Linux:**
  ```bash
  curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
  sudo mv migrate /usr/local/bin/
  ```

---

## 2. Start Infrastructure

Use Docker Compose to start the PostgreSQL database and Redis:

```bash
docker-compose up -d db redis
```

Verify that the services are running:
```bash
docker-compose ps
```

---

## 3. Configure Environment

Create a `.env` file from the example:

```bash
cp .env.example .env
```

The `DATABASE_URL` in `.env` should point to the mapped port in `docker-compose.yml` (usually `5433` for local host connection):
```
DATABASE_URL=postgres://fb_user:fb_pass@localhost:5433/futurebuild?sslmode=disable
```

---

## 4. Run Migrations

Before seeding data, you must apply the database schema:

```bash
make migrate-up
```

---

## 5. Seed WBS Data

Now you can run the seeding script we just created:

```bash
DATABASE_URL=postgres://fb_user:fb_pass@localhost:5433/futurebuild?sslmode=disable go run scripts/seed_wbs.go
```

---

## Troubleshooting

- **Connection Refused:** Ensure `docker-compose up -d db` was successful and use port `5433` if you are running the Go script outside of Docker.
- **Permission Denied:** If running `docker` commands fails, try prefixing with `sudo` or ensure your user is in the `docker` group.
