package crds

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/kube"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	chart "helm.sh/helm/v4/pkg/chart/v2"
	"helm.sh/helm/v4/pkg/chart/v2/loader"
)

const (
	fieldManager    = "rtifact-deploy"
	establishedWait = 5 * time.Minute
)

// Upgrade applies chart CRDs when crds: upgrade (SSA, Established wait, non-additive refusal).
func Upgrade(ctx context.Context, kcfg kube.Config, mode string, chartData []byte) error {
	if mode != "upgrade" {
		return nil
	}
	if len(chartData) == 0 {
		return fmt.Errorf("crds upgrade: empty chart archive")
	}
	ch, err := loader.LoadArchive(bytes.NewReader(chartData))
	if err != nil {
		return fmt.Errorf("crds upgrade: load chart: %w", err)
	}
	docs, err := collectCRDs(ch)
	if err != nil {
		return err
	}
	if len(docs) == 0 {
		return nil
	}
	ext, err := apiExtensionsClient(kcfg)
	if err != nil {
		return err
	}
	for _, crd := range docs {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := refuseBreaking(ctx, ext, crd); err != nil {
			return err
		}
		if err := applySSA(ctx, ext, crd); err != nil {
			return err
		}
		if err := waitEstablished(ctx, ext, crd.Name); err != nil {
			return err
		}
	}
	return nil
}

func apiExtensionsClient(kcfg kube.Config) (*apiextensionsclient.Clientset, error) {
	restCfg, err := kube.RESTConfig(kcfg)
	if err != nil {
		return nil, err
	}
	return apiextensionsclient.NewForConfig(restCfg)
}

func collectCRDs(ch *chart.Chart) ([]*apiextensionsv1.CustomResourceDefinition, error) {
	var raw [][]byte
	for _, o := range ch.CRDObjects() {
		if o.File != nil && len(o.File.Data) > 0 {
			raw = append(raw, o.File.Data)
		}
	}
	for _, f := range ch.Templates {
		if f == nil || len(f.Data) == 0 {
			continue
		}
		raw = append(raw, f.Data)
	}
	seen := map[string]struct{}{}
	var out []*apiextensionsv1.CustomResourceDefinition
	for _, blob := range raw {
		docs, err := decodeCRDDocuments(blob)
		if err != nil {
			return nil, err
		}
		for _, crd := range docs {
			if crd.Name == "" {
				return nil, fmt.Errorf("CRD manifest missing metadata.name")
			}
			if _, ok := seen[crd.Name]; ok {
				continue
			}
			seen[crd.Name] = struct{}{}
			out = append(out, crd)
		}
	}
	return out, nil
}

func decodeCRDDocuments(data []byte) ([]*apiextensionsv1.CustomResourceDefinition, error) {
	dec := utilyaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 4096)
	var out []*apiextensionsv1.CustomResourceDefinition
	for {
		var doc map[string]any
		err := dec.Decode(&doc)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("decode CRD yaml: %w", err)
		}
		if doc == nil {
			continue
		}
		kind, _ := doc["kind"].(string)
		if kind != "CustomResourceDefinition" {
			continue
		}
		b, err := json.Marshal(doc)
		if err != nil {
			return nil, err
		}
		var crd apiextensionsv1.CustomResourceDefinition
		if err := json.Unmarshal(b, &crd); err != nil {
			return nil, fmt.Errorf("parse CRD: %w", err)
		}
		out = append(out, &crd)
	}
	return out, nil
}

func refuseBreaking(ctx context.Context, ext apiextensionsclient.Interface, next *apiextensionsv1.CustomResourceDefinition) error {
	cur, err := ext.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, next.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return versionDropRefusal(cur, next)
}

func versionDropRefusal(cur, next *apiextensionsv1.CustomResourceDefinition) error {
	newVers := map[string]struct{}{}
	for _, v := range next.Spec.Versions {
		newVers[v.Name] = struct{}{}
	}
	for _, v := range cur.Spec.Versions {
		if _, ok := newVers[v.Name]; ok {
			continue
		}
		if v.Storage {
			return fmt.Errorf(
				"CRD %q upgrade refused: chart drops stored version %q still present in the cluster",
				next.Name, v.Name,
			)
		}
	}
	return nil
}

func applySSA(ctx context.Context, ext apiextensionsclient.Interface, crd *apiextensionsv1.CustomResourceDefinition) error {
	crd.SetManagedFields(nil)
	b, err := json.Marshal(crd)
	if err != nil {
		return err
	}
	force := true
	_, err = ext.ApiextensionsV1().CustomResourceDefinitions().Patch(
		ctx, crd.Name, types.ApplyPatchType, b,
		metav1.PatchOptions{FieldManager: fieldManager, Force: &force},
	)
	return err
}

func waitEstablished(ctx context.Context, ext apiextensionsclient.Interface, name string) error {
	deadline := time.Now().Add(establishedWait)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		crd, err := ext.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		for _, c := range crd.Status.Conditions {
			if c.Type == apiextensionsv1.Established && c.Status == apiextensionsv1.ConditionTrue {
				return nil
			}
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("CRD %q did not become Established within %s", name, establishedWait)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}