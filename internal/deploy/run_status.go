package deploy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/helmclient"
	"github.com/cipherops-io/rtifact-platform-installer/internal/deploy/ocireg"
)

// Status prints per-release wanted vs live versions.
func Status(ctx context.Context, o RunOptions, asJSON bool) error {
	doc, disc, kcfg, err := loadRuntime(o)
	if err != nil {
		return err
	}
	if !asJSON {
		printBanner(o.out(), doc, disc, kcfg)
	}
	regCfg := registryFromDoc(doc)
	cred, err := ocireg.LoadAuth(regCfg)
	if err != nil {
		return err
	}
	reg, err := ocireg.NewClient(regCfg, cred)
	if err != nil {
		return err
	}
	cs, err := helmclient.New(helmclient.Options{
		Kube: kcfg, Deployment: DeploymentID(doc), RegistryClient: reg,
		Defaults: helmclient.Defaults{
			Timeout: doc.Defaults.Timeout, Wait: doc.Defaults.Wait,
			RollbackOnFailure: doc.Defaults.RollbackOnFailure, HistoryMax: doc.Defaults.HistoryMax,
		},
		Registry: helmclient.RegistryOpts{PlainHTTP: doc.Registry.PlainHTTP},
	})
	if err != nil {
		return err
	}
	want, err := buildWanted(o.FilePath, doc, disc.OutputsPath)
	if err != nil {
		return err
	}
	live, err := cs.Observe(ctx)
	if err != nil {
		return err
	}
	liveBy := map[string]string{}
	stateBy := map[string]string{}
	for _, l := range live {
		liveBy[l.Name] = l.Version
		stateBy[l.Name] = l.Status
	}
	type row struct {
		Name, Namespace, Chart, Wanted, Deployed, Status string
	}
	var rows []row
	for _, w := range want {
		rows = append(rows, row{
			Name: w.Name, Namespace: w.Namespace, Chart: w.Chart,
			Wanted: w.Version, Deployed: liveBy[w.Name], Status: stateBy[w.Name],
		})
	}
	if asJSON {
		return json.NewEncoder(o.out()).Encode(rows)
	}
	for _, r := range rows {
		fmt.Fprintf(o.out(), "%s/%s  chart=%s  wanted=%s  deployed=%s  status=%s\n",
			r.Namespace, r.Name, r.Chart, r.Wanted, r.Deployed, r.Status)
	}
	return nil
}

// SetWriters allows CLI to redirect status output.
func (o *RunOptions) SetWriters(out, err io.Writer) {
	o.Stdout = out
	o.Stderr = err
}