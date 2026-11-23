# Deploy end-to-end scenarios

Test charts: `tests/deploy/charts/`. `make deploy-e2e` (`build/scripts/deploy-e2e.sh`) starts an authenticated Zot registry and a kind cluster, pushes the charts, runs `go test -tags e2e -count=1 ./tests/deploy/...`, then runs `rehearsal.sh` against a freshly built binary. The CI job `deploy-e2e` in `.github/workflows/ci.yml` is the record of whether these pass; this table lists what each test asserts.

`make deploy-unit` runs the cluster-free tests (Kubernetes `fake` clientset and `helmclient/helmfake`).

### Plan exit contract

| Mode | Exit 0 | Exit 1 | Exit 2 |
|------|--------|--------|--------|
| default | no pending changes | refused rows or error | — |
| `--detailed-exitcode` | no pending changes | refused rows or error | pending changes and no blocking refusal |

Refusal always wins over exit 2 (`internal/deploy/plan_exit_test.go`). An interrupted `apply`, `destroy` or `rollback` exits 130.

### Isolation

Each test writes its own deploy file: `environment` and `name_prefix` are derived from the test name, so every test has its own deployment ID (`customer.code-environment`), inventory ConfigMap, lock Lease and namespaces. Cleanup runs `destroy` and fails the test if it errors.

Helm assertions use the deploy Helm SDK client (`helmclient.Client.Lookup`), not the Helm CLI, so cluster checks match production Helm v4.3.2.

| ID | Scenario | Test function | Assertions |
|---|---|---|---|
| S1 | Fresh apply `ok-a`, `ok-b` | `TestE2E_S1_FreshApply` | Helm rev 1 each; inventory lists both |
| S2 | Re-apply | `TestE2E_S2_ReapplyNoChanges` | No new Helm revision; plan exit 0 |
| S3 | Version bump | `TestE2E_S3_VersionBump` | Plan exit 2; upgrade text; rev +1 on apply |
| S4 | Value-only change | `TestE2E_S4_ValueChange` | Plan exit 2; ok-a rev +1; ok-b unchanged |
| S5 | `broken` rollback | `TestE2E_S5_BrokenRollback` | Apply error; ok-b rev unchanged; inventory ok-a, ok-b |
| S6 | CRD 1→2 `crds: upgrade` | `TestE2E_S6_CRDUpgrade` | CRD schema gains `color` |
| S7 | Remove `ok-b` | `TestE2E_S7_RemoveRelease` | ok-b uninstalled; inventory only ok-a |
| S8 | Remove required `ok-a` | `TestValidate_removeRequiredDependency_S8` | Validate refuses (unit) |
| S9 | Remove `protect: true` | `TestE2E_S9_ProtectedRefusal` | Apply refused; inventory unchanged |
| S10 | Empty releases list | `TestE2E_S10_MassRemovalGuard` | Plan exit 1 (refusal); apply refused; releases remain |
| S11 | Pending + recovery | `TestE2E_S11_StuckReportsRecovery` | Plan shows rollback command; `deploy rollback` leaves `deployed` |
| S12 | Lock held by another run | `TestE2E_S12_DeployLock` | Apply refused naming the holder; nothing installed; apply succeeds after release; Lease gone afterwards |
| S13 | Wrong `expected_server` | `TestE2E_S13_IdentityGuard` | Refused before cluster writes; empty inventory |
| S14 | Invalid values file | `TestE2E_S14_InvalidValuesFile` / `TestValidateFile_unknownKey` | Plan and apply fail; validate rejects typo (unit) |
| S15 | Pull secret refresh | `TestE2E_S15_PullSecretRefresh` | Secret digest changes after password change |
| S16 | Downgrade guard | `TestE2E_S16_DowngradeGuard` | Plan exit 1 (refusal); apply refused; `--allow-downgrade` applies |
| S17 | Bad registry password | `TestE2E_S17_BadRegistryPassword` | Error names registry; no Helm releases |
| S18 | Destroy keeps PVC | `TestE2E_S18_DestroyKeepsPVC` | Release removed; PVC remains |
| S19 | Single-cloud binary | `TestDeployValidateExample` / `build-binary` workflow | `deploy validate` on examples in CI smoke |

### Extras

| Test function | Purpose |
|---|---|
| `TestE2E_Canary_SecretNotInPlanOutput` | Resolved secret never appears in apply, plan, status, or destroy output (with `--debug`) |
| `TestE2E_EnvOnlyChangeTriggersUpgrade` | Changing only a `${env.*}` value makes plan exit 2 and apply upgrade |
| `TestE2E_Adopt_UnmanagedRelease` | Refused without `--adopt` (exit 1); plan + apply with it |
| `TestE2E_Interrupt_CancelApply` | Cancel while a never-ready release is `pending-install`: `ErrInterrupted`, CLI exit 130, Lease released, inventory keeps only completed releases |
| `TestE2E_RollbackHealthyUpgrade` | Rollback after a successful upgrade returns to the superseded revision (rev 3, `deployed`) instead of uninstalling |

### Unit coverage without a cluster

| Area | Tests |
|---|---|
| Destroy resume | `TestDestroyRun_resumesAfterMidListFailure` — failed uninstall keeps only un-removed entries; rerun finishes |
| Purge scope | `TestPurge_*` — sibling release PVCs and pre-existing namespaces survive |
| Rollback decisions | `TestRollbackRelease`, `helmclient.TestPreviousDeployedRevision`, `TestFailedWithoutDeployed` (Helm memory storage) |
| Lock | `lock.TestAcquire_*`, `TestRelease_deletesOwnLeaseOnly`, `TestRenew_failureReachesCallback` |
| Status states | `TestMergeStatusRows_states` |

### Rehearsal (`rehearsal.sh`)

Operator journey on the built binary: validate → plan (exit 2, JSON `schema_version`) → invalid `-o` rejected → apply → plan exit 0 → status (`deployed`) → upgrade → rollback (`drifted`, still installed) → SIGINT during a slow install (exit 130, lock released) → destroy → status (`missing`).

### `status -o json`

`{"schema_version": 1, "deployment": "...", "releases": [...]}`; each release has `name`, `namespace`, `chart`, `wanted`, `deployed`, `status` (Helm) and `state`: `deployed`, `drifted`, `missing`, `pending`, `failed`, `unmanaged` or `orphaned`.
