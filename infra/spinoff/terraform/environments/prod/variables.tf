variable "hetzner_token" {
  description = "Hetzner Cloud API token"
  type        = string
  sensitive   = true
}

variable "ssh_key_name" {
  description = "Name of the SSH key in Hetzner Cloud"
  type        = string
}

variable "mrpack_url" {
  description = "Direct URL to the .mrpack file (raw GitHub)"
  type        = string
  default     = ""
}

variable "neoforge_version" {
  description = "NeoForge version"
  type        = string
  default     = "21.1.228"
}

variable "server_name" {
  description = "Hetzner server name"
  type        = string
  default     = "minecraft-server"
}

variable "ssh_private_key_path" {
  description = "SSH private key for the server"
  type        = string
  default     = "~/.ssh/id_rsa"
}