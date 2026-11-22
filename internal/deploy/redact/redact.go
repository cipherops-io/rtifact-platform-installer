package redact

import "strings"

// Redactor records resolved secret literals and masks them in output.
type Redactor struct {
	literals []string
}

// New returns an empty redactor.
func New() *Redactor { return &Redactor{} }

// Record registers a secret literal.
func (r *Redactor) Record(value string) {
	if value == "" {
		return
	}
	r.literals = append(r.literals, value)
}

// Text masks recorded literals. Callers should mask ${env.*} tokens separately.
func (r *Redactor) Text(s string) string {
	out := s
	for _, lit := range r.literals {
		if lit != "" {
			out = strings.ReplaceAll(out, lit, "***")
		}
	}
	return out
}