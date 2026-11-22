package engine

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/usererr"
	"github.com/mattn/go-isatty"
)

// ErrAborted is returned when the user declines or aborts a confirmation prompt.
var ErrAborted = errors.New("aborted")

// ConfirmDeployApply gates deploy apply for installs and upgrades (not removals).
func ConfirmDeployApply(yes bool) error {
	if yes {
		return nil
	}
	if err := requireInteractiveStdin(); err != nil {
		return err
	}
	fmt.Fprint(os.Stderr, "Apply deploy changes? [y/N]: ")
	return readYesNo()
}

// ConfirmDeployRemovals gates uninstall rows shown under Removals in the plan.
func ConfirmDeployRemovals(yes bool) error {
	if yes {
		return nil
	}
	if err := requireInteractiveStdin(); err != nil {
		return err
	}
	fmt.Fprint(os.Stderr, "Remove the releases listed under Removals? [y/N]: ")
	return readYesNo()
}

// ConfirmDeployDestroy gates full destroy; the operator must type the deployment name.
func ConfirmDeployDestroy(yes bool, deployment string) error {
	if yes {
		return nil
	}
	if err := requireInteractiveStdin(); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Destroy all releases for deployment %q.\n", deployment)
	fmt.Fprintf(os.Stderr, "Type the deployment name to confirm: ")
	return readTypedDeployment(deployment)
}

// ConfirmDeployRollback gates rollback.
func ConfirmDeployRollback(yes bool, release string) error {
	if yes {
		return nil
	}
	if err := requireInteractiveStdin(); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Rollback release %q? [y/N]: ", release)
	return readYesNo()
}

func requireInteractiveStdin() error {
	if isatty.IsTerminal(os.Stdin.Fd()) {
		return nil
	}
	return usererr.WithHint(
		"Cannot prompt for confirmation on non-interactive stdin.",
		"Re-run with --yes, or attach a terminal.",
	)
}

func readYesNo() error {
	sc := bufio.NewScanner(os.Stdin)
	if !sc.Scan() {
		return ErrAborted
	}
	ans := strings.ToLower(strings.TrimSpace(sc.Text()))
	if ans != "y" && ans != "yes" {
		return ErrAborted
	}
	return nil
}

func readTypedDeployment(want string) error {
	sc := bufio.NewScanner(os.Stdin)
	if !sc.Scan() {
		return ErrAborted
	}
	if strings.TrimSpace(sc.Text()) != want {
		return ErrAborted
	}
	return nil
}