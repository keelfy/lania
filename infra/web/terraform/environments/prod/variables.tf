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
  description = "MariaDB host. Defaults to the local mariadb container; shared with the Minecraft plugins and read by the shell service"
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

variable "luckperms_user_permissions_table_name" {
  description = "LuckPerms user permissions table in the main database"
  type        = string
  default     = "luckperms_user_permissions"
}

variable "plan_users_table_name" {
  description = "Plan users table in the main database"
  type        = string
  default     = "plan_users"
}

variable "plan_sessions_table_name" {
  description = "Plan sessions table in the main database"
  type        = string
  default     = "plan_sessions"
}

variable "flectone_player_table_name" {
  description = "FlectonePulse player table in the main database"
  type        = string
  default     = "player"
}

variable "rcon_address" {
  description = "host:port of the Minecraft server RCON reachable from the shell container. Empty disables live role sync and fails whitelist and prefix changes"
  type        = string
  default     = ""
}

variable "rcon_password" {
  description = "Password for the Minecraft server RCON"
  type        = string
  sensitive   = true
  default     = ""
}

variable "shell_token" {
  description = "Shared secret the API sends to the shell service"
  type        = string
  sensitive   = true
}

variable "max_profiles_per_user" {
  type    = number
  default = 2
}

variable "default_name_color_id" {
  type    = string
  default = "2628bf9d-5b7c-438b-900a-67753261a823"
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

variable "grafana_admin_password" {
  description = "Password of the Grafana `admin` user at grafana.<domain>"
  type        = string
  sensitive   = true
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

# --- S3 (admin image uploads: glyth previews, season screenshots and previews) ---

variable "s3_bucket" {
  description = "Bucket the API uploads admin images into. imgproxy reads from it with its own credentials"
  type        = string
  default     = "lania-web-134312503254-eu-central-1-an"
}

variable "s3_region" {
  type    = string
  default = "eu-central-1"
}

variable "s3_endpoint" {
  description = "Object storage endpoint. Empty uses the AWS default endpoint for s3_region"
  type        = string
  default     = ""
}

variable "s3_force_path_style" {
  description = "Most S3-compatible providers other than AWS need bucket-in-path addressing"
  type        = bool
  default     = false
}

variable "aws_access_key_id" {
  description = "Credentials for the API's own s3:PutObject-only IAM user, distinct from imgproxy's read credentials"
  type        = string
  sensitive   = true
}

variable "aws_secret_access_key" {
  type      = string
  sensitive = true
}
