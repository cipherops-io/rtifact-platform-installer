package assets

// ExtractWorkDir is a no-op when using ContentRoot; kept for API compatibility.
func ExtractWorkDir(workDir string) error {
	return nil
}
