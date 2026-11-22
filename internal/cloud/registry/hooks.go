package registry

// Hooks are cloud-specific defaults (schema, maturity, quota).
type Hooks interface {
	DefaultMaturity() string
	CustomerSchemaUsesCommonAllOf() bool
	DefaultQuotaCheckOnValidate() bool
}

// HooksFor returns defaults for name.
func HooksFor(name string) (Hooks, error) {
	p, err := Get(name)
	if err != nil {
		return nil, err
	}
	h, ok := p.(Hooks)
	if !ok {
		return nil, err
	}
	return h, nil
}