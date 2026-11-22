package engine

import (
	"fmt"
	"os"
)

func ConfirmApply(layer string) error {
	fmt.Fprintf(os.Stderr, "Apply stack %q? [y/N]: ", layer)
	return readYesNo()
}

func ConfirmDestroy(layer string) error {
	fmt.Fprintf(os.Stderr, "Destroy stack %q? [y/N]: ", layer)
	return readYesNo()
}