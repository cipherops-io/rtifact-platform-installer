# EKS Pod Identity associations (one IAM role per entry).
# Named policy documents for common controllers ship under policies/; scope ARNs at apply time.

locals {
  entries = var.entries
}

data "aws_iam_policy_document" "trust" {
  statement {
    actions = ["sts:AssumeRole", "sts:TagSession"]
    principals {
      type        = "Service"
      identifiers = ["pods.eks.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "entry" {
  for_each = local.entries
  name     = "${each.key}-pod-identity"
  assume_role_policy = data.aws_iam_policy_document.trust.json
  tags                 = var.tags
}

resource "aws_iam_role_policy" "inline" {
  for_each = {
    for k, v in local.entries : k => v if trimspace(try(v.policy_json, "")) != ""
  }
  name   = "${each.key}-inline"
  role   = aws_iam_role.entry[each.key].id
  policy = each.value.policy_json
}

resource "aws_iam_role_policy_attachment" "managed" {
  for_each = {
    for pair in flatten([
      for name, e in local.entries : [
        for arn in coalesce(e.policy_arns, []) : { key = "${name}:${arn}", name = name, arn = arn }
      ]
    ]) : pair.key => pair
  }
  role       = aws_iam_role.entry[each.value.name].name
  policy_arn = each.value.arn
}

resource "aws_eks_pod_identity_association" "entry" {
  for_each        = local.entries
  cluster_name    = var.cluster_name
  namespace       = each.value.namespace
  service_account = each.value.service_account
  role_arn        = aws_iam_role.entry[each.key].arn
}