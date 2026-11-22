# Deploy end-to-end scenarios

Test charts used by these scenarios live in `tests/deploy/charts/`.

| Done | ID | Scenario | Expected result |
|---|---|---|---|
| [ ] | S1 | Fresh apply of `ok-a`, `ok-b` | Installed in order; inventory lists both |
| [ ] | S2 | Apply again | “No changes”; no new Helm revision |
| [ ] | S3 | Raise a chart version | Plan shows upgrade and diff; one new revision |
| [ ] | S4 | Change only a value | Plan shows upgrade; other release unchanged |
| [ ] | S5 | Upgrade to `broken`, with a release after it | Rolled back to the previous revision; run stops; later release untouched; non-zero exit |
| [ ] | S6 | `crd` 1 to 2 with `crds: upgrade` | The new field exists in the cluster’s CRD |
| [ ] | S7 | Remove `ok-b` from the file | Only `ok-b` uninstalled; inventory updated |
| [ ] | S8 | Remove `ok-a` while `ok-b` requires it | Refused at validate |
| [ ] | S9 | Remove a `protect: true` release | Refused without `--allow-destroy` |
| [ ] | S10 | Apply a file with an empty releases list | Refused without `--allow-mass-removal`; nothing removed |
| [ ] | S11 | Leave a release in a pending state | Plan reports it with the recovery command |
| [ ] | S12 | Two applies at once | The second stops on the lock and names the holder |
| [ ] | S13 | Kubeconfig for another cluster than `expected_server` | Refused before any cluster read |
| [ ] | S14 | Misspelt key in the file | Rejected by `validate` |
| [ ] | S15 | `pull_secret: true` | Secret exists; password change is picked up on the next apply |
| [ ] | S16 | Lower a version | Refused without `--allow-downgrade` |
| [ ] | S17 | Wrong registry password | Clear message naming the registry and the variable; nothing changed |
| [ ] | S18 | `destroy` | Reverse order; persistent volume claim remains |
| [ ] | S19 | Single-cloud binary | In `build-binary`: deploy validates the example |
