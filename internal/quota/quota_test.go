package quota

import (
	"context"
	"strings"
	"testing"
)

func TestFetchUsagesPlaceholderSubscription(t *testing.T) {
	_, err := fetchUsages(context.Background(), "00000000-0000-0000-0000-000000000000", "centralindia")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "--check-azure=false") {
		t.Fatalf("unexpected error: %v", err)
	}
}
