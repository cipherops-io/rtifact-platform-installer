package cloudconfig

import (
	"context"
	"fmt"
	"sync"
)

// ProfileCloud supplies cloud-specific merge, validation and preflight hooks.
type ProfileCloud interface {
	TFVarSections() []string
	Normalize(merged map[string]any)
	Validate(merged map[string]any) error
	Preflight(ctx context.Context, merged map[string]any) error
}

var (
	mu       sync.RWMutex
	profiles = map[string]ProfileCloud{}
)

func Register(cloud string, p ProfileCloud) {
	mu.Lock()
	defer mu.Unlock()
	profiles[cloud] = p
}

func Get(cloud string) (ProfileCloud, error) {
	mu.RLock()
	defer mu.RUnlock()
	p, ok := profiles[cloud]
	if !ok {
		return nil, fmt.Errorf("no cloudconfig profile for %q", cloud)
	}
	return p, nil
}