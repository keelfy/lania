output "ssh_connect" {
  value = "ssh ${var.ssh_user}@${var.server_ipv4}"
}

output "deploy_path" {
  value = var.deploy_path
}

output "network_name" {
  description = "External Docker network other site stacks must join to be routed by Traefik"
  value       = "edge"
}
