output "ssh_connect" {
  description = "Command to connect to the server"
  value       = "ssh ${var.ssh_user}@${var.server_ipv4}"
}

output "dns_records" {
  description = "DNS records managed by this stack"
  value       = { for k, r in cloudflare_record.web : k => "${r.name} -> ${r.content}" }
}

output "deploy_path" {
  value = var.deploy_path
}
