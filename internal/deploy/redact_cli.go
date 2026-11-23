package deploy

import (
	"errors"
	"io"
	"sync/atomic"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/redact"
)

var activeRedactor atomic.Pointer[redact.Redactor]

// SetActiveRedactor records the deploy redactor for the current CLI command (if any).
func SetActiveRedactor(rd *redact.Redactor) {
	activeRedactor.Store(rd)
}

// ActiveRedactor returns the redactor bound for the current deploy command.
func ActiveRedactor() *redact.Redactor {
	return activeRedactor.Load()
}

// redactingWriter masks each Write independently, so a secret split across two
// writes is not caught; deploy output goes through fmt.Fprint* and json.Encoder,
// which write each formatted value in one call.
type redactingWriter struct {
	w  io.Writer
	rd *redact.Redactor
}

func (rw *redactingWriter) Write(p []byte) (int, error) {
	_, err := io.WriteString(rw.w, MaskedPrint(rw.rd, string(p)))
	return len(p), err
}

func (o *RunOptions) withRedactedOutput(rd *redact.Redactor) {
	if rd == nil {
		return
	}
	o.Stdout = &redactingWriter{w: o.out(), rd: rd}
	o.Stderr = &redactingWriter{w: o.err(), rd: rd}
}

type redactedErr struct {
	inner error
	msg   string
}

func (e *redactedErr) Error() string { return e.msg }
func (e *redactedErr) Unwrap() error { return e.inner }

// RedactError masks secret literals recorded on rd in err's message.
func RedactError(rd *redact.Redactor, err error) error {
	if err == nil || rd == nil {
		return err
	}
	masked := MaskedPrint(rd, err.Error())
	if masked == err.Error() {
		return err
	}
	var already *redactedErr
	if errors.As(err, &already) {
		return err
	}
	return &redactedErr{inner: err, msg: masked}
}