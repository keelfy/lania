terraform {
  required_providers {
    hetznercloud = {
      source  = "hetznercloud/hcloud"
      version = "~> 1.47"
    }
  }
}

provider "hcloud" {
  token = var.hetzner_token
}

data "hcloud_ssh_key" "key" {
  name = var.ssh_key_name
}

resource "hcloud_server" "minecraft" {
  name        = var.server_name
  image       = "ubuntu-22.04"
  server_type = "ccx33"
  location    = "fsn1"
  ssh_keys    = [data.hcloud_ssh_key.key.id]

  firewall_ids = [hetznercloud_firewall.minecraft.id]

  user_data = templatefile("${path.module}/scripts/setup.sh", {
    neoforge_version = var.neoforge_version
    mrpack_url       = var.mrpack_url
  })

  labels = {
    role = "minecraft"
  }
}