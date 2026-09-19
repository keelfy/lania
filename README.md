# Lania Monorepo

This repository now tracks all services and apps in a single Git repository.

## Projects

- `NetworkJoinMessages`
- `Scythe`
- `VelocityWhitelist`
- `front` (Next.js frontend, Bun-based workflow)
- `monolith/apps/api` (Go backend, `mise` task runner)
- `admin` ([standalone admin service](apps/admin/README.md), Go + server-rendered HTML)
- `shell` (Go gRPC service, the API's gateway to the Minecraft server and plugins)
- `proto` (protobuf contracts, `mise run proto-generate`)
- `season-extractor`
- `infra/terraform` (Terraform infrastructure scripts)

## Tooling

- Use `mise` to manage tool versions and run tasks.
- Use Bun for frontend dependency management and scripts.

## Quick Start

```bash
mise install
mise run front-install
mise run api-deps
```

## Local Test Environment

Runs the API and front against a local database without Docker. Everything lives in `.local-dev/` (ignored by git).

```bash
mise run dev-setup   # once: downloads MariaDB and Kratos, builds helper tools
mise run dev-up      # starts MariaDB, Redis, Kratos, API, front; migrates and seeds on first run
mise run dev-down
```

- Front: http://localhost:3000, API: http://localhost:8080/v1, Kratos: http://localhost:4433
- Login: `test@lania.local` / `lania-local-pass` (owns the seeded profiles `Keelfy` and `Alex`)
- `mise run dev-seed` reloads the test profiles, `mise run dev-reset` wipes all local data
- Config lives in `dev/dev.env`, test data in `dev/seed.sql`, logs in `.local-dev/logs`
- Redis and imgproxy are small stand-ins (`dev/tools`). The Minecraft shell is not started, so every player shows as offline.
- OAuth providers, payments and Donation Alerts are not configured.

## Common Commands

```bash
# Frontend
mise run front-dev
mise run front-build

# Go API
mise run api-run
mise run api-test
mise run api-build
```
