# Lania Edge — Shared Reverse Proxy

Traefik, deployed once, independent of any single site. Owns host ports
80/443 on the shared Hetzner Robot dedicated server and creates the external
Docker network `edge` that every site stack (e.g. `infra/web`) joins to get
routed and get a Let's Encrypt certificate.

This exists so a second (or third) website can be deployed to the same server
later without a port conflict or a second reverse proxy — each site just
carries its own Traefik labels and joins `edge`; this stack itself doesn't
need to change to onboard a new site.

## Deploy

```bash
cd infra/edge/terraform/environments/prod
cp terraform.tfvars.example terraform.tfvars   # fill in real values, never commit this
mise run edge-terraform-prod-init
mise run edge-terraform-prod-plan
mise run edge-terraform-prod-apply
```

This installs Docker on the server if missing, uploads the compose stack, and
brings up Traefik. Deploy this **before** any site stack that depends on the
`edge` network.

## Notes

- The Traefik dashboard is disabled (`--api.dashboard=false`). If you want it,
  add a router with `traefik.http.middlewares.<name>.basicauth.users` — don't
  expose it unauthenticated on a shared box.
- Certificates are stored in the `edge-letsencrypt` volume via the ACME
  HTTP-01 challenge on port 80. If `lania.network`'s DNS is proxied through
  Cloudflare (orange-cloud), the challenge still works as long as port 80
  reaches this container; DNS-only is simpler and is this project's default.
- Redeploying this stack (`terraform apply` again) is safe — Traefik picks up
  routing labels from any running container automatically, no restart of
  site stacks needed.
