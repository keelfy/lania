# Lania Web — Deployment

Deploys the website stack (frontend, API, auth, databases) to the shared Hetzner
Robot dedicated server, alongside other projects already running there. No
service in this stack publishes ports directly to the host — routing and TLS
are handled by the shared `lania-edge` Traefik stack (see `infra/edge/`),
which must be deployed once, first, before this one. Container names are
prefixed `lania-` to avoid clashing with anything else on the box.

## Architecture

- **front** — Next.js (`ghcr.io/lania-smp/frontend`), routed at `www.lania.network`
  (and `lania.network` redirects to it) via Traefik labels.
- **admin** — standalone admin service at `admin.lania.network`, using existing
  Kratos accounts. Set Terraform `admin_identity_ids` before granting access;
  an empty list denies everyone. See [admin setup](../../apps/admin/README.md).
- **api** — Go backend (`ghcr.io/lania-smp/backend`), routed at `api.lania.network`.
- **shell** — Go gRPC service (`ghcr.io/lania-smp/shell`), the API's only way to
  reach the Minecraft server and its plugin data (Plan, Flectone, LuckPerms,
  whitelist). Internal only.
- **mariadb** — shared instance: API database `lania` plus the Minecraft plugin
  databases (`flectone`, `plan`) that only `shell` reads. Internal only.
- **redis** — API cache. Internal only.
- **postgres** + **kratos** — Ory Kratos auth. Public API routed at
  `accounts.lania.network`; the Kratos admin API (4434) has no route and is
  not reachable from outside the `lania-web-net` network.

`front`, `api`, `admin`, and `kratos` each join two networks: `lania-web-net` (talk to
the databases) and the external `edge` network (so the shared Traefik can
reach them). `mariadb`, `redis`, `postgres`, `kratos-migrate` stay on
`lania-web-net` only.

> **Shared `edge` network caveat**: any other project's containers that also
> join the `edge` network can reach `lania-front`, `lania-api`, and
> `lania-kratos` directly by container name/port, bypassing Traefik's routing
> (plain Docker bridge networks don't do inter-container ACLs). This is a
> single-owner box, so the practical risk is low, but don't put anything on
> `edge` that shouldn't be reachable by every other site sharing this server.

Elasticsearch and imgproxy env vars exist in the API's config package but are
not wired into any service today (see `apps/api/cmd/wire_gen.go`), so they're
left out of this stack. Add them later if a feature actually needs them.

> **Plugin databases**: `flectone`, `plan`, LuckPerms and whitelist tables are
> written by the Minecraft plugins. Only `shell` touches them; the API talks to
> `shell` over gRPC (`SHELL_ADDRESS`, `SHELL_TOKEN`).

## One-time setup

0. **Deploy the edge stack first** — see `infra/edge/README.md`. This stack's
   `docker-compose.prod.yml` declares `edge` as an `external` network and will
   fail to come up if it doesn't exist yet.
1. **Cloudflare**: create an API token scoped to `Zone:DNS:Edit` for the
   `lania.network` zone, and note the zone ID.
2. **SSH**: make sure your public key is authorized on the dedicated server
   for the user you'll deploy as (`root` by default).
3. **Terraform**:
   ```bash
   cd infra/web/terraform/environments/prod
   cp terraform.tfvars.example terraform.tfvars   # fill in real values, never commit this
   mise run web-terraform-prod-init
   mise run web-terraform-prod-plan
   mise run web-terraform-prod-apply
   ```
   This will:
   - create Cloudflare A records for `lania.network`, `www`, `api`, `accounts`, `admin`
   - install Docker on the server if missing
   - render a real `.env` from your `terraform.tfvars` (kept only in
     `generated/.env`, gitignored, never committed)
   - upload the compose stack and bring it up

4. **GitHub Actions secrets** (repo settings → Secrets and variables → Actions):
   - `DEPLOY_SSH_HOST` — server IP
   - `DEPLOY_SSH_USER` — deploy SSH user
   - `DEPLOY_SSH_KEY` — private key matching an authorized key on the server
   - Repo **variables** (not secrets): `NEXT_PUBLIC_PREREGISTRATION`,
     `NEXT_PUBLIC_ACTIVE_SEASON_ID` — the rest of the `NEXT_PUBLIC_*` build args
     are hardcoded to the production domain in the workflow.

## Ongoing deploys

Push to `main`. `.github/workflows/deploy.yml`:
1. detects whether `apps/api/**` and/or `apps/front/**` changed,
2. builds/pushes only the images that changed to GHCR (`:latest` and `:<sha>`),
3. SSHes into the server and runs `docker compose pull && up -d`.

This only rolls new images — it does not touch compose/Kratos config or
Traefik routing labels. Any change to those (or to secrets) goes through
`terraform apply` again, which re-uploads the compose directory and restarts
affected containers.

**Rollback**: re-run `terraform apply` with `image_tag` pinned to a previous
`:<sha>` tag, or manually `docker compose` up a specific tag on the server.

## Manual server access

Kratos admin API and the databases are not reachable from outside the Docker
network. To inspect them, SSH into the server first:

```bash
ssh root@<server-ip>
docker exec -it lania-mariadb mysql -u root -p
docker exec -it lania-postgres psql -U postgres -d kratos
docker exec -it lania-kratos wget -qO- http://localhost:4434/admin/identities
```

## Adding a second site to this server

Don't add another reverse proxy. Give the new site its own compose stack (its
own `docker-compose.yml`, its own internal network for its databases), and:

- join the external `edge` network on its public-facing service(s) only,
- never publish host ports 80/443 from it,
- add `traefik.*` labels the same way `front`/`api`/`kratos` do here, with a
  `Host()` rule for its own domain,
- give it its own Terraform under `infra/<site>/` for its own DNS records —
  the `edge` Traefik stack itself needs no changes to onboard it.

## Firewall

Only ports 80 and 443 need to be open to the public internet, and those are
owned by the `lania-edge` stack, not this one. Nothing here needs its own
firewall rule.
