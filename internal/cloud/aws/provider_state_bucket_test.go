package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cipherops-io/rtifact-platform-installer/internal/cloud"
)

func moduleStyleBucket(prefix, accountID string) string {
	base := strings.ToLower(strings.ReplaceAll(prefix, "_", "-"))
	if len(base) > 45 {
		base = base[:45]
	}
	return fmt.Sprintf("%s-tfstate-%s", base, accountID)
}

func TestStateBackendBucketMatchesStateBackendModule(t *testing.T) {
	spec := cloud.StateSpec{
		NamePrefix: "rtfc-sbx2",
		Account:    cloud.Account{AccountID: "635702036154"},
	}
	bucket, _ := stateBackendNames(spec)
	want := moduleStyleBucket("rtfc-sbx2", "635702036154")
	if bucket != want {
		t.Fatalf("bucket %q != module default %q", bucket, want)
	}
}

func TestStateBackendBucketUsesYAMLOverride(t *testing.T) {
	spec := cloud.StateSpec{
		NamePrefix:      "rtfc-sbx2",
		Account:         cloud.Account{AccountID: "635702036154"},
		StateBucketName: "rtfc-sbx2-tfstate-635702036154",
	}
	bucket, _ := stateBackendNames(spec)
	if bucket != "rtfc-sbx2-tfstate-635702036154" {
		t.Fatalf("bucket %q", bucket)
	}
}