data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"]

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }
}

resource "tls_private_key" "bastion" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_key_pair" "bastion" {
  key_name   = "${var.name_prefix}-bastion"
  public_key = tls_private_key.bastion.public_key_openssh
  tags       = var.tags
}

resource "aws_instance" "bastion" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = var.vm_size
  subnet_id              = var.subnet_id
  key_name               = aws_key_pair.bastion.key_name
  vpc_security_group_ids = []
  tags                   = merge(var.tags, { Name = "${var.name_prefix}-bastion" })

  root_block_device {
    volume_size = var.os_disk_size_gb
    volume_type = "gp3"
  }
}