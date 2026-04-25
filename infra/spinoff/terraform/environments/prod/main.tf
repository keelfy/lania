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
  })

  labels = {
    role = "minecraft"
  }
}

resource "null_resource" "update_modpack" {
  triggers = {
    mrpack_url      = var.mrpack_url
    bootstrap_force = var.bootstrap_force
  }

  connection {
    type        = "ssh"
    user        = "minecraft"
    private_key = file(var.ssh_private_key_path)
    host        = hcloud_server.minecraft.ipv4_address
  }
  
  provisioner "file" {
    source      = "${path.module}/scripts/update.sh"
    destination = "/opt/minecraft/update.sh"
  }

  provisioner "remote-exec" {
    inline = [
      "echo 'Waiting for setup.sh to finish...'",
      "while [ ! -f /var/log/minecraft-setup-done ]; do sleep 5; done",
      "chmod +x /opt/minecraft/update.sh",
      "chown minecraft:minecraft /opt/minecraft/update.sh",
      "/opt/minecraft/update.sh '${var.mrpack_url}' '${var.neoforge_version}'"
    ]
  }

  depends_on = [hcloud_server.minecraft]
}
