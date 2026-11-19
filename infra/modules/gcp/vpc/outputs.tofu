output "vnet_id" { value = google_compute_network.main.id }
output "vnet_name" { value = google_compute_network.main.name }
output "node_subnet_id" { value = google_compute_subnetwork.node.id }
output "load_balancer_subnet_id" { value = google_compute_subnetwork.load_balancer.id }
output "private_endpoint_subnet_id" { value = google_compute_subnetwork.private_endpoint.id }
output "management_subnet_id" { value = google_compute_subnetwork.management.id }
output "nat_gateway_id" { value = google_compute_router_nat.main.id }
