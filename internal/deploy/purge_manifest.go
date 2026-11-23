package deploy

import (
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

func pvcNamesFromManifest(manifest string) ([]string, error) {
	if strings.TrimSpace(manifest) == "" {
		return nil, nil
	}
	dec := yaml.NewDecoder(strings.NewReader(manifest))
	var names []string
	for {
		var doc map[string]any
		err := dec.Decode(&doc)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if doc == nil {
			continue
		}
		kind, _ := doc["kind"].(string)
		meta, _ := doc["metadata"].(map[string]any)
		switch kind {
		case "PersistentVolumeClaim":
			if meta != nil {
				if n, ok := meta["name"].(string); ok && n != "" {
					names = append(names, n)
				}
			}
		case "StatefulSet":
			spec, _ := doc["spec"].(map[string]any)
			if spec == nil {
				continue
			}
			vcts, _ := spec["volumeClaimTemplates"].([]any)
			for _, v := range vcts {
				m, _ := v.(map[string]any)
				if m == nil {
					continue
				}
				md, _ := m["metadata"].(map[string]any)
				if md == nil {
					continue
				}
				if n, ok := md["name"].(string); ok && n != "" {
					names = append(names, n)
				}
			}
		}
	}
	return uniqSorted(names), nil
}

func uniqSorted(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}