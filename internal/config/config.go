package config

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"

	"github.com/infracost/infracost/internal/logging"
)

// Project defines a specific terraform project config. This can be used
// specify per folder/project configurations so that users don't have
// to provide flags every run. Fields are documented below. More info
// is outlined here: https://www.infracost.io/config-file
type Project struct {
	// Path to the Terraform directory or JSON/plan file.
	// A path can be repeated with different parameters, e.g. for multiple workspaces.
	Path string `yaml:"path,omitempty" ignored:"true"`
	// ExcludePaths defines a list of directories that the provider should ignore.
	ExcludePaths []string `yaml:"exclude_paths,omitempty" ignored:"true"`
	// Name is a user defined name for the project
	Name string `yaml:"name,omitempty" ignored:"true"`
	// TerraformVarFiles is any var files that are to be used with the project.
	TerraformVarFiles []string `yaml:"terraform_var_files"`
	// TerraformVars is a slice of input vars that are to be used with the project.
	TerraformVars map[string]string `yaml:"terraform_vars"`
	// TerraformForceCLI will run a project by calling out to the terraform/terragrunt binary to generate a plan JSON file.
	TerraformForceCLI bool `yaml:"terraform_force_cli,omitempty"`
	// TerraformPlanFlags are flags to pass to terraform plan with Terraform directory paths
	TerraformPlanFlags string `yaml:"terraform_plan_flags,omitempty" ignored:"true"`
	// TerraformInitFlags are flags to pass to terraform init
	TerraformInitFlags string `yaml:"terraform_init_flags,omitempty" ignored:"true"`
	// TerraformBinary is an optional field used to change the path to the terraform or terragrunt binary
	TerraformBinary string `yaml:"terraform_binary,omitempty" envconfig:"TERRAFORM_BINARY"`
	// TerraformWorkspace is an optional field used to set the Terraform workspace
	TerraformWorkspace string `yaml:"terraform_workspace,omitempty" envconfig:"TERRAFORM_WORKSPACE"`
	// TerraformCloudHost is used to override the default app.terraform.io backend host. Only applicable for
	// terraform cloud/enterprise users.
	TerraformCloudHost string `yaml:"terraform_cloud_host,omitempty" envconfig:"TERRAFORM_CLOUD_HOST"`
	// TerraformCloudToken sets the Team API Token or User API Token so infracost can use it to access the plan.
	// Only applicable for terraform cloud/enterprise users.
	TerraformCloudToken string `yaml:"terraform_cloud_token,omitempty" envconfig:"TERRAFORM_CLOUD_TOKEN"`
	// TerragruntFlags set additional flags that should be passed to terragrunt.
	TerragruntFlags string `envconfig:"TERRAGRUNT_FLAGS"`
	// UsageFile is the full path to usage file that specifies values for usage-based resources
	UsageFile string `yaml:"usage_file,omitempty" ignored:"true"`
	// TerraformUseState sets if the users wants to use the terraform state for infracost ops.
	TerraformUseState bool              `yaml:"terraform_use_state,omitempty" ignored:"true"`
	Env               map[string]string `yaml:"env,omitempty" ignored:"true"`
}

type Config struct {
	Credentials   Credentials
	Configuration Configuration

	Version         string `yaml:"version,omitempty" ignored:"true"`
	LogLevel        string `yaml:"log_level,omitempty" envconfig:"LOG_LEVEL"`
	DebugReport     bool   `ignored:"true"`
	NoColor         bool   `yaml:"no_color,omitempty" envconfig:"NO_COLOR"`
	SkipUpdateCheck bool   `yaml:"skip_update_check,omitempty" envconfig:"SKIP_UPDATE_CHECK"`
	Parallelism     *int   `envconfig:"PARALLELISM"`

	APIKey                    string `envconfig:"API_KEY"`
	PricingAPIEndpoint        string `yaml:"pricing_api_endpoint,omitempty" envconfig:"PRICING_API_ENDPOINT"`
	DefaultPricingAPIEndpoint string `yaml:"default_pricing_api_endpoint,omitempty" envconfig:"DEFAULT_PRICING_API_ENDPOINT"`
	DashboardAPIEndpoint      string `yaml:"dashboard_api_endpoint,omitempty" envconfig:"DASHBOARD_API_ENDPOINT"`
	DashboardEndpoint         string `yaml:"dashboard_endpoint,omitempty" envconfig:"DASHBOARD_ENDPOINT"`
	EnableDashboard           bool   `yaml:"enable_dashboard,omitempty" envconfig:"ENABLE_DASHBOARD"`
	EnableCloud               *bool  `yaml:"enable_cloud,omitempty" envconfig:"ENABLE_CLOUD"`
	DisableHCLParsing         bool   `yaml:"disable_hcl_parsing,omitempty" envconfig:"DISABLE_HCL_PARSING"`

	TLSInsecureSkipVerify *bool  `envconfig:"TLS_INSECURE_SKIP_VERIFY"`
	TLSCACertFile         string `envconfig:"TLS_CA_CERT_FILE"`

	Currency string `envconfig:"CURRENCY"`

	// Org settings
	EnableCloudForOrganization bool

	Projects      []*Project `yaml:"projects" ignored:"true"`
	Format        string     `yaml:"format,omitempty" ignored:"true"`
	ShowSkipped   bool       `yaml:"show_skipped,omitempty" ignored:"true"`
	SyncUsageFile bool       `yaml:"sync_usage_file,omitempty" ignored:"true"`
	Fields        []string   `yaml:"fields,omitempty" ignored:"true"`
	CompareTo     string

	ConfigFilePath string

	NoCache bool `yaml:"fields,omitempty" ignored:"true"`

	SkipErrLine bool

	// for testing
	EventsDisabled       bool
	logWriter            io.Writer
	logDisableTimestamps bool
	disableReportCaller  bool
}

func init() {
	err := loadDotEnv()
	if err != nil {
		log.Fatal(err)
	}
}

func DefaultConfig() *Config {
	return &Config{
		LogLevel: "",
		NoColor:  false,

		DefaultPricingAPIEndpoint: "https://pricing.api.infracost.io",
		PricingAPIEndpoint:        "",
		DashboardAPIEndpoint:      "https://dashboard.api.infracost.io",
		DashboardEndpoint:         "https://dashboard.infracost.io",
		EnableDashboard:           false,

		Projects: []*Project{{}},

		Format: "table",
		Fields: []string{"monthlyQuantity", "unit", "monthlyCost"},

		EventsDisabled: IsTest(),
	}
}

func (c *Config) LoadFromConfigFile(path string) error {
	cfgFile, err := loadConfigFile(path)
	if err != nil {
		return err
	}

	c.Projects = cfgFile.Projects

	// Reload the environment to overwrite any of the config file configs
	err = c.LoadFromEnv()
	if err != nil {
		return err
	}

	return nil
}

// DisableReportCaller sets whether the log entry writes the filename to the log line.
func (c *Config) DisableReportCaller() {
	c.disableReportCaller = true
}

// ReportCaller returns if the log entry writes the filename to the log line.
func (c *Config) ReportCaller() bool {
	level := c.WriteLevel()

	return level == "debug" && !c.disableReportCaller
}

// WriteLevel is the log level that the Logger writes to LogWriter.
func (c *Config) WriteLevel() string {
	if c.DebugReport {
		return logrus.DebugLevel.String()
	}

	return c.LogLevel
}

// LogFields sets the meta fields that are added to any log line entries.
func (c *Config) LogFields() map[string]interface{} {
	if c.WriteLevel() == "debug" {
		f := map[string]interface{}{
			"enable_cloud_org": c.EnableCloudForOrganization,
			"currency":         c.Currency,
			"sync_usage":       c.SyncUsageFile,
		}

		if c.EnableCloud != nil {
			f["enable_cloud_os"] = *c.EnableCloud
		}

		return f
	}

	return nil
}

// SetLogDisableTimestamps sets if logs should contain the timestamp the line is written at.
func (c *Config) SetLogDisableTimestamps(v bool) {
	c.logDisableTimestamps = v
}

// LogFormatter returns the log formatting to be used by a Logger.
func (c *Config) LogFormatter() logrus.Formatter {
	if c.DebugReport {
		return &logrus.JSONFormatter{
			DisableTimestamp: c.logDisableTimestamps,
			PrettyPrint:      true,
		}
	}

	return &logrus.TextFormatter{
		FullTimestamp:    true,
		DisableTimestamp: c.logDisableTimestamps,
		DisableColors:    true,
		SortingFunc: func(keys []string) {
			// Put message at the end
			for i, key := range keys {
				if key == "msg" && i != len(keys)-1 {
					keys[i], keys[len(keys)-1] = keys[len(keys)-1], keys[i]
					break
				}
			}
		},
	}
}

// SetLogWriter sets the io.Writer that the logs should be piped to.
func (c *Config) SetLogWriter(w io.Writer) {
	c.logWriter = w
}

// LogWriter returns the writer the Logger should use to write logs to.
// In most cases this should be stderr, but it can also be a file.
func (c *Config) LogWriter() io.Writer {
	return c.logWriter
}

func (c *Config) LoadFromEnv() error {
	err := c.loadEnvVars()
	if err != nil {
		return err
	}

	err = logging.ConfigureBaseLogger(c)
	if err != nil {
		return err
	}

	err = loadCredentials(c)
	if err != nil {
		return err
	}

	err = loadConfiguration(c)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) loadEnvVars() error {
	err := envconfig.Process("INFRACOST", c)
	if err != nil {
		return err
	}

	for _, project := range c.Projects {
		err = envconfig.Process("INFRACOST", project)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) IsLogging() bool {
	return c.LogLevel != ""
}

func (c *Config) IsSelfHosted() bool {
	return c.PricingAPIEndpoint != "" && c.PricingAPIEndpoint != c.DefaultPricingAPIEndpoint
}

func IsTest() bool {
	return os.Getenv("INFRACOST_ENV") == "test" || strings.HasSuffix(os.Args[0], ".test")
}

func IsDev() bool {
	return os.Getenv("INFRACOST_ENV") == "dev"
}

func loadDotEnv() error {
	envLocalPath := filepath.Join(RootDir(), ".env.local")
	if FileExists(envLocalPath) {
		err := godotenv.Load(envLocalPath)
		if err != nil {
			return err
		}
	}

	if FileExists(".env") {
		err := godotenv.Load()
		if err != nil {
			return err
		}
	}

	return nil
}
