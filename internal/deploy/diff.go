package deploy

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/helmclient"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/ocireg"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/plan"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/redact"
	"github.com/pmezard/go-difflib/difflib"
)

func unifiedDiff(fromName, toName, before, after string) string {
	before = strings.TrimSpace(before)
	after = strings.TrimSpace(after)
	if before == after {
		return ""
	}
	diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(before + "\n"),
		B:        difflib.SplitLines(after + "\n"),
		FromFile: fromName,
		ToFile:   toName,
		Context:  3,
	})
	if err != nil {
		return fmt.Sprintf("(diff unavailable: %v)", err)
	}
	return strings.TrimSpace(diff)
}

func printManifestDiff(w io.Writer, rd *redact.Redactor, ns, name, live, planned string) {
	d := unifiedDiff(ns+"/"+name+" (live)", ns+"/"+name+" (planned)", live, planned)
	if d == "" {
		return
	}
	fmt.Fprintf(w, "\n--- %s/%s ---\n%s\n", ns, name, MaskedPrint(rd, d))
}

type planDiffOpts struct {
	out, stderr         io.Writer
	debug               bool
	cs                  *helmclient.Client
	doc                 *Document
	res                 plan.Result
	regCfg              ocireg.Registry
	resolved            map[string]map[string]any
	rd                  *redact.Redactor
}

func printPlanManifestDiffs(ctx context.Context, o planDiffOpts) error {
	relByName := map[string]Release{}
	for _, rel := range o.doc.Releases {
		relByName[rel.Name] = rel
	}
	for _, row := range o.res.Rows {
		if row.Action != plan.ActionUpgrade && row.Action != plan.ActionInstall {
			continue
		}
		rel, ok := relByName[row.Name]
		if !ok {
			continue
		}
		values, ok := o.resolved[rel.Name]
		if !ok {
			continue
		}
		ref, err := ocireg.ChartRef(o.regCfg, rel.Chart)
		if err != nil {
			return err
		}
		planned, err := o.cs.PlannedManifest(ctx, helmclient.ReleaseRequest{
			Name: rel.Name, Namespace: rel.Namespace, Chart: rel.Chart, Version: rel.Version,
		}, ref, rel.Version, values, row.Action == plan.ActionUpgrade)
		if err != nil {
			return err
		}
		planned, err = manifestForDiff(planned)
		if err != nil {
			return err
		}
		planned = MaskedPrint(o.rd, planned)
		var live string
		if row.Action == plan.ActionUpgrade {
			live, err = o.cs.ReleaseManifest(ctx, rel.Name, rel.Namespace)
			if err != nil {
				return err
			}
			live, err = manifestForDiff(live)
			if err != nil {
				return err
			}
			live = MaskedPrint(o.rd, live)
		}
		printManifestDiff(o.out, o.rd, row.Namespace, row.Name, live, planned)
		if o.debug {
			fmt.Fprintln(o.stderr, MaskedPrint(o.rd, fmt.Sprintf(
				"debug: planned manifest for %s/%s (%d bytes, secrets masked)",
				row.Namespace, row.Name, len(planned))))
		}
	}
	return nil
}