package inventory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/helmclient"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/plan"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	SystemNS        = "rtifact-system"
	ConfigMapPrefix = "rtifact-deploy-"
	// DataKey is the ConfigMap key holding the inventory JSON.
	DataKey = "inventory.json"
	// SchemaVersion is the current inventory schema.
	SchemaVersion = 2
)

// Entry is one managed release in inventory (schema v2).
type Entry = plan.Entry

// Document is the v2 on-disk shape.
type Document struct {
	SchemaVersion int     `json:"schema_version"`
	Entries       []Entry `json:"entries"`
}

// legacyEntry is the v1 shape: no digests of the chart, no requires, no timestamp.
type legacyEntry struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	Chart        string `json:"chart"`
	Version      string `json:"version"`
	ValuesDigest string `json:"values_digest"`
	Protect      bool   `json:"protect,omitempty"`
}

func (l legacyEntry) entry() Entry {
	return Entry{
		Name: l.Name, Namespace: l.Namespace, Chart: l.Chart, Version: l.Version,
		ValuesDigest: l.ValuesDigest, Protect: l.Protect,
	}
}

// Parse decodes inventory JSON. A bare array, or an object without
// schema_version (or with schema_version 1), is the v1 format and is converted
// to v2 entries. Newer versions are rejected.
func Parse(raw string) ([]Entry, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	if raw[0] == '[' {
		var old []legacyEntry
		if err := json.Unmarshal([]byte(raw), &old); err != nil {
			return nil, fmt.Errorf("inventory: %w", err)
		}
		return validate(convertLegacy(old))
	}
	var head struct {
		SchemaVersion int `json:"schema_version"`
	}
	if err := json.Unmarshal([]byte(raw), &head); err != nil {
		return nil, fmt.Errorf("inventory: %w", err)
	}
	switch {
	case head.SchemaVersion > SchemaVersion:
		return nil, fmt.Errorf("inventory: schema_version %d is newer than this binary supports (%d); upgrade rtifact",
			head.SchemaVersion, SchemaVersion)
	case head.SchemaVersion < 0:
		return nil, fmt.Errorf("inventory: invalid schema_version %d", head.SchemaVersion)
	case head.SchemaVersion == SchemaVersion:
		var doc Document
		if err := strictDecode(raw, &doc); err != nil {
			return nil, fmt.Errorf("inventory: %w", err)
		}
		return validate(doc.Entries)
	default: // missing or 1: v1 object
		var old struct {
			Entries []legacyEntry `json:"entries"`
		}
		if err := json.Unmarshal([]byte(raw), &old); err != nil {
			return nil, fmt.Errorf("inventory: %w", err)
		}
		return validate(convertLegacy(old.Entries))
	}
}

func strictDecode(raw string, v any) error {
	dec := json.NewDecoder(bytes.NewReader([]byte(raw)))
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func convertLegacy(old []legacyEntry) []Entry {
	out := make([]Entry, len(old))
	for i, l := range old {
		out[i] = l.entry()
	}
	return out
}

func validate(entries []Entry) ([]Entry, error) {
	seen := map[string]bool{}
	for i, e := range entries {
		if e.Name == "" || e.Namespace == "" {
			return nil, fmt.Errorf("inventory: entry %d: name and namespace are required", i)
		}
		k := e.Namespace + "/" + e.Name
		if seen[k] {
			return nil, fmt.Errorf("inventory: duplicate entry %s", k)
		}
		seen[k] = true
	}
	return entries, nil
}

// Marshal encodes entries as a v2 document, sorted by namespace and name.
// Entries without updated_at are stamped with now.
func Marshal(entries []Entry, now time.Time) ([]byte, error) {
	out := append([]Entry{}, entries...)
	for i := range out {
		if out[i].UpdatedAt.IsZero() {
			out[i].UpdatedAt = now.UTC()
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Namespace != out[j].Namespace {
			return out[i].Namespace < out[j].Namespace
		}
		return out[i].Name < out[j].Name
	})
	return json.Marshal(Document{SchemaVersion: SchemaVersion, Entries: out})
}

// Read loads inventory from the cluster ConfigMap (missing CM → empty).
func Read(ctx context.Context, cs kubernetes.Interface, deployment string) ([]Entry, error) {
	name := ConfigMapPrefix + deployment
	cm, err := cs.CoreV1().ConfigMaps(SystemNS).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return Parse(cm.Data[DataKey])
}

// Write stores the full inventory snapshot as schema v2.
func Write(ctx context.Context, cs kubernetes.Interface, deployment string, entries []Entry) error {
	b, err := Marshal(entries, time.Now())
	if err != nil {
		return err
	}
	name := ConfigMapPrefix + deployment
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: SystemNS,
			Labels:    map[string]string{helmclient.LabelDeployment: deployment},
		},
		Data: map[string]string{DataKey: string(b)},
	}
	_, err = cs.CoreV1().ConfigMaps(SystemNS).Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = cs.CoreV1().ConfigMaps(SystemNS).Create(ctx, cm, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return err
	}
	_, err = cs.CoreV1().ConfigMaps(SystemNS).Update(ctx, cm, metav1.UpdateOptions{})
	return err
}
