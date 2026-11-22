package deploy

import (
	"reflect"
	"testing"
)

func TestMergeValMaps_deep(t *testing.T) {
	tests := []struct {
		name    string
		base    map[string]any
		overlay map[string]any
		want    map[string]any
	}{
		{
			name:    "overlay replaces scalar",
			base:    map[string]any{"a": 1},
			overlay: map[string]any{"a": 2},
			want:    map[string]any{"a": 2},
		},
		{
			name: "nested maps merge",
			base: map[string]any{
				"config": map[string]any{"replicas": 1, "image": "v1"},
			},
			overlay: map[string]any{
				"config": map[string]any{"replicas": 3},
			},
			want: map[string]any{
				"config": map[string]any{"replicas": 3, "image": "v1"},
			},
		},
		{
			name: "overlay replaces map with scalar",
			base: map[string]any{
				"config": map[string]any{"replicas": 1},
			},
			overlay: map[string]any{"config": "inline"},
			want:    map[string]any{"config": "inline"},
		},
		{
			name: "multiple keys",
			base: map[string]any{
				"keep":  "x",
				"merge": map[string]any{"a": 1},
			},
			overlay: map[string]any{
				"new":   true,
				"merge": map[string]any{"b": 2},
			},
			want: map[string]any{
				"keep":  "x",
				"new":   true,
				"merge": map[string]any{"a": 1, "b": 2},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeValMaps(tc.base, tc.overlay)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %#v want %#v", got, tc.want)
			}
		})
	}
}