output "vpc_id" {
  value = aws_vpc.main.id
}

output "vpc_cidrs" {
  value = concat([aws_vpc.main.cidr_block], [for a in aws_vpc_ipv4_cidr_block_association.secondary : a.cidr_block])
}

output "availability_zones" {
  value = local.azs
}

output "subnet_ids_by_tier" {
  value = {
    for tier in distinct([for s in local.subnet_specs : s.tier]) :
    tier => [for k, s in aws_subnet.tier : s.id if s.tags["Tier"] == tier]
  }
}

output "route_table_ids" {
  value = {
    public   = aws_route_table.public.id
    isolated = aws_route_table.isolated.id
    private  = { for az, rt in aws_route_table.private : az => rt.id }
  }
}

output "nat_gateway_public_ips" {
  value = [for e in aws_eip.nat : e.public_ip]
}

output "internet_gateway_id" {
  value = aws_internet_gateway.main.id
}