# Lania Monorepo

This repository now tracks all services and apps in a single Git repository.

## Projects

- `NetworkJoinMessages`
- `Scythe`
- `VelocityWhitelist`
- `front` (Next.js frontend, Bun-based workflow)
- `monolith/apps/api` (Go backend, `mise` task runner)
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
