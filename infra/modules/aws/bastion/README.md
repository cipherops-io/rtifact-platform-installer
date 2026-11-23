
<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| tls | >= 4.0 |

## Providers

| Name | Version |
|------|---------|
| aws | n/a |
| tls | >= 4.0 |

## Resources

| Name | Type |
|------|------|
| [aws_eip.bastion](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/eip) | resource |
| [aws_eip_association.bastion](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/eip_association) | resource |
| [aws_iam_instance_profile.bastion](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/iam_instance_profile) | resource |
| [aws_iam_role.bastion](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/iam_role) | resource |
| [aws_iam_role_policy_attachment.bastion](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [aws_instance.bastion](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/instance) | resource |
| [aws_key_pair.bastion](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/key_pair) | resource |
| [aws_security_group.bastion](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/security_group) | resource |
| [tls_private_key.bastion](https://registry.terraform.io/providers/opentofu/tls/latest/docs/resources/private_key) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| bastion | n/a | <pre>object({<br/>    access          = optional(string, "ssm")<br/>    instance_type   = optional(string, "t3.medium")<br/>    public_ip       = optional(bool, false)<br/>    allowed_cidrs   = optional(list(string), [])<br/>    ssh_port        = optional(number, 22)<br/>    ssh_public_key  = optional(string, "")<br/>    admin_username  = optional(string, "ec2-user")<br/>    architecture    = optional(string, "x86_64")<br/>    ami_id          = optional(string, "")<br/>    iam_policy_arns = optional(list(string), ["arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"])<br/>    ssh_key = optional(object({<br/>      algorithm = optional(string, "RSA")<br/>      rsa_bits  = optional(number, 4096)<br/>    }), {})<br/>    metadata = optional(object({<br/>      http_endpoint = optional(string, "enabled")<br/>      http_tokens   = optional(string, "required")<br/>      hop_limit     = optional(number, 2)<br/>    }), {})<br/>    disk = optional(object({<br/>      size_gb               = optional(number, 100)<br/>      type                  = optional(string, "gp3")<br/>      encrypted             = optional(bool, true)<br/>      delete_on_termination = optional(bool, true)<br/>    }), {})<br/>  })</pre> | `{}` | no |
| location | n/a | `string` | n/a | yes |
| name_prefix | n/a | `string` | n/a | yes |
| subnet_id | n/a | `string` | n/a | yes |
| tags | n/a | `map(string)` | `{}` | no |
| vpc_id | n/a | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| access | n/a |
| admin_username | n/a |
| instance_id | n/a |
| private_ip | n/a |
| public_ip | n/a |
| ssh_port | n/a |
| ssh_private_key_pem | Generated SSH private key (PEM for RSA, OpenSSH for ED25519); null when access is ssm or a public key was supplied. |
<!-- END_TF_DOCS -->
