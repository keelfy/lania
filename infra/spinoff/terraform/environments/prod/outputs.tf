output "server_ip" {
  description = "IP адрес сервера"
  value       = hcloud_server.minecraft.ipv4_address
}

output "ssh_connect" {
  description = "Команда для подключения к серверу"
  value       = "ssh minecraft@${hcloud_server.minecraft.ipv4_address}"
}