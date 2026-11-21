data "aws_ssm_parameter" "al2023_ami" {
  count = var.bastion.ami_id == "" ? 1 : 0
  name = var.bastion.architecture == "arm64" ? "/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-arm64" : "/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-x86_64"
}

locals {
  ssh_mode = var.bastion.access == "ssh"
  disk     = var.bastion.disk
  ami_id   = var.bastion.ami_id != "" ? var.bastion.ami_id : data.aws_ssm_parameter.al2023_ami[0].value
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
      from_port   = 22
      to_port     = 22
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

resource "aws_key_pair" "bastion" {
  count      = local.ssh_mode && var.bastion.ssh_public_key != "" ? 1 : 0
  key_name   = "${var.name_prefix}-bastion"
  public_key = var.bastion.ssh_public_key
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
  key_name               = local.ssh_mode && length(aws_key_pair.bastion) > 0 ? aws_key_pair.bastion[0].key_name : null

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 2
  }

  root_block_device {
    volume_size           = local.disk.size_gb
    volume_type           = local.disk.type
    encrypted             = local.disk.encrypted
    delete_on_termination = true
  }

  tags = merge(var.tags, { Name = "${var.name_prefix}-bastion" })
}

resource "aws_eip_association" "bastion" {
  count         = var.bastion.public_ip ? 1 : 0
  instance_id   = aws_instance.bastion.id
  allocation_id = aws_eip.bastion[0].id
}