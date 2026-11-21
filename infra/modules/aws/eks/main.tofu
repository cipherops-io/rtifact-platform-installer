locals {
  cfg       = var.kubernetes_cluster
  name      = try(local.cfg.name, "${var.name_prefix}-eks")
  endpoint  = try(local.cfg.endpoint, {})
  cluster_subnet_ids = distinct(flatten([
    for tier in ["node", "cluster_private"] : try(var.subnet_ids_by_tier[tier], [])
  ]))
  node_subnet_ids   = try(var.subnet_ids_by_tier[try(local.cfg.node_groups.initial.subnet_tier, "node")], local.cluster_subnet_ids)
  vpc_cidr_strings  = var.vpc_cidrs
}

resource "aws_iam_role" "cluster" {
  name = "${var.name_prefix}-eks-cluster"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "eks.amazonaws.com" }
    }]
  })

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "cluster" {
  for_each = toset(["AmazonEKSClusterPolicy", "AmazonEKSVPCResourceController"])
  role       = aws_iam_role.cluster.name
  policy_arn = "arn:aws:iam::aws:policy/${each.value}"
}

resource "aws_security_group" "cluster_extra" {
  name        = "${local.name}-cluster-sg"
  description = "Extra cluster security group"
  vpc_id      = var.vpc_id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = local.vpc_cidr_strings
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(var.tags, { Name = "${local.name}-cluster-sg", "karpenter.sh/discovery" = local.name })
}

resource "aws_security_group" "service" {
  for_each = try(local.cfg.service_security_groups, {})

  name   = "${local.name}-${each.key}-svc"
  vpc_id = var.vpc_id

  dynamic "ingress" {
    for_each = try(each.value.ports, [80, 443])
    content {
      from_port = ingress.value
      to_port   = ingress.value
      protocol  = "tcp"
      cidr_blocks = flatten([
        for c in try(each.value.cidrs, []) :
        c == "vpc" ? local.vpc_cidr_strings : [c]
      ])
    }
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(var.tags, { Name = "${local.name}-${each.key}-svc" })
}

resource "aws_eks_cluster" "main" {
  name     = local.name
  role_arn = aws_iam_role.cluster.arn
  version  = local.cfg.kubernetes_version
  tags     = var.tags

  deletion_protection = try(local.cfg.deletion_protection, false)

  dynamic "upgrade_policy" {
    for_each = try(local.cfg.support_type, "") != "" ? [1] : []
    content {
      support_type = local.cfg.support_type
    }
  }

  access_config {
    authentication_mode                         = try(local.cfg.authentication_mode, "API_AND_CONFIG_MAP")
    bootstrap_cluster_creator_admin_permissions = false
  }

  vpc_config {
    subnet_ids              = local.cluster_subnet_ids
    endpoint_private_access = try(local.endpoint.private, true)
    endpoint_public_access  = try(local.endpoint.public, false)
    public_access_cidrs     = try(local.endpoint.public_access_cidrs, [])
    security_group_ids      = [aws_security_group.cluster_extra.id]
  }

  kubernetes_network_config {
    service_ipv4_cidr = try(local.cfg.service_cidr, null)
  }

  enabled_cluster_log_types = try(local.cfg.log_types, [])

  dynamic "encryption_config" {
    for_each = try(local.cfg.secrets_encryption.mode, "none") == "kms" ? [1] : []
    content {
      provider {
        key_arn = local.cfg.secrets_encryption.kms_key_arn
      }
      resources = ["secrets"]
    }
  }

  depends_on = [aws_iam_role_policy_attachment.cluster]
}

resource "aws_eks_access_entry" "admin" {
  for_each = toset(try(local.cfg.admin_principal_arns, []))

  cluster_name  = aws_eks_cluster.main.name
  principal_arn = each.value
  type          = "STANDARD"
}

resource "aws_eks_access_policy_association" "admin" {
  for_each = toset(try(local.cfg.admin_principal_arns, []))

  cluster_name  = aws_eks_cluster.main.name
  principal_arn = aws_eks_access_entry.admin[each.key].principal_arn
  policy_arn    = "arn:aws:eks::aws:cluster-access-policy/AmazonEKSClusterAdminPolicy"

  access_scope {
    type = "cluster"
  }
}

resource "aws_iam_role" "node" {
  name = "${var.name_prefix}-eks-node"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ec2.amazonaws.com" }
    }]
  })

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "node" {
  for_each   = toset(try(local.cfg.node_role_policies, ["AmazonEKSWorkerNodePolicy", "AmazonEKS_CNI_Policy", "AmazonEC2ContainerRegistryReadOnly"]))
  role       = aws_iam_role.node.name
  policy_arn = "arn:aws:iam::aws:policy/${each.value}"
}

resource "aws_eks_access_entry" "node" {
  cluster_name  = aws_eks_cluster.main.name
  principal_arn = aws_iam_role.node.arn
  type          = "EC2_LINUX"
}

resource "aws_launch_template" "node" {
  for_each = try(local.cfg.node_groups, {})

  name = "${local.name}-${each.key}-lt"

  block_device_mappings {
    device_name = "/dev/xvda"
    ebs {
      volume_size           = try(each.value.disk.size_gb, 100)
      volume_type           = try(each.value.disk.type, "gp3")
      encrypted             = try(each.value.disk.encrypted, true)
      delete_on_termination = true
    }
  }

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = try(each.value.metadata.http_tokens, "optional")
    http_put_response_hop_limit = try(each.value.metadata.hop_limit, 2)
  }

  tag_specifications {
    resource_type = "instance"
    tags          = merge(var.tags, try(each.value.labels, {}))
  }
}

resource "aws_eks_node_group" "pool" {
  for_each = try(local.cfg.node_groups, {})

  cluster_name    = aws_eks_cluster.main.name
  node_group_name = each.key
  node_role_arn   = aws_iam_role.node.arn
  subnet_ids      = try(var.subnet_ids_by_tier[each.value.subnet_tier], local.node_subnet_ids)
  capacity_type  = try(each.value.capacity_type, "ON_DEMAND")
  instance_types = try(each.value.instance_types, ["m6a.large"])
  labels          = try(each.value.labels, {})
  tags            = var.tags

  launch_template {
    id      = aws_launch_template.node[each.key].id
    version = aws_launch_template.node[each.key].latest_version
  }

  scaling_config {
    desired_size = try(each.value.desired_size, 2)
    max_size     = try(each.value.max_size, 3)
    min_size     = try(each.value.min_size, 1)
  }

  update_config {
    max_unavailable = try(each.value.max_unavailable, 1)
  }

  depends_on = [aws_iam_role_policy_attachment.node]
}

resource "aws_eks_addon" "addon" {
  for_each = try(local.cfg.addons, {})

  cluster_name                = aws_eks_cluster.main.name
  addon_name                  = each.key
  addon_version               = try(each.value.version, null)
  configuration_values        = try(jsonencode(each.value.configuration), null)
  resolve_conflicts_on_create = "OVERWRITE"
  resolve_conflicts_on_update = "OVERWRITE"
}