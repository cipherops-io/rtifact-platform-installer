package aws

func (*Provider) DefaultMaturity() string             { return "preview" }
func (*Provider) CustomerSchemaUsesCommonAllOf() bool { return true }
func (*Provider) DefaultQuotaCheckOnValidate() bool   { return false }