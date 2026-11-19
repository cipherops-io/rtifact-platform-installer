package packaged

import (
	"embed"
	"io/fs"
)

// Content is synced from repo root by build/scripts/package-content.sh.

//go:embed placeholder.txt
var minimal embed.FS

//go:embed all:profiles all:infra all:schemas all:build/embed
var full embed.FS

// FS returns embedded packaged content when present, otherwise minimal placeholder only.
func FS() fs.FS {
	if entries, err := fs.ReadDir(full, "profiles"); err == nil && len(entries) > 0 {
		return full
	}
	return minimal
}