package cluster

import "testing"

func TestPvcNamesFromManifest(t *testing.T) {
	m := `---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: data-one
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: st
spec:
  volumeClaimTemplates:
  - metadata:
      name: data-two
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 1Gi
`
	got, err := PVCNamesFromManifest(m)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "data-one" || got[1] != "data-two" {
		t.Fatalf("got %v", got)
	}
}
