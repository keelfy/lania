resource "local_file" "bash_aliases" {
  content = templatefile("${path.module}/scripts/bash_aliases.tftpl", {
    rcon_password = var.rcon_password
    rcon_port     = var.rcon_port
  })

  filename = "${path.module}/generated/bash_aliases"
}

resource "null_resource" "upload_bash_aliases" {
  triggers = {
    server_ip = hcloud_server.minecraft.ipv4_address
    hash      = sha1(templatefile("${path.module}/scripts/bash_aliases.tftpl", {
      server_ip     = hcloud_server.minecraft.ipv4_address
      rcon_password = var.rcon_password
      rcon_port     = var.rcon_port
    }))
  }

  connection {
    type        = "ssh"
    user        = "root"
    private_key = file(var.ssh_private_key_path)
    host        = hcloud_server.minecraft.ipv4_address
  }

  provisioner "remote-exec" {
    inline = [
      "echo 'Waiting for setup.sh to finish...'",
      "while [ ! -f /var/log/minecraft-setup-done ]; do sleep 5; done"
    ]
  }

  provisioner "file" {
    source      = local_file.bash_aliases.filename
    destination = "/home/minecraft/.bash_aliases"
  }

  provisioner "remote-exec" {
    inline = [
      "sudo touch /home/minecraft/.bash_aliases",
      "sudo chown minecraft:minecraft /home/minecraft/.bash_aliases",
      "sudo chmod 644 /home/minecraft/.bash_aliases",
    ]
  }

  depends_on = [local_file.bash_aliases, hcloud_server.minecraft]
}