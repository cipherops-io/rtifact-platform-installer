data "aws_availability_zones" "available" {
  state = "available"
}

locals {
  az = data.aws_availability_zones.available.names[0]
}

resource "aws_vpc" "main" {
  cidr_block           = var.vnet_cidr
  enable_dns_support   = true
  enable_dns_hostnames = true
  tags                 = merge(var.tags, { Name = "${var.name_prefix}-vpc" })
}

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id
  tags   = merge(var.tags, { Name = "${var.name_prefix}-igw" })
}

resource "aws_subnet" "node" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = var.node_subnet_cidr
  availability_zone       = local.az
  map_public_ip_on_launch = false
  tags                    = merge(var.tags, { Name = "aks-nodes" })
}

resource "aws_subnet" "load_balancer" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = var.load_balancer_subnet_cidr
  availability_zone       = local.az
  map_public_ip_on_launch = false
  tags                    = merge(var.tags, { Name = "load-balancer" })
}

resource "aws_subnet" "private_endpoint" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = var.private_endpoint_subnet_cidr
  availability_zone = local.az
  tags              = merge(var.tags, { Name = "private-endpoints" })
}

resource "aws_subnet" "management" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = var.management_subnet_cidr
  availability_zone = local.az
  tags              = merge(var.tags, { Name = "management" })
}

resource "aws_eip" "nat" {
  domain = "vpc"
  tags   = merge(var.tags, { Name = "${var.name_prefix}-nat-eip" })
}

resource "aws_nat_gateway" "main" {
  allocation_id = aws_eip.nat.id
  subnet_id     = aws_subnet.load_balancer.id
  tags          = merge(var.tags, { Name = "${var.name_prefix}-nat" })

  depends_on = [aws_internet_gateway.main]
}

resource "aws_route_table" "private" {
  vpc_id = aws_vpc.main.id
  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.main.id
  }
  tags = merge(var.tags, { Name = "${var.name_prefix}-private-rt" })
}

resource "aws_route_table_association" "node" {
  subnet_id      = aws_subnet.node.id
  route_table_id = aws_route_table.private.id
}

resource "aws_route_table_association" "private_endpoint" {
  subnet_id      = aws_subnet.private_endpoint.id
  route_table_id = aws_route_table.private.id
}

resource "aws_route_table_association" "management" {
  subnet_id      = aws_subnet.management.id
  route_table_id = aws_route_table.private.id
}
