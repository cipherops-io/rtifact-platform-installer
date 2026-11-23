data "aws_ssm_parameter" "al2023_ami" {
  count = var.bastion.ami_id == "" ? 1 : 0
  name  = var.bastion.architecture == "arm64" ? "/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-arm64" : "/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-x86_64"
}

locals {
  ssh_mode     = var.bastion.access == "ssh"
  generate_key = local.ssh_mode && var.bastion.ssh_public_key == ""
  disk         = var.bastion.disk
  ami_id       = var.bastion.ami_id != "" ? var.bastion.ami_id : data.aws_ssm_parameter.al2023_ami[0].value
}

resource "aws_iam_role" "bastion" {
  name = "${var.name_prefix}-bastion"

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

resource "aws_iam_role_policy_attachment" "bastion" {
  for_each   = toset(var.bastion.iam_policy_arns)
  role       = aws_iam_role.bastion.name
  policy_arn = each.value
}

resource "aws_iam_instance_profile" "bastion" {
  name = "${var.name_prefix}-bastion"
  role = aws_iam_role.bastion.name
}

resource "aws_security_group" "bastion" {
  name   = "${var.name_prefix}-bastion"
  vpc_id = var.vpc_id

  dynamic "ingress" {
    for_each = local.ssh_mode ? var.bastion.allowed_cidrs : []
    content {
      from_port   = var.bastion.ssh_port
      to_port     = var.bastion.ssh_port
      protocol    = "tcp"
      cidr_blocks = [ingress.value]
    }
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(var.tags, { Name = "${var.name_prefix}-bastion-sg" })
}

resource "tls_private_key" "bastion" {
  count     = local.generate_key ? 1 : 0
  algorithm = var.bastion.ssh_key.algorithm
  rsa_bits  = var.bastion.ssh_key.algorithm == "RSA" ? var.bastion.ssh_key.rsa_bits : null
}

resource "aws_key_pair" "bastion" {
  count      = local.ssh_mode ? 1 : 0
  key_name   = "${var.name_prefix}-bastion"
  public_key = local.generate_key ? tls_private_key.bastion[0].public_key_openssh : var.bastion.ssh_public_key
  tags       = var.tags
}

resource "aws_eip" "bastion" {
  count  = var.bastion.public_ip ? 1 : 0
  domain = "vpc"
  tags   = merge(var.tags, { Name = "${var.name_prefix}-bastion-eip" })
}

resource "aws_instance" "bastion" {
  ami                    = local.ami_id
  instance_type          = var.bastion.instance_type
  subnet_id              = var.subnet_id
  vpc_security_group_ids = [aws_security_group.bastion.id]
  iam_instance_profile   = aws_iam_instance_profile.bastion.name
  key_name               = one(aws_key_pair.bastion[*].key_name)

  metadata_options {
    http_endpoint               = var.bastion.metadata.http_endpoint
    http_tokens                 = var.bastion.metadata.http_tokens
    http_put_response_hop_limit = var.bastion.metadata.hop_limit
  }

  root_block_device {
    volume_size           = local.disk.size_gb
    volume_type           = local.disk.type
    encrypted             = local.disk.encrypted
    delete_on_termination = local.disk.delete_on_termination
  }

  tags = merge(var.tags, { Name = "${var.name_prefix}-bastion" })
}

resource "aws_eip_association" "bastion" {
  count         = var.bastion.public_ip ? 1 : 0
  instance_id   = aws_instance.bastion.id
  allocation_id = aws_eip.bastion[0].id
}
