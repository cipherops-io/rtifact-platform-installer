package deploy

const (
	SchemaVersion = 1
	Kind          = "deploy"
)

// Document is the deploy configuration file (infrastructure YAML is separate).
type Document struct {
	SchemaVersion int              `yaml:"schema_version"`
	Kind          string           `yaml:"kind"`
	Customer      Customer         `yaml:"customer"`
	Environment   string           `yaml:"environment"`
	NamePrefix    string           `yaml:"name_prefix"`
	Cluster       Cluster          `yaml:"cluster"`
	Registry      Registry         `yaml:"registry"`
	InfraOutputs  string           `yaml:"infra_outputs"`
	Defaults      Defaults         `yaml:"defaults"`
	Namespaces    []Namespace      `yaml:"namespaces"`
	Releases      []Release        `yaml:"releases"`
}

type Customer struct {
	Name string `yaml:"name"`
	Code string `yaml:"code"`
}

type Cluster struct {
	Kubeconfig     string `yaml:"kubeconfig"`
	Context        string `yaml:"context"`
	ExpectedServer string `yaml:"expected_server"`
}

type Registry struct {
	Host        string           `yaml:"host"`
	ChartsPath  string           `yaml:"charts_path"`
	CAFile      string           `yaml:"ca_file"`
	PlainHTTP   bool             `yaml:"plain_http"`
	Auth        *RegistryAuth    `yaml:"auth"`
	PullSecret  *PullSecretRef   `yaml:"pull_secret"`
}

type RegistryAuth struct {
	UsernameEnv string `yaml:"username_env"`
	PasswordEnv string `yaml:"password_env"`
}

type PullSecretRef struct {
	Name string `yaml:"name"`
}

type Defaults struct {
	Timeout             string `yaml:"timeout"`
	Wait                bool   `yaml:"wait"`
	RollbackOnFailure   bool   `yaml:"rollback_on_failure"`
	HistoryMax          int    `yaml:"history_max"`
}

type Namespace struct {
	Name                         string            `yaml:"name"`
	PullSecret                   bool              `yaml:"pull_secret"`
	AttachPullSecretToDefaultSA  bool              `yaml:"attach_pull_secret_to_default_sa"`
	Labels                       map[string]string `yaml:"labels"`
}

type Release struct {
	Name         string         `yaml:"name"`
	Namespace    string         `yaml:"namespace"`
	Chart        string         `yaml:"chart"`
	Version      string         `yaml:"version"`
	CRDs         string         `yaml:"crds"`
	Requires     []string       `yaml:"requires"`
	ValuesFiles  []string       `yaml:"values_files"`
	Values       map[string]any `yaml:"values"`
	Protect      bool           `yaml:"protect"`
}