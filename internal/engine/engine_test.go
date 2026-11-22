package engine

import (
	"context"
	"errors"
	"testing"
)

type stepFunc func(context.Context) error

func (f stepFunc) Run(ctx context.Context) error { return f(ctx) }

func TestRun_Table(t *testing.T) {
	ctx := context.Background()
	errBoom := errors.New("boom")

	tests := []struct {
		name    string
		steps   []Step
		wantErr bool
	}{
		{name: "empty", steps: nil, wantErr: false},
		{name: "all_ok", steps: []Step{
			stepFunc(func(context.Context) error { return nil }),
			stepFunc(func(context.Context) error { return nil }),
		}},
		{name: "stops_on_error", steps: []Step{
			stepFunc(func(context.Context) error { return nil }),
			stepFunc(func(context.Context) error { return errBoom }),
			stepFunc(func(context.Context) error {
				t.Fatal("should not run after error")
				return nil
			}),
		}, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := Run(ctx, tc.steps)
			if tc.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected: %v", err)
			}
		})
	}
}

func TestConfirmIfNeeded_Table(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		ch      Change
		wantNil bool
	}{
		{opts: Options{Yes: true}, ch: Change{Action: "apply", Stack: "core-network"}, wantNil: true},
		{opts: Options{Yes: true}, ch: Change{Action: "destroy", Stack: "core-network"}, wantNil: true},
		{opts: Options{}, ch: Change{Action: "plan", Stack: "x"}, wantNil: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "" {
				tc.name = tc.ch.Action
			}
			err := ConfirmIfNeeded(tc.opts, tc.ch)
			if tc.wantNil && err != nil {
				t.Fatalf("unexpected: %v", err)
			}
		})
	}
}