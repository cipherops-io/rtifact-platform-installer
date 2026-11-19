output "vnet_id" {
  value = aws_vpc.main.id
}

output "vnet_name" {
  value = "${var.name_prefix}-vpc"
}

output "node_subnet_id" {
  value = aws_subnet.node.id
}

output "load_balancer_subnet_id" {
  value = aws_subnet.load_balancer.id
}

output "private_endpoint_subnet_id" {
  value = aws_subnet.private_endpoint.id
}

output "management_subnet_id" {
  value = aws_subnet.management.id
}

output "nat_gateway_id" {
  value = aws_nat_gateway.main.id
}
