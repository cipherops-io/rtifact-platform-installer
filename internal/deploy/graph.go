package deploy

import (
	"fmt"
)

// InstallOrder returns release names in requires-respecting install order.
func InstallOrder(releases []Release) ([]string, error) {
	idx := make(map[string]int, len(releases))
	for i, r := range releases {
		idx[r.Name] = i
	}
	graph := make(map[string][]string, len(releases))
	inDegree := make(map[string]int, len(releases))
	for _, r := range releases {
		inDegree[r.Name] = 0
	}
	for _, r := range releases {
		for _, req := range r.Requires {
			if _, ok := idx[req]; !ok {
				return nil, fmt.Errorf("release %q requires unknown %q", r.Name, req)
			}
			graph[req] = append(graph[req], r.Name)
			inDegree[r.Name]++
		}
	}
	var queue []string
	for name, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, name)
		}
	}
	var order []string
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		order = append(order, n)
		for _, dep := range graph[n] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}
	if len(order) != len(releases) {
		return nil, userErrCycle()
	}
	return order, nil
}

// RemoveOrder returns uninstall order (reverse of install).
func RemoveOrder(releases []Release) ([]string, error) {
	order, err := InstallOrder(releases)
	if err != nil {
		return nil, err
	}
	for i, j := 0, len(order)-1; i < j; i, j = i+1, j-1 {
		order[i], order[j] = order[j], order[i]
	}
	return order, nil
}

func userErrCycle() error {
	return fmt.Errorf("requires: dependency cycle detected among releases")
}