package plan

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// ValuesDigest returns a stable hash of release values for change detection.
func ValuesDigest(v map[string]any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}