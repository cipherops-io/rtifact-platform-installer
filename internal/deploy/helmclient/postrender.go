package helmclient

import (
	"bytes"
	"strings"

	"helm.sh/helm/v4/pkg/postrenderer"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

const pvcKeepAnnotation = "helm.sh/resource-policy"
const pvcKeepValue = "keep"

var decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

// pvcKeepPostRenderer marks persistent volume claims so uninstall leaves data (S18).
type pvcKeepPostRenderer struct{}

func (pvcKeepPostRenderer) Run(rendered *bytes.Buffer) (*bytes.Buffer, error) {
	var docs []string
	for _, part := range strings.Split(rendered.String(), "\n---") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		docs = append(docs, part)
	}
	out := &bytes.Buffer{}
	for i, doc := range docs {
		obj := &unstructured.Unstructured{}
		if _, _, err := decUnstructured.Decode([]byte(doc), nil, obj); err == nil && obj.GetKind() == "PersistentVolumeClaim" {
			ann := obj.GetAnnotations()
			if ann == nil {
				ann = map[string]string{}
			}
			if ann[pvcKeepAnnotation] == "" {
				ann[pvcKeepAnnotation] = pvcKeepValue
			}
			obj.SetAnnotations(ann)
			b, err := runtime.Encode(unstructured.UnstructuredJSONScheme, obj)
			if err != nil {
				return nil, err
			}
			doc = string(b)
		}
		if i > 0 {
			out.WriteString("\n---\n")
		}
		out.WriteString(doc)
	}
	return out, nil
}

func pvcKeepRenderer() postrenderer.PostRenderer {
	return pvcKeepPostRenderer{}
}