terraform {
  required_providers {
    hcloud = {
      source  = "hetznercloud/hcloud"
      version = "1.61.0"
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

  firewall_ids = [hcloud_firewall.minecraft.id]

  user_data = templatefile("${path.module}/scripts/setup.sh", {
    neoforge_version = var.neoforge_version
    mrpack_url       = var.mrpack_url
  })

  labels = {
    role = "minecraft"
  }
}

resource "null_resource" "update_modpack" {
  triggers = {
    mrpack_url = var.mrpack_url
  }

  connection {
    type        = "ssh"
    user        = "minecraft"
    private_key = file("~/.ssh/id_rsa")
    host        = hcloud_server.minecraft.ipv4_address
  }

  provisioner "remote-exec" {
    inline = [
      "sudo /opt/minecraft/update.sh '${var.mrpack_url}'"
    ]
  }

  depends_on = [hcloud_server.minecraft]
}
