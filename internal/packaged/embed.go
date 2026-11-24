package packaged

import (
	"embed"
	"io/fs"
	"runtime"
)

// Content is synced from repo root by build/scripts/package-content.sh.

//go:embed placeholder.txt
var minimal embed.FS

//go:embed all:profiles all:infra all:schemas all:examples all:build/embed
var full embed.FS

// Bundled reports whether this binary embeds OpenTofu for the platform it runs on.
func Bundled() bool {
	_, err := fs.Stat(FS(), "build/embed/tofu/"+runtime.GOOS+"_"+runtime.GOARCH+"/tofu")
	return err == nil
}

// FS returns embedded packaged content when present, otherwise minimal placeholder only.
func FS() fs.FS {
	if entries, err := fs.ReadDir(full, "profiles"); err == nil && len(entries) > 0 {
		return full
	}
	return minimal
}
