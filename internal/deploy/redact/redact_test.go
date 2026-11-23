package redact

import "testing"

func TestText_masksAllRecordedLiterals(t *testing.T) {
	r := New()
	r.Record("s3cret")
	r.Record("tok-123")
	got := r.Text("user=admin pass=s3cret token=tok-123 again=s3cret")
	want := "user=admin pass=*** token=*** again=***"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestText_longerSecretMaskedWhole(t *testing.T) {
	r := New()
	r.Record("abc")
	r.Record("abcdef")
	if got := r.Text("x=abcdef y=abc"); got != "x=*** y=***" {
		t.Fatalf("prefix secret leaked the tail of a longer one: %q", got)
	}
}

func TestRecord_ignoresEmptyAndDuplicates(t *testing.T) {
	r := New()
	r.Record("")
	r.Record("k")
	r.Record("k")
	if len(r.literals) != 1 {
		t.Fatalf("literals: %q", r.literals)
	}
	if got := r.Text("plain"); got != "plain" {
		t.Fatalf("empty literal must not alter text: %q", got)
	}
}
