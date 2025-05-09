package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/infracost/infracost/internal/logging"
)

const InfracostDir = ".infracost"

type AutodetectConfig struct {
	// EnvNames is the list of environment names that we should use to group
	// terraform var files.
	EnvNames []string `yaml:"env_names,omitempty" ignored:"true"`
	// ExcludeDirs is a list of directories that the autodetect should ignore.
	ExcludeDirs []string `yaml:"exclude_dirs,omitempty" ignored:"true"`
	// IncludeDirs is a list of directories that the autodetect should append
	// to the already detected directories.
	IncludeDirs []string `yaml:"include_dirs,omitempty" ignored:"true"`
	// PathOverrides defines paths that should be overridden with specific
	// environment variable grouping.
	PathOverrides []PathOverride `yaml:"path_overrides,omitempty" ignored:"true"`
	// MaxSearchDepth configures the number of folders that Infracost should
	// traverse to detect projects.
	MaxSearchDepth int `yaml:"max_search_depth,omitempty" ignored:"true"`
	// ForceProjectType is used to force the project type to be a specific value.
	// This is useful when autodetect collides with two project types, normally
	// Terragrunt and Terraform and we want to force the project type to be one or
	// the other.
	ForceProjectType string `yaml:"force_project_type,omitempty" ignored:"true"`
	// TerraformVarFileExtensions is a list of suffixes that should be used to group terraform
	// var files. This is useful when there are non-standard terraform var file
	// names which use different extensions.
	TerraformVarFileExtensions []string `yaml:"terraform_var_file_extensions,omitempty" ignored:"true"`
	// PreferFolderNameForEnv tells the autodetect to prefer the folder name over the
	// over a env specified in a tfvars file. For example, given the following
	// folder structure:
	//
	// .
	// ├── qa
	// │   └── dev.tfvars
	// └── staging
	//     └── prod.tfvars
	//
	// If PreferFolderNameForEnv is true, then the autodetect will group the projects
	// by the folder name so the projects will be named "qa" and "staging".
	PreferFolderNameForEnv bool `yaml:"prefer_folder_name_for_env,omitempty" ignored:"true"`
}

type PathOverride struct {
	Path    string   `yaml:"path"`
	Exclude []string `yaml:"exclude"`
	Only    []string `yaml:"only"`
}

// Project defines a specific terraform project config. This can be used
// specify per folder/project configurations so that users don't have
// to provide flags every run. Fields are documented below. More info
// is outlined here: https://www.infracost.io/config-file
type Project struct {
	// ConfigSha can be provided to identify the configuration used for the project
	ConfigSha string `yaml:"config_sha,omitempty"  ignored:"true"`
	// Path to the Terraform directory or JSON/plan file.
	// A path can be repeated with different parameters, e.g. for multiple workspaces.
	Path string `yaml:"path" ignored:"true"`
	// ExcludePaths defines a list of directories that the provider should ignore.
	ExcludePaths []string `yaml:"exclude_paths,omitempty" ignored:"true"`
	// DependencyPaths is a list of any paths that this project depends on. These paths are relative to the
	// config file and NOT the project.
	DependencyPaths []string `yaml:"dependency_paths,omitempty"`
	// IncludeAllPaths tells autodetect to use all folders with valid project files.
	IncludeAllPaths bool `yaml:"include_all_paths,omitempty" ignored:"true"`
	// SkipAutodetect tells autodetect to skip this project.
	SkipAutodetect bool `yaml:"skip_autodetect,omitempty" ignored:"true"`
	// Name is a user defined name for the project
	Name string `yaml:"name,omitempty" ignored:"true"`
	// TerraformVarFiles is any var files that are to be used with the project.
	TerraformVarFiles []string `yaml:"terraform_var_files,omitempty"`
	// TerraformVars is a slice of input vars that are to be used with the project.
	TerraformVars map[string]interface{} `yaml:"terraform_vars,omitempty"`
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
	// TerraformCloudWorkspace is used to override the terraform configuration blocks workspace value.
	TerraformCloudWorkspace string `yaml:"terraform_cloud_workspace,omitempty" envconfig:"TERRAFORM_CLOUD_WORKSPACE"`
	// TerraformCloudOrg is used to override the terraform configuration blocks organization value.
	TerraformCloudOrg string `yaml:"terraform_cloud_org,omitempty" envconfig:"TERRAFORM_CLOUD_ORG"`
	// TerraformCloudHost is used to override the default app.terraform.io backend host. Only applicable for
	// terraform cloud/enterprise users.
	TerraformCloudHost string `yaml:"terraform_cloud_host,omitempty" envconfig:"TERRAFORM_CLOUD_HOST"`
	// TerraformCloudToken sets the Team API Token or User API Token so infracost can use it to access the plan.
	// Only applicable for terraform cloud/enterprise users.
	TerraformCloudToken string `yaml:"terraform_cloud_token,omitempty" envconfig:"TERRAFORM_CLOUD_TOKEN"`
	// SpaceliftAPIKeyEndpoint is the endpoint that the spacelift API client will communicate with.
	SpaceliftAPIKeyEndpoint string `yaml:"spacelift_api_key_endpoint,omitempty" envconfig:"SPACELIFT_API_KEY_ENDPOINT"`
	// SpaceliftAPIKeyID is the spacelift API key ID. This is used in combination
	// with the API key secret to generate a JWT token.
	SpaceliftAPIKeyID string `yaml:"spacelift_api_key_id,omitempty" envconfig:"SPACELIFT_API_KEY_ID"`
	// SpaceliftAPIKeySecret is the spacelift API key secret.This is used in combination
	// with the API key id to generate a JWT token.
	SpaceliftAPIKeySecret string `yaml:"spacelift_api_key_secret,omitempty" envconfig:"SPACELIFT_API_KEY_SECRET"`
	// TerragruntFlags set additional flags that should be passed to terragrunt.
	TerragruntFlags string `yaml:"terragrunt_flags,omitempty" envconfig:"TERRAGRUNT_FLAGS"`
	// UsageFile is the full path to usage file that specifies values for usage-based resources
	UsageFile string `yaml:"usage_file,omitempty" ignored:"true"`
	// TerraformUseState sets if the users wants to use the terraform state for infracost ops.
	TerraformUseState bool              `yaml:"terraform_use_state,omitempty" ignored:"true"`
	Env               map[string]string `yaml:"env,omitempty" ignored:"true"`
	// YorConfigPath is the path to a Yor config file, which we can extract default tags from
	YorConfigPath string `yaml:"yor_config_path,omitempty" ignored:"true"`
	// Metadata is a map of key-value pairs that can be used to store additional information about the project.
	// This is useful for storing flexible project information that needs to be accessed by other parts
	// of the application.
	Metadata map[string]string `yaml:"metadata,omitempty" ignored:"true"`
}

type Config struct {
	Credentials   Credentials
	Configuration Configuration

	Version    string           `yaml:"version,omitempty" ignored:"true"`
	Autodetect AutodetectConfig `yaml:"autodetect,omitempty" ignored:"true"`

	LogLevel        string `yaml:"log_level,omitempty" envconfig:"LOG_LEVEL"`
	DebugReport     bool   `ignored:"true"`
	NoColor         bool   `yaml:"no_color,omitempty" envconfig:"NO_COLOR"`
	SkipUpdateCheck bool   `yaml:"skip_update_check,omitempty" envconfig:"SKIP_UPDATE_CHECK"`
	Parallelism     *int   `envconfig:"PARALLELISM"`

	APIKey                    string `envconfig:"API_KEY"`
	PricingAPIEndpoint        string `yaml:"pricing_api_endpoint,omitempty" envconfig:"PRICING_API_ENDPOINT"`
	PricingCacheDisabled      bool   `yaml:"pricing_cache_disabled" envconfig:"PRICING_CACHE_DISABLED"`
	PricingCacheObjectSize    int    `yaml:"pricing_cache_object_size" envconfig:"PRICING_CACHE_OBJECT_SIZE"`
	DefaultPricingAPIEndpoint string `yaml:"default_pricing_api_endpoint,omitempty" envconfig:"DEFAULT_PRICING_API_ENDPOINT"`
	DashboardAPIEndpoint      string `yaml:"dashboard_api_endpoint,omitempty" envconfig:"DASHBOARD_API_ENDPOINT"`
	DashboardEndpoint         string `yaml:"dashboard_endpoint,omitempty" envconfig:"DASHBOARD_ENDPOINT"`
	UsageAPIEndpoint          string `yaml:"usage_api_endpoint,omitempty" envconfig:"USAGE_API_ENDPOINT"`
	UsageActualCosts          bool   `yaml:"usage_actual_costs,omitempty" envconfig:"USAGE_ACTUAL_COSTS"`
	PolicyV2APIEndpoint       string `yaml:"policy_v2_api_endpoint,omitempty" envconfig:"POLICY_V2_API_ENDPOINT"`
	PoliciesEnabled           bool
	TagPoliciesEnabled        bool
	EnableDashboard           bool  `yaml:"enable_dashboard,omitempty" envconfig:"ENABLE_DASHBOARD"`
	EnableCloud               *bool `yaml:"enable_cloud,omitempty" envconfig:"ENABLE_CLOUD"`
	EnableCloudUpload         *bool `yaml:"enable_cloud_upload,omitempty" envconfig:"ENABLE_CLOUD_UPLOAD"`
	DisableHCLParsing         bool  `yaml:"disable_hcl_parsing,omitempty" envconfig:"DISABLE_HCL_PARSING"`
	GraphEvaluator            bool  `yaml:"graph_evaluator,omitempty" envconfig:"GRAPH_EVALUATOR"`

	TLSInsecureSkipVerify *bool  `envconfig:"TLS_INSECURE_SKIP_VERIFY"`
	TLSCACertFile         string `envconfig:"TLS_CA_CERT_FILE"`

	Currency       string `envconfig:"CURRENCY"`
	CurrencyFormat string `envconfig:"CURRENCY_FORMAT"`

	AWSOverrideRegion    string `envconfig:"AWS_OVERRIDE_REGION"`
	AzureOverrideRegion  string `envconfig:"AZURE_OVERRIDE_REGION"`
	GoogleOverrideRegion string `envconfig:"GOOGLE_OVERRIDE_REGION"`

	// TerraformSourceMap replaces any source URL with the provided value.
	TerraformSourceMap TerraformSourceMap `envconfig:"TERRAFORM_SOURCE_MAP"`

	// TerraformSourceMapRegex is a more flexible source mapping that supports regex patterns.
	TerraformSourceMapRegex TerraformSourceMapRegex `yaml:"terraform_source_map,omitempty"`

	S3ModuleCacheRegion  string `envconfig:"S3_MODULE_CACHE_REGION"`
	S3ModuleCacheBucket  string `envconfig:"S3_MODULE_CACHE_BUCKET"`
	S3ModuleCachePrefix  string `envconfig:"S3_MODULE_CACHE_PREFIX"`
	S3ModuleCachePrivate bool   `envconfig:"S3_MODULE_CACHE_PRIVATE, default=false"`

	// metrics dump path
	MetricsPath string `envconfig:"METRICS_PATH"`

	// Org settings
	EnableCloudForOrganization bool

	Projects        []*Project `yaml:"projects" ignored:"true"`
	Format          string     `yaml:"format,omitempty" ignored:"true"`
	ShowAllProjects bool       `yaml:"show_all_projects,omitempty" ignored:"true"`
	ShowSkipped     bool       `yaml:"show_skipped,omitempty" ignored:"true"`
	SyncUsageFile   bool       `yaml:"sync_usage_file,omitempty" ignored:"true"`
	Fields          []string   `yaml:"fields,omitempty" ignored:"true"`
	CompareTo       string
	GitDiffTarget   *string

	// Base configuration settings
	// RootPath defines the raw value of the `--path` flag provided by the user
	RootPath string
	// ConfigFilePath defines the raw value of the `--config-file` flag provided by the user
	ConfigFilePath string
	// UsageFilePath defines the raw value of the `--usage-file` flag provided by the user
	UsageFilePath string

	NoCache bool `yaml:"no_cache,omitempty" ignored:"true"`

	SkipErrLine bool

	// for testing
	EventsDisabled       bool
	logWriter            io.Writer
	logDisableTimestamps bool
}

func init() {
	err := loadDotEnv()
	if err != nil {
		logging.Logger.Fatal().Msg(err.Error())
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

// WorkingDirectory returns the filepath to either the directory specified by the --path
// flag or the directory that the binary has been run from.
func (c *Config) WorkingDirectory() string {
	if c.ConfigFilePath != "" {
		wd, err := os.Getwd()
		if err != nil {
			logging.Logger.Debug().Err(err).Msg("error getting working directory for repo path")
			return ""
		}

		return wd
	}

	return c.RootPath
}

// CachePath finds path which contains the .infracost directory. It traverses parent directories until a .infracost
// folder is found. If no .infracost folders exist then CachePath uses the current wd.
func (c *Config) CachePath() string {
	dir := c.WorkingDirectory()

	if s := c.cachePath(dir); s != "" {
		return s
	}

	// now let's try to traverse the parent directories outside the working directory.
	// We don't do this initially as this causing path problems when the cache directory
	// is created by concurrently running projects.
	abs, err := filepath.Abs(dir)
	if err == nil {
		if s := c.cachePath(abs); s != "" {
			return s
		}
	}

	return dir
}

func (c *Config) cachePath(dir string) string {
	for {
		cachePath := filepath.Join(dir, InfracostDir)
		if _, err := os.Stat(cachePath); err == nil {
			return filepath.Dir(cachePath)
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			break
		}
		dir = parentDir
	}

	return ""
}

func (c *Config) LoadFromConfigFile(path string, cmd *cobra.Command) error {
	cfgFile, err := LoadConfigFile(path)
	if err != nil {
		return err
	}

	c.Projects = cfgFile.Projects

	if len(cfgFile.TerraformSourceMapRegex) > 0 {
		c.TerraformSourceMapRegex = cfgFile.TerraformSourceMapRegex
		err = c.TerraformSourceMapRegex.Compile()
		if err != nil {
			return fmt.Errorf("error compiling terraform_source_map regex patterns: %w", err)
		}
	}

	// Reload the environment and global flags to overwrite any of the config file configs
	err = c.LoadFromEnv()
	if err != nil {
		return err
	}

	err = c.LoadGlobalFlags(cmd)
	if err != nil {
		return err
	}

	return nil
}

// WriteLevel is the log level that the Logger writes to LogWriter.
func (c *Config) WriteLevel() string {
	if c.DebugReport {
		return zerolog.LevelDebugValue
	}

	if c.LogLevel == "" {
		return zerolog.LevelInfoValue
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

// SetLogWriter sets the io.Writer that the logs should be piped to.
func (c *Config) SetLogWriter(w io.Writer) {
	c.logWriter = w
}

// LogWriter returns the writer the Logger should use to write logs to.
// In most cases this should be stderr, but it can also be a file.
func (c *Config) LogWriter() io.Writer {
	isCI := ciPlatform() != "" && !IsTest()
	return zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.PartsExclude = []string{"time"}
		w.FormatLevel = func(i interface{}) string {
			if i == nil {
				return ""
			}

			if isCI {
				return strings.ToUpper(fmt.Sprintf("%s", i))
			}

			if ll, ok := i.(string); ok {
				upper := strings.ToUpper(ll)

				switch ll {
				case zerolog.LevelTraceValue:
					return color.CyanString("%s", upper)
				case zerolog.LevelDebugValue:
					return color.MagentaString("%s", upper)
				case zerolog.LevelWarnValue:
					return color.YellowString("%s", upper)
				case zerolog.LevelErrorValue, zerolog.LevelFatalValue, zerolog.LevelPanicValue:
					return color.RedString("%s", upper)
				case zerolog.LevelInfoValue:
					return color.GreenString("%s", upper)
				default:
				}
			}

			return strings.ToUpper(fmt.Sprintf("%s", i))
		}

		if isCI {
			w.NoColor = true
			w.TimeFormat = time.RFC3339
			w.PartsExclude = nil
		}

		if c.logDisableTimestamps {
			w.PartsExclude = []string{"time"}
		}

		w.Out = os.Stderr
		if c.logWriter != nil {
			w.Out = c.logWriter
		}
	})
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

func (c *Config) LoadGlobalFlags(cmd *cobra.Command) error {
	if cmd.Flags().Changed("no-color") {
		c.NoColor, _ = cmd.Flags().GetBool("no-color")
	}
	color.NoColor = c.NoColor

	if cmd.Flags().Changed("log-level") {
		c.LogLevel, _ = cmd.Flags().GetString("log-level")
		err := logging.ConfigureBaseLogger(c)
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("debug-report") {
		c.DebugReport, _ = cmd.Flags().GetBool("debug-report")
		err := logging.ConfigureBaseLogger(c)
		if err != nil {
			return err
		}
	}

	return nil
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
	if os.Getenv("INFRACOST_DISABLE_ENVFILE") == "true" {
		return nil
	}

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

func CleanProjectName(name string) string {
	name = strings.TrimSuffix(name, "/")
	name = strings.ReplaceAll(name, "/", "-")

	if name == "." {
		return "main"
	}
	return name
}
