package deploy

import (
	"fmt"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/helmclient"
	"github.com/cipherops-io/rtifact-platform-installer/internal/usererr"
)

func helmDefaults(o RunOptions, doc *Document) helmclient.Defaults {
	d := helmclient.Defaults{
		Timeout:           doc.Defaults.Timeout,
		Wait:              doc.Defaults.Wait,
		RollbackOnFailure: doc.Defaults.RollbackOnFailure,
		HistoryMax:        doc.Defaults.HistoryMax,
	}
	if o.Timeout > 0 {
		d.Timeout = o.Timeout.String()
	}
	return d
}

func debugf(o RunOptions, format string, args ...any) {
	if !o.Debug {
		return
	}
	fmt.Fprintf(o.err(), "debug: "+format+"\n", args...)
}

// validateOnlyRelease ensures --only names a release in the file.
func validateOnlyRelease(doc *Document, only string) error {
	if only == "" {
		return nil
	}
	for _, r := range doc.Releases {
		if r.Name == only {
			return nil
		}
	}
	return usererr.WithHint(
		fmt.Sprintf("Unknown release %q in --only.", only),
		"Use a release name from the deploy file.",
	)
}