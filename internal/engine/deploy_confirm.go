package engine

import (
	"fmt"
	"os"
)

// ConfirmDeployApply gates deploy apply (separate from stack apply strings).
func ConfirmDeployApply(yes bool, hasRemovals bool) error {
	if yes {
		return nil
	}
	if hasRemovals {
		fmt.Fprint(os.Stderr, "Apply deploy changes including removals? [y/N]: ")
	} else {
		fmt.Fprint(os.Stderr, "Apply deploy changes? [y/N]: ")
	}
	return readYesNo()
}

// ConfirmDeployDestroy gates full destroy.
func ConfirmDeployDestroy(yes bool, deployment string) error {
	if yes {
		return nil
	}
	fmt.Fprintf(os.Stderr, "Destroy all releases for deployment %q? [y/N]: ", deployment)
	return readYesNo()
}

// ConfirmDeployRollback gates rollback.
func ConfirmDeployRollback(yes bool, release string) error {
	if yes {
		return nil
	}
	fmt.Fprintf(os.Stderr, "Rollback release %q? [y/N]: ", release)
	return readYesNo()
}