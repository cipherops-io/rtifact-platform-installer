package helmclient

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"helm.sh/helm/v4/pkg/postrenderer"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
)

const pvcKeepAnnotation = "helm.sh/resource-policy"
const pvcKeepValue = "keep"

var decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

// pvcKeepPostRenderer marks persistent volume claims so uninstall leaves data (S18).
type pvcKeepPostRenderer struct{}

func (pvcKeepPostRenderer) Run(rendered *bytes.Buffer) (*bytes.Buffer, error) {
	reader := utilyaml.NewYAMLReader(bufio.NewReader(rendered))
	out := &bytes.Buffer{}
	first := true
	for {
		docBytes, err := reader.Read()
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("post-render: read manifest: %w", err)
		}
		if err == io.EOF {
			break
		}
		docBytes = bytes.TrimSpace(docBytes)
		if len(docBytes) == 0 {
			continue
		}
		obj := &unstructured.Unstructured{}
		if _, _, err := decUnstructured.Decode(docBytes, nil, obj); err != nil {
			return nil, fmt.Errorf("post-render: decode manifest: %w", err)
		}
		doc := string(docBytes)
		if obj.GetKind() == "PersistentVolumeClaim" {
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
		if !first {
			out.WriteString("\n---\n")
		}
		first = false
		out.WriteString(doc)
	}
	return out, nil
}

func pvcKeepRenderer() postrenderer.PostRenderer {
	return pvcKeepPostRenderer{}
}