# Golden fixtures (W0)

Layer tfvars omit Azure resource_group_create and resource_group_name because
PrepareLayerVars queries the live subscription (Azure CLI).
All other preparation matches internal/runner before WriteAutoTofuVarsJSON.
