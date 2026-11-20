package terratest

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"

	"github.com/cipherops-io/boilerplates/devops-templates/tofu-stack/terratest/testutil"
)

type TestCase struct {
	Name     string
	Parallel bool
	Options  func(*testing.T) *terraform.Options
	Run      func(*testing.T, *terraform.Options)
}

var testCases []TestCase

func defaultPlanOptions(t *testing.T, cloud, dir string, varFiles ...string) *terraform.Options {
	t.Helper()
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("stack dir %s: %v (set TERRATEST_INFRA_ROOT)", dir, err)
	}
	env := map[string]string{
		"TF_CLI_ARGS_init": "-backend=false -input=false",
		"TF_CLI_ARGS_plan": "-refresh=false",
		"TF_IN_AUTOMATION": "1",
	}
	for k, v := range cloudPlanEnv(cloud) {
		env[k] = v
	}
	return &terraform.Options{
		TerraformDir:    dir,
		TerraformBinary: testutil.IACBinary(),
		NoColor:         true,
		VarFiles:        varFiles,
		EnvVars:         env,
	}
}

func cloudPlanEnv(cloud string) map[string]string {
	switch cloud {
	case "aws":
		return map[string]string{
			"AWS_ACCESS_KEY_ID":         "testing",
			"AWS_SECRET_ACCESS_KEY":     "testing",
			"AWS_EC2_METADATA_DISABLED": "true",
		}
	case "azure":
		return map[string]string{
			"ARM_SUBSCRIPTION_ID":            "00000000-0000-0000-0000-000000000001",
			"ARM_TENANT_ID":                  "00000000-0000-0000-0000-000000000002",
			"ARM_CLIENT_ID":                  "00000000-0000-0000-0000-000000000003",
			"ARM_CLIENT_SECRET":              "terratest-not-a-real-secret",
			"ARM_USE_CLI":                    "false",
			"ARM_USE_MSI":                    "false",
			"ARM_SKIP_PROVIDER_REGISTRATION": "true",
		}
	case "gcp":
		return map[string]string{
			"GOOGLE_PROJECT":     "terratest-demo-project",
			"GOOGLE_CREDENTIALS": `{"type":"service_account","project_id":"terratest-demo-project","private_key_id":"x","private_key":"-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEA0Z3VS5JJcds3xfn/ygWyF8PbnGy0ANx4z4vG+uElZ4hD+nH\n-----END RSA PRIVATE KEY-----\n","client_email":"terratest@terratest-demo-project.iam.gserviceaccount.com","client_id":"1","auth_uri":"https://accounts.google.com/o/oauth2/auth","token_uri":"https://oauth2.googleapis.com/token"}`, //nolint:lll
		}
	default:
		return nil
	}
}

func initAndValidateRun(t *testing.T, opts *terraform.Options) {
	t.Helper()
	terraform.InitAndValidate(t, opts)
}

func initAndPlanRun(t *testing.T, opts *terraform.Options) {
	t.Helper()
	terraform.InitAndPlan(t, opts)
}

func initAndApplyAndDestroyRun(t *testing.T, opts *terraform.Options) {
	t.Helper()
	if os.Getenv("IAC_E2E") != "true" {
		t.Skip("set IAC_E2E=true to run apply/destroy e2e scenarios")
	}
	defer terraform.Destroy(t, opts)
	terraform.InitAndApply(t, opts)
}

func TestScenarios(t *testing.T) {
	if len(testCases) == 0 {
		t.Fatal("no scenarios configured — edit scenarios_test.go")
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Parallel {
				t.Parallel()
			}
			opts := tc.Options(t)
			tc.Run(t, opts)
		})
	}
}