# Deploy end-to-end scenarios

Test charts: `tests/deploy/charts/`. Cluster suite: `go test -tags e2e -count=1 ./tests/deploy/...` via `make deploy-e2e`.

| Done | ID | Scenario | Test function | Assertions |
|---|---|---|---|---|
| [x] | S1 | Fresh apply `ok-a`, `ok-b` | `TestE2E_S1_FreshApply` | Helm rev 1 each; inventory lists both |
| [x] | S2 | Re-apply | `TestE2E_S2_ReapplyNoChanges` | No new Helm revision; plan exit 0 |
| [x] | S3 | Version bump | `TestE2E_S3_VersionBump` | Plan exit 2; upgrade text; rev +1 on apply |
| [x] | S4 | Value-only change | `TestE2E_S4_ValueChange` | Plan exit 2; ok-a rev +1; ok-b unchanged |
| [x] | S5 | `broken` rollback | `TestE2E_S5_BrokenRollback` | Apply error; ok-b rev unchanged; inventory ok-a, ok-b |
| [x] | S6 | CRD 1→2 `crds: upgrade` | `TestE2E_S6_CRDUpgrade` | CRD schema gains `color` |
| [x] | S7 | Remove `ok-b` | `TestE2E_S7_RemoveRelease` | ok-b uninstalled; inventory only ok-a |
| [x] | S8 | Remove required `ok-a` | `TestValidate_removeRequiredDependency_S8` | Validate refuses (unit) |
| [x] | S9 | Remove `protect: true` | `TestE2E_S9_ProtectedRefusal` | Apply refused; inventory unchanged |
| [x] | S10 | Empty releases list | `TestE2E_S10_MassRemovalGuard` | Plan exit 2; apply refused; releases remain |
| [x] | S11 | Pending + recovery | `TestE2E_S11_StuckReportsRecovery` | Plan shows rollback command; `deploy rollback` OK |
| [x] | S12 | Concurrent apply | `TestE2E_S12_DeployLock` | Second apply refused; lock holder named |
| [x] | S13 | Wrong `expected_server` | `TestE2E_S13_IdentityGuard` | Refused before cluster writes; empty inventory |
| [x] | S14 | Invalid values file | `TestE2E_S14_InvalidValuesFile` / `TestValidateFile_unknownKey` | Plan and apply fail; validate rejects typo (unit) |
| [x] | S15 | Pull secret refresh | `TestE2E_S15_PullSecretRefresh` | Secret digest changes after password change |
| [x] | S16 | Downgrade guard | `TestE2E_S16_DowngradeGuard` | Plan exit 2; apply refused; `--allow-downgrade` applies |
| [x] | S17 | Bad registry password | `TestE2E_S17_BadRegistryPassword` | Error names registry; no Helm releases |
| [x] | S18 | Destroy keeps PVC | `TestE2E_S18_DestroyKeepsPVC` | Release removed; PVC remains |
| [x] | S19 | Single-cloud binary | `TestDeployValidateExample` / `build-binary` workflow | `deploy validate` on examples in CI smoke |

### Extras

| Test function | Purpose |
|---|---|
| `TestE2E_Canary_SecretNotInPlanOutput` | Resolved secret never appears in plan output |
| `TestE2E_Adopt_UnmanagedRelease` | `--adopt` plan + apply for pre-existing release |
| `TestE2E_Interrupt_CancelApply` | Cancelled apply returns `ErrInterrupted` |