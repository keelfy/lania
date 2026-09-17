locals {
  compose_src_dir = "${path.module}/../../../compose"
}

resource "local_sensitive_file" "env" {
  filename = "${path.module}/generated/.env"

  content = templatefile("${path.module}/templates/env.tftpl", {
    image_tag = var.image_tag
    domain    = var.domain

    api_key    = var.api_key
    jwt_secret = var.jwt_secret

    database_host          = var.database_host
    database_port          = var.database_port
    database_user          = var.database_user
    database_password      = var.database_password
    database_name          = var.database_name
    database_flectone_name = var.database_flectone_name
    database_plan_name     = var.database_plan_name

    active_season_id      = var.active_season_id
    max_profiles_per_user = var.max_profiles_per_user
    default_name_color_id = var.default_name_color_id
    preregistration       = var.preregistration

    freekassa_base_payment_url    = var.freekassa_base_payment_url
    freekassa_merchant_id         = var.freekassa_merchant_id
    freekassa_merchant_password_1 = var.freekassa_merchant_password_1
    freekassa_merchant_password_2 = var.freekassa_merchant_password_2

    donation_alerts_client_id               = var.donation_alerts_client_id
    donation_alerts_client_secret           = var.donation_alerts_client_secret
    donation_alerts_socket_connection_token = var.donation_alerts_socket_connection_token
    donation_alerts_user_id                 = var.donation_alerts_user_id

    twitch_client_id     = var.twitch_client_id
    twitch_client_secret = var.twitch_client_secret

    support_email        = var.support_email
    telegram_channel_url = var.telegram_channel_url
    discord_server_url   = var.discord_server_url

    mariadb_root_password = var.mariadb_root_password

    postgres_user     = var.postgres_user
    postgres_password = var.postgres_password
    postgres_db       = var.postgres_db

    kratos_secrets_cookie       = var.kratos_secrets_cookie
    kratos_secrets_cipher       = var.kratos_secrets_cipher
    courier_smtp_connection_uri = var.courier_smtp_connection_uri

    oidc_google_client_id      = var.oidc_google_client_id
    oidc_google_client_secret  = var.oidc_google_client_secret
    oidc_discord_client_id     = var.oidc_discord_client_id
    oidc_discord_client_secret = var.oidc_discord_client_secret
    oidc_yandex_client_id      = var.oidc_yandex_client_id
    oidc_yandex_client_secret  = var.oidc_yandex_client_secret
    oidc_twitch_client_id      = var.oidc_twitch_client_id
    oidc_twitch_client_secret  = var.oidc_twitch_client_secret
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
    destination = "/tmp/lania-web-bootstrap.sh"
  }

  provisioner "remote-exec" {
    inline = [
      "chmod +x /tmp/lania-web-bootstrap.sh",
      "/tmp/lania-web-bootstrap.sh '${var.deploy_path}'",
    ]
  }
}

resource "null_resource" "deploy" {
  triggers = {
    compose_hash = sha1(join("", [
      for f in fileset(local.compose_src_dir, "**") : filesha1("${local.compose_src_dir}/${f}")
    ]))
    env_hash  = sha1(local_sensitive_file.env.content)
    image_tag = var.image_tag
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
      "chmod +x ${var.deploy_path}/compose/mariadb-init.sh",
      "cd ${var.deploy_path}/compose && docker compose -f docker-compose.prod.yml pull",
      "cd ${var.deploy_path}/compose && docker compose -f docker-compose.prod.yml up -d --remove-orphans",
    ]
  }
}
