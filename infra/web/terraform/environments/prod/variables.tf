# --- Server (existing Hetzner Robot dedicated server, not managed by Terraform) ---

variable "server_ipv4" {
  description = "Public IPv4 of the existing Hetzner Robot dedicated server this stack deploys to"
  type        = string
}

variable "ssh_user" {
  description = "SSH user with docker access on the dedicated server"
  type        = string
  default     = "root"
}

variable "ssh_private_key_path" {
  description = "Path to the SSH private key used to connect to the server"
  type        = string
  default     = "~/.ssh/id_ed25519"
}

variable "deploy_path" {
  description = "Directory on the server this stack is deployed into"
  type        = string
  default     = "/opt/lania-web"
}

variable "image_tag" {
  description = "Tag of the ghcr.io/lania-smp/{backend,frontend} images to deploy"
  type        = string
  default     = "latest"
}

# --- DNS (Cloudflare) ---

variable "cloudflare_api_token" {
  description = "Cloudflare API token (Zone:DNS:Edit) for the lania.network zone"
  type        = string
  sensitive   = true
}

variable "cloudflare_zone_id" {
  description = "Cloudflare zone ID for lania.network"
  type        = string
}

variable "domain" {
  description = "Root domain"
  type        = string
  default     = "lania.network"
}

variable "cloudflare_proxied" {
  description = "Whether DNS records are proxied through Cloudflare (orange-cloud) or DNS-only"
  type        = bool
  default     = false
}

# --- Secrets rendered into the deployed .env (never committed, only via terraform.tfvars) ---

variable "api_key" {
  type      = string
  sensitive = true
}

variable "jwt_secret" {
  type      = string
  sensitive = true
}

variable "database_password" {
  description = "Root password for the MariaDB instance backing the API"
  type        = string
  sensitive   = true
}

variable "database_host" {
  description = "MariaDB host. Defaults to the local mariadb container; override if the flectone/plan databases live on the Minecraft network's own MariaDB instance"
  type        = string
  default     = "mariadb"
}

variable "database_port" {
  type    = string
  default = "3306"
}

variable "database_user" {
  type    = string
  default = "root"
}

variable "database_name" {
  type    = string
  default = "lania"
}

variable "database_flectone_name" {
  type    = string
  default = "flectone"
}

variable "database_plan_name" {
  type    = string
  default = "plan"
}

variable "active_season_id" {
  type    = string
  default = "00000000-0000-0000-0000-000000000000"
}

variable "max_profiles_per_user" {
  type    = number
  default = 2
}

variable "default_name_color_id" {
  type    = string
  default = "00000000-0000-0000-0000-000000000000"
}

variable "preregistration" {
  type    = bool
  default = false
}

variable "freekassa_base_payment_url" {
  type    = string
  default = ""
}

variable "freekassa_merchant_id" {
  type    = string
  default = ""
}

variable "freekassa_merchant_password_1" {
  type      = string
  default   = ""
  sensitive = true
}

variable "freekassa_merchant_password_2" {
  type      = string
  default   = ""
  sensitive = true
}

variable "donation_alerts_client_id" {
  type    = string
  default = ""
}

variable "donation_alerts_client_secret" {
  type      = string
  default   = ""
  sensitive = true
}

variable "donation_alerts_socket_connection_token" {
  type      = string
  default   = ""
  sensitive = true
}

variable "donation_alerts_user_id" {
  type    = string
  default = ""
}

variable "twitch_client_id" {
  type    = string
  default = ""
}

variable "twitch_client_secret" {
  type      = string
  default   = ""
  sensitive = true
}

variable "support_email" {
  type    = string
  default = "contact@lania.network"
}

variable "telegram_channel_url" {
  type    = string
  default = "https://t.me/laniamc"
}

variable "discord_server_url" {
  type    = string
  default = "https://discord.gg/laniamc"
}

variable "mariadb_root_password" {
  description = "Same value as database_password, used for the MariaDB container's own bootstrap"
  type        = string
  sensitive   = true
}

variable "postgres_user" {
  type    = string
  default = "postgres"
}

variable "postgres_password" {
  type      = string
  sensitive = true
}

variable "postgres_db" {
  type    = string
  default = "kratos"
}

variable "kratos_secrets_cookie" {
  type      = string
  sensitive = true
}

variable "kratos_secrets_cipher" {
  description = "Must be exactly 32 bytes"
  type        = string
  sensitive   = true
}

variable "courier_smtp_connection_uri" {
  description = "SMTP URI Kratos uses to send verification/recovery emails"
  type        = string
  sensitive   = true
}

variable "oidc_google_client_id" {
  type    = string
  default = ""
}

variable "oidc_google_client_secret" {
  type      = string
  default   = ""
  sensitive = true
}

variable "oidc_discord_client_id" {
  type    = string
  default = ""
}

variable "oidc_discord_client_secret" {
  type      = string
  default   = ""
  sensitive = true
}

variable "oidc_yandex_client_id" {
  type    = string
  default = ""
}

variable "oidc_yandex_client_secret" {
  type      = string
  default   = ""
  sensitive = true
}

variable "oidc_twitch_client_id" {
  type    = string
  default = ""
}

variable "oidc_twitch_client_secret" {
  type      = string
  default   = ""
  sensitive = true
}
