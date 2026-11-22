package azure

func (*Provider) DefaultMaturity() string             { return "operational" }
func (*Provider) CustomerSchemaUsesCommonAllOf() bool { return false }
func (*Provider) DefaultQuotaCheckOnValidate() bool   { return true }