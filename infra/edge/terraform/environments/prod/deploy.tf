locals {
  compose_src_dir = "${path.module}/../../../compose"
}

resource "local_sensitive_file" "env" {
  filename = "${path.module}/generated/.env"

  content = templatefile("${path.module}/templates/env.tftpl", {
    acme_email = var.acme_email
  })
}

resource "null_resource" "bootstrap" {
  triggers = {
    server_ip = var.server_ipv4
  }

  connection {
    type        = "ssh"
    user        = var.ssh_user
    private_key = file(var.ssh_private_key_path)
    host        = var.server_ipv4
  }

  provisioner "file" {
    source      = "${path.module}/../../../../scripts/bootstrap-docker.sh"
    destination = "/tmp/lania-edge-bootstrap.sh"
  }

  provisioner "remote-exec" {
    inline = [
      "chmod +x /tmp/lania-edge-bootstrap.sh",
      "/tmp/lania-edge-bootstrap.sh '${var.deploy_path}'",
    ]
  }
}

resource "null_resource" "deploy" {
  triggers = {
    compose_hash = sha1(join("", [
      for f in fileset(local.compose_src_dir, "**") : filesha1("${local.compose_src_dir}/${f}")
    ]))
    env_hash = sha1(local_sensitive_file.env.content)
  }

  depends_on = [null_resource.bootstrap, local_sensitive_file.env]

  connection {
    type        = "ssh"
    user        = var.ssh_user
    private_key = file(var.ssh_private_key_path)
    host        = var.server_ipv4
  }

  provisioner "file" {
    source      = "${local.compose_src_dir}/"
    destination = "${var.deploy_path}/compose"
  }

  provisioner "file" {
    source      = local_sensitive_file.env.filename
    destination = "${var.deploy_path}/compose/.env"
  }

  provisioner "remote-exec" {
    inline = [
      "cd ${var.deploy_path}/compose && docker compose -f docker-compose.yml pull",
      "cd ${var.deploy_path}/compose && docker compose -f docker-compose.yml up -d --remove-orphans",
    ]
  }
}
