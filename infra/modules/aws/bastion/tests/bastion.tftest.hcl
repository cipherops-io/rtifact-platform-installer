mock_provider "aws" {
  mock_data "aws_ssm_parameter" {
    defaults = {
      value = "ami-0123456789abcdef0"
    }
  }
}

mock_provider "tls" {}

variables {
  name_prefix = "t"
  location    = "us-east-1"
  vpc_id      = "vpc-0123456789abcdef0"
  subnet_id   = "subnet-0123456789abcdef0"
}

run "ssm_defaults_match_previous_literals" {
  command = plan

  assert {
    condition = (
      aws_instance.bastion.metadata_options[0].http_endpoint == "enabled" &&
      aws_instance.bastion.metadata_options[0].http_tokens == "required" &&
      aws_instance.bastion.metadata_options[0].http_put_response_hop_limit == 2
    )
    error_message = "metadata defaults changed"
  }
  assert {
    condition = (
      aws_instance.bastion.root_block_device[0].volume_size == 100 &&
      aws_instance.bastion.root_block_device[0].volume_type == "gp3" &&
      aws_instance.bastion.root_block_device[0].encrypted &&
      aws_instance.bastion.root_block_device[0].delete_on_termination
    )
    error_message = "disk defaults changed"
  }
  assert {
    condition     = length(tls_private_key.bastion) == 0 && length(aws_key_pair.bastion) == 0 && length(aws_security_group.bastion.ingress) == 0
    error_message = "ssm access must not create keys or SSH ingress"
  }
}

run "ssh_generates_key_pair" {
  variables {
    bastion = {
      access        = "ssh"
      public_ip     = true
      allowed_cidrs = ["203.0.113.0/24"]
    }
  }

  assert {
    condition     = tls_private_key.bastion[0].algorithm == "RSA" && tls_private_key.bastion[0].rsa_bits == 4096
    error_message = "default generated key must be RSA 4096"
  }
  assert {
    condition     = aws_key_pair.bastion[0].public_key == tls_private_key.bastion[0].public_key_openssh
    error_message = "key pair must use the generated public key"
  }
  assert {
    condition     = output.ssh_private_key_pem == tls_private_key.bastion[0].private_key_pem
    error_message = "RSA key must be output as PEM"
  }
  assert {
    condition     = alltrue([for r in aws_security_group.bastion.ingress : r.from_port == 22 && r.to_port == 22 && r.cidr_blocks == tolist(["203.0.113.0/24"])])
    error_message = "SSH ingress must allow only bastion.allowed_cidrs on port 22"
  }
}

run "ssh_with_supplied_public_key" {
  command = plan

  variables {
    bastion = {
      access         = "ssh"
      allowed_cidrs  = ["203.0.113.0/24"]
      ssh_public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIG9vZm9vZm9vZm9vZm9vZm9vZm9vZm9vZm9vZm9vZm9v user@host"
    }
  }

  assert {
    condition     = length(tls_private_key.bastion) == 0 && output.ssh_private_key_pem == null
    error_message = "no key may be generated when ssh_public_key is set"
  }
}

run "ed25519_outputs_openssh_key" {
  variables {
    bastion = {
      access        = "ssh"
      allowed_cidrs = ["203.0.113.0/24"]
      ssh_key       = { algorithm = "ED25519" }
    }
  }

  assert {
    condition     = tls_private_key.bastion[0].algorithm == "ED25519" && output.ssh_private_key_pem == tls_private_key.bastion[0].private_key_openssh
    error_message = "ED25519 key must be output in OpenSSH format"
  }
}

run "rejects_ssh_open_to_internet" {
  command = plan

  variables {
    bastion = {
      access        = "ssh"
      allowed_cidrs = ["0.0.0.0/0"]
    }
  }

  expect_failures = [var.bastion]
}

run "rejects_ssh_without_cidrs" {
  command = plan

  variables {
    bastion = { access = "ssh" }
  }

  expect_failures = [var.bastion]
}
