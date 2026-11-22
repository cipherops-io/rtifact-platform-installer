package crds

import (
	"path/filepath"
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"helm.sh/helm/v4/pkg/chart/v2/loader"
)

func TestCollectCRDs_fromCrdsDir(t *testing.T) {
	dir := filepath.Join("..", "..", "..", "tests", "deploy", "charts", "crd")
	ch, err := loader.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := collectCRDs(ch)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "widgets.deploytests.rtifact.ai" {
		t.Fatalf("got %#v", got)
	}
}

func TestVersionDropRefusal(t *testing.T) {
	cur := &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: "widgets.example.com"},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{Name: "v1", Served: true, Storage: true},
			},
		},
	}
	next := &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: "widgets.example.com"},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{Name: "v2", Served: true, Storage: true},
			},
		},
	}
	if err := versionDropRefusal(cur, next); err == nil {
		t.Fatal("expected refusal when dropping stored v1")
	}
}