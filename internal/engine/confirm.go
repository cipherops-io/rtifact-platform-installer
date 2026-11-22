package engine

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ConfirmApply(layer string) error {
	fmt.Fprintf(os.Stderr, "Apply stack %q? [y/N]: ", layer)
	return readYesNo()
}

func ConfirmDestroy(layer string) error {
	fmt.Fprintf(os.Stderr, "Destroy stack %q? [y/N]: ", layer)
	return readYesNo()
}

func readYesNo() error {
	sc := bufio.NewScanner(os.Stdin)
	if !sc.Scan() {
		return fmt.Errorf("aborted")
	}
	ans := strings.ToLower(strings.TrimSpace(sc.Text()))
	if ans != "y" && ans != "yes" {
		return fmt.Errorf("aborted")
	}
	return nil
}