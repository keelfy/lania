variable "server_ipv4" {
  description = "Public IPv4 of the shared Hetzner Robot dedicated server"
  type        = string
}

variable "ssh_user" {
  type    = string
  default = "root"
}

variable "ssh_private_key_path" {
  type    = string
  default = "~/.ssh/id_ed25519"
}

variable "deploy_path" {
  description = "Directory on the server this stack is deployed into"
  type        = string
  default     = "/opt/lania-edge"
}

variable "acme_email" {
  description = "Email used for Let's Encrypt certificate registration"
  type        = string
}
