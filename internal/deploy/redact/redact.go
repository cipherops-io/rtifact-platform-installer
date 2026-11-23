package redact

import (
	"sort"
	"strings"
)

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
	for _, l := range r.literals {
		if l == value {
			return
		}
	}
	r.literals = append(r.literals, value)
	// Longest first: masking a short secret that prefixes a longer one would leak the tail.
	sort.SliceStable(r.literals, func(i, j int) bool { return len(r.literals[i]) > len(r.literals[j]) })
}

// Text masks recorded literals. Callers should mask ${env.*} tokens separately.
func (r *Redactor) Text(s string) string {
	out := s
	for _, lit := range r.literals {
		out = strings.ReplaceAll(out, lit, "***")
	}
	return out
}
