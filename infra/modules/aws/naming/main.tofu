resource "aws_resourcegroups_group" "main" {
  name = "${var.prefix}-rg"

  resource_query {
    query = jsonencode({
      ResourceTypeFilters = ["AWS::AllSupported"]
      TagFilters = [{
        Key    = "scope"
        Values = [var.prefix]
      }]
    })
  }

  tags = local.tags
}
