# tofu-stack Terratest

Pattern follows `devops-templates/iac-root-module` (`terra_test.go` / `scenarios_test.go`, `IAC_E2E`, `make iac-test-plan`): `terraform.InitAndPlan` per scenario; optional `IAC_E2E=true` apply/destroy.

Stacks are exercised from the rendered layout in **`rtifact-platform-installer/infra`** (`stacks/aios-infra/<cloud>/<layer>`), which matches cloud-pack output and module paths (`infra/modules/<cloud>/`).

**Default CI (`make iac-test-plan`):** Terratest `InitAndValidate` for every cloud × layer (no cloud credentials).

**Optional plan (`TERRATEST_PLAN=true make iac-test-plan-full`):** runs `InitAndPlan` as well (requires real AWS/Azure/GCP credentials).

```bash
# From boilerplates/devops-templates/tofu-stack
make iac-test-plan

# Full apply/destroy (requires cloud credentials; state-backend layer)
make iac-test-e2e
```

Override infra root:

```bash
export TERRATEST_INFRA_ROOT=/path/to/rtifact-platform-installer/infra
```