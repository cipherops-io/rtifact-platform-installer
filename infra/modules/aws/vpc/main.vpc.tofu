locals {
  azs = var.network.availability_zones

  subnet_specs = flatten([
    for tier, cfg in var.network.subnets : [
      for i, cidr in cfg.cidrs : {
        key                 = "${tier}-${local.azs[i]}"
        tier                = tier
        cidr                = cidr
        az                  = local.azs[i]
        route               = cfg.route
        public_ip_on_launch = try(cfg.public_ip_on_launch, false)
        roles               = try(cfg.roles, [])
      }
    ]
  ])

  role_tag_pairs = {
    cluster      = { "kubernetes.io/cluster/${var.cluster_name}" = "owned" }
    nodes        = { "kubernetes.io/role/internal-elb" = "1" }
    internal_lb  = { "kubernetes.io/role/internal-elb" = "1" }
    public_lb    = { "kubernetes.io/role/elb" = "1" }
    karpenter    = { "karpenter.sh/discovery" = var.cluster_name }
  }

  subnet_route = { for s in local.subnet_specs : s.key => s.route }
}

resource "aws_vpc" "main" {
  cidr_block           = var.network.vpc_cidr
  enable_dns_support   = true
  enable_dns_hostnames = true
  tags                 = merge(var.tags, { Name = "${var.name_prefix}-vpc" })
}

resource "aws_vpc_ipv4_cidr_block_association" "secondary" {
  for_each   = toset(var.network.secondary_cidrs)
  vpc_id     = aws_vpc.main.id
  cidr_block = each.value
}

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id
  tags   = merge(var.tags, { Name = "${var.name_prefix}-igw" })
}

resource "aws_subnet" "tier" {
  for_each = { for s in local.subnet_specs : s.key => s }

  vpc_id                  = aws_vpc.main.id
  cidr_block              = each.value.cidr
  availability_zone       = each.value.az
  map_public_ip_on_launch = each.value.public_ip_on_launch
  tags = merge(
    var.tags,
    { Name = "${var.name_prefix}-${each.value.tier}-${each.value.az}", Tier = each.value.tier },
    merge([for r in each.value.roles : lookup(local.role_tag_pairs, r, {})]...)
  )
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }
  tags = merge(var.tags, { Name = "${var.name_prefix}-public-rt" })
}

resource "aws_route_table" "isolated" {
  vpc_id = aws_vpc.main.id
  tags   = merge(var.tags, { Name = "${var.name_prefix}-isolated-rt" })
}

resource "aws_route_table" "private" {
  for_each = toset(local.azs)
  vpc_id   = aws_vpc.main.id
  tags     = merge(var.tags, { Name = "${var.name_prefix}-private-${each.key}-rt" })
}

locals {
  public_subnet_by_az = {
    for s in local.subnet_specs : s.az => s.key if s.route == "internet"
  }
  first_public_az = local.azs[0]
}

resource "aws_eip" "nat" {
  count  = var.network.nat_gateway == "none" ? 0 : var.network.nat_gateway == "per_az" ? length(local.azs) : 1
  domain = "vpc"
  tags   = merge(var.tags, { Name = "${var.name_prefix}-nat-eip-${count.index}" })
}

resource "aws_nat_gateway" "main" {
  count = var.network.nat_gateway == "none" ? 0 : var.network.nat_gateway == "per_az" ? length(local.azs) : 1

  allocation_id = aws_eip.nat[count.index].id
  subnet_id = aws_subnet.tier[
    var.network.nat_gateway == "per_az" ? local.public_subnet_by_az[local.azs[count.index]] : local.public_subnet_by_az[local.first_public_az]
  ].id
  tags          = merge(var.tags, { Name = "${var.name_prefix}-nat-${count.index}" })

  depends_on = [aws_internet_gateway.main]
}

resource "aws_route" "private_nat" {
  for_each = var.network.nat_gateway == "none" ? {} : aws_route_table.private

  route_table_id         = each.value.id
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = var.network.nat_gateway == "per_az" ? aws_nat_gateway.main[index(local.azs, each.key)].id : aws_nat_gateway.main[0].id
}

resource "aws_route_table_association" "tier" {
  for_each = aws_subnet.tier

  subnet_id = each.value.id
  route_table_id = (
    local.subnet_route[each.key] == "isolated" ? aws_route_table.isolated.id :
    local.subnet_route[each.key] == "internet" ? aws_route_table.public.id :
    aws_route_table.private[each.value.availability_zone].id
  )
}

resource "aws_vpc_endpoint" "gateway" {
  for_each = toset(var.network.gateway_endpoints)

  vpc_id       = aws_vpc.main.id
  service_name = "com.amazonaws.${var.location}.${each.key}"
  tags         = merge(var.tags, { Name = "${var.name_prefix}-${each.key}-endpoint" })
}

resource "aws_vpc_endpoint_route_table_association" "gateway" {
  for_each = {
    for pair in setproduct(keys(aws_vpc_endpoint.gateway), concat([aws_route_table.public.id, aws_route_table.isolated.id], [for rt in aws_route_table.private : rt.id])) :
    "${pair[0]}-${pair[1]}" => {
      endpoint_key = pair[0]
      rt_id        = pair[1]
    }
  }

  route_table_id  = each.value.rt_id
  vpc_endpoint_id = aws_vpc_endpoint.gateway[each.value.endpoint_key].id
}