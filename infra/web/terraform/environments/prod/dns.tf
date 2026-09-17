locals {
  dns_subdomains = {
    root     = "@"
    www      = "www"
    api      = "api"
    accounts = "accounts"
  }
}

resource "cloudflare_record" "web" {
  for_each = local.dns_subdomains

  zone_id = var.cloudflare_zone_id
  name    = each.value
  type    = "A"
  content = var.server_ipv4
  proxied = var.cloudflare_proxied
  ttl     = var.cloudflare_proxied ? 1 : 300
}
