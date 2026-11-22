package usererr

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

var (
	profileMu sync.RWMutex
	profileID string
)

// SetProfileID sets the active customer profile for generic CLI error hints.
func SetProfileID(id string) {
	profileMu.Lock()
	profileID = id
	profileMu.Unlock()
}

func activeProfileID() string {
	profileMu.RLock()
	defer profileMu.RUnlock()
	if profileID != "" {
		return profileID
	}
	return "<profile-id>"
}

// Error is a user-facing failure (no stack trace, no CLI usage dump).
type Error struct {
	Summary string
	Details []string
	Hint    string
}

func (e *Error) Error() string {
	return e.format()
}

func (e *Error) format() string {
	var b strings.Builder
	b.WriteString(e.Summary)
	for _, d := range e.Details {
		if d == "" {
			continue
		}
		b.WriteString("\n  • ")
		b.WriteString(d)
	}
	if e.Hint != "" {
		b.WriteString("\n\nWhat you can do:\n  ")
		b.WriteString(e.Hint)
	}
	return b.String()
}

// New builds a user-facing error with bullet details.
func New(summary string, details ...string) *Error {
	return &Error{Summary: summary, Details: details}
}

// WithHint adds an actionable next step.
func WithHint(summary, hint string, details ...string) *Error {
	return &Error{Summary: summary, Details: details, Hint: hint}
}

// Wrap annotates any error for humans; usererr.Error values pass through.
func Wrap(err error, summary, hint string) error {
	if err == nil {
		return nil
	}
	var ue *Error
	if errors.As(err, &ue) {
		return ue
	}
	details := []string{cleanMsg(err.Error())}
	if hint != "" {
		return WithHint(summary, hint, details...)
	}
	return New(summary, details...)
}

// Human formats any error for terminal output (strips Cobra usage noise).
func Human(err error) string {
	if err == nil {
		return ""
	}
	var ue *Error
	if errors.As(err, &ue) {
		return ue.format()
	}
	msg := cleanMsg(err.Error())
	msg = stripCobraUsage(msg)
	return prettify(msg)
}

func cleanMsg(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "Error: ")
	return s
}

func stripCobraUsage(msg string) string {
	if i := strings.Index(msg, "\nUsage:"); i >= 0 {
		return strings.TrimSpace(msg[:i])
	}
	if i := strings.Index(msg, "\n\nUsage:"); i >= 0 {
		return strings.TrimSpace(msg[:i])
	}
	return msg
}

func prettify(msg string) string {
	switch {
	case strings.HasPrefix(msg, "schema validation:"):
		detail := strings.TrimSpace(strings.TrimPrefix(msg, "schema validation:"))
		return fmt.Sprintf("Config file failed schema validation.\n  • %s\n\nWhat you can do:\n  Fix the field noted above in your YAML, or run `rtifact validate --explain` after fixing parse errors.", detail)
	case strings.HasPrefix(msg, "yaml:"):
		detail := strings.TrimSpace(strings.TrimPrefix(msg, "yaml:"))
		return fmt.Sprintf("Config file is not valid YAML.\n  • %s\n\nWhat you can do:\n  Check indentation and quoting in your config file.", detail)
	case strings.Contains(msg, "schema_version must be"):
		return fmt.Sprintf("Unsupported config schema version.\n  • %s\n\nWhat you can do:\n  Set schema_version: 1 at the top of your YAML.", msg)
	case strings.HasPrefix(msg, "features."):
		id := activeProfileID()
		return fmt.Sprintf("Invalid or unsupported feature flags.\n  • %s\n\nWhat you can do:\n  See profiles/%s/features.yaml for allowed features and dependencies.", msg, id)
	case strings.HasPrefix(msg, "stack "):
		return fmt.Sprintf("OpenTofu operation failed.\n  • %s\n\nWhat you can do:\n  Scroll up for provider messages, fix the reported issue, then re-run plan/apply for that stack only (--stack).", msg)
	case strings.HasPrefix(msg, "unknown --stack"):
		return fmt.Sprintf("Unknown stack name.\n  • %s\n\nWhat you can do:\n  Use --stack with one of: state-backend, core-network, data-platform, app-platform, network-edge, ai-integrations, all.", msg)
	default:
		return msg
	}
}
