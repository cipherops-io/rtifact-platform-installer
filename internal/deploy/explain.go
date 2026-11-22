package deploy

import (
	"fmt"
	"os"
)

// PrintExplain validates then prints a secret-masked summary (for validate --explain).
func PrintExplain(path string) error {
	if err := ValidateFile(path); err != nil {
		return err
	}
	doc, err := LoadFile(path)
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, ExplainSummary(doc))
	return nil
}