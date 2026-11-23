package deploy

import (
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

const secretDiffPlaceholder = "<sensitive>"

func manifestForDiff(manifest string) (string, error) {
	manifest = strings.TrimSpace(manifest)
	if manifest == "" {
		return "", nil
	}
	dec := yaml.NewDecoder(strings.NewReader(manifest))
	var docs []string
	for {
		var doc map[string]any
		err := dec.Decode(&doc)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if doc == nil {
			continue
		}
		if kind, _ := doc["kind"].(string); kind == "Secret" {
			maskSecretDataFields(doc)
		}
		b, err := yaml.Marshal(doc)
		if err != nil {
			return "", err
		}
		docs = append(docs, strings.TrimSpace(string(b)))
	}
	if len(docs) == 0 {
		return "", nil
	}
	return strings.Join(docs, "\n---\n"), nil
}

func maskSecretDataFields(doc map[string]any) {
	for _, key := range []string{"data", "stringData"} {
		raw, ok := doc[key]
		if !ok || raw == nil {
			continue
		}
		m, ok := raw.(map[string]any)
		if !ok {
			doc[key] = fmt.Sprintf("<%s keys changed>", key)
			continue
		}
		masked := make(map[string]any, len(m))
		for k := range m {
			masked[k] = secretDiffPlaceholder
		}
		doc[key] = masked
	}
}