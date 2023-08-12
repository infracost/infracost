package tftest

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/usage"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/prices"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform"
)

var tfProviders = `
	terraform {
		required_providers {
			aws = {
				source  = "hashicorp/aws"
			}
			google = {
				source  = "hashicorp/google"
			}
			azurerm = {
				source  = "hashicorp/azurerm"
			}
		}
	}

	provider "aws" {
		region                      = "us-east-1"
		skip_credentials_validation = true
		skip_metadata_api_check     = true
		skip_requesting_account_id  = true
		skip_region_validation      = true
		access_key                  = "mock_access_key"
		secret_key                  = "mock_secret_key"
	}

	provider "google" {
		credentials = "{\"type\":\"service_account\"}"
		region = "us-central1"
	}

	provider "google-beta" {
		credentials = "{\"type\":\"service_account\"}"
		region = "us-central1"
	}

	provider "azurerm" {
		skip_provider_registration = true
		features {}
	}
`

var (
	pluginCache = filepath.Join(config.RootDir(), ".test_cache/terraform_plugins")
	initCache   = filepath.Join(config.RootDir(), ".test_cache/terraform_init")
	once        sync.Once
)

type TerraformProject struct {
	Files []File
}

type File struct {
	Path     string
	Contents string
}

type ResourceTestOptions struct {
	UsageData    schema.UsageMap
	RequiresInit bool
}

type LoadResourcesOptions struct {
	ProviderType     string
	TerraformProject TerraformProject
	RunCtx           *config.RunContext
	UsageData        schema.UsageMap
	RequiresInit     bool
}

func WithProviders(tf string) string {
	return fmt.Sprintf("%s%s", tfProviders, tf)
}

func EnsurePluginsInstalled() {
	flag.Parse()
	if !testing.Short() {
		once.Do(func() {
			// Ensure plugins are installed and cached
			err := installPlugins()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		})
	}
}

func installPlugins() error {
	project := TerraformProject{
		Files: []File{
			{
				Path:     "init.tf",
				Contents: WithProviders(""),
			},
		},
	}

	err := os.MkdirAll(initCache, os.ModePerm)
	if err != nil {
		log.Errorf("Error creating init cache directory: %s", err.Error())
	}

	tfdir, err := writeToTmpDir(initCache, project)
	if err != nil {
		return errors.Wrap(err, "Error creating Terraform project")
	}

	err = os.MkdirAll(pluginCache, os.ModePerm)
	if err != nil {
		log.Errorf("Error creating plugin cache directory: %s", err.Error())
	} else {
		os.Setenv("TF_PLUGIN_CACHE_DIR", pluginCache)
	}

	opts := &terraform.CmdOptions{
		Dir: tfdir,
	}

	_, err = terraform.Cmd(opts, "init", "-no-color")
	if err != nil {
		return errors.Wrap(err, "Error initializing Terraform working directory")
	}

	return nil
}

func generatePlanJSON(t *testing.T, tfdir string, requiresInit bool) string {
	t.Helper()

	opts := &terraform.CmdOptions{
		Dir: tfdir,
	}

	tmpDir, err := os.MkdirTemp("", "infracost-tftest-")
	require.NoError(t, err, "error creating temp dir")

	if requiresInit {
		_, err = terraform.Cmd(opts, "init", "-input=false", "-no-color", "-upgrade")
		msgs := []string{"error running terraform init"}
		if err != nil && err.(*terraform.CmdError) != nil {
			msgs = append(msgs, string(err.(*terraform.CmdError).Stderr))
		}
		require.NoError(t, err, msgs)
	}

	planFile := filepath.Join(tmpDir, "tfplan.binary")
	defer os.Remove(planFile)

	_, err = terraform.Cmd(opts, "plan", "-input=false", "-lock=false", "-no-color", fmt.Sprintf("-out=%s", planFile))
	msgs := []string{"error running terraform plan"}
	if err != nil && err.(*terraform.CmdError) != nil {
		msgs = append(msgs, string(err.(*terraform.CmdError).Stderr))
	}
	require.NoError(t, err, msgs)

	out, err := terraform.Cmd(opts, "show", "-no-color", "-json", planFile)
	msgs = []string{"error running terraform show"}
	if err != nil && err.(*terraform.CmdError) != nil {
		msgs = append(msgs, string(err.(*terraform.CmdError).Stderr))
	}
	require.NoError(t, err, msgs)

	// Write out to tmp file
	planJSONFile := filepath.Join(tmpDir, "plan.json")
	err = os.WriteFile(planJSONFile, out, os.ModePerm)
	require.NoError(t, err, "error writing plan JSON")

	return planJSONFile
}

func ResourceTests(t *testing.T, tf string, checks []testutil.ResourceCheck, opts ResourceTestOptions) {
	project := TerraformProject{
		Files: []File{
			{
				Path:     "main.tf",
				Contents: WithProviders(tf),
			},
		},
	}

	ResourceTestsForTerraformProject(t, project, checks, opts)
}

func ResourceTestsForTerraformProject(t *testing.T, project TerraformProject, checks []testutil.ResourceCheck, opts ResourceTestOptions) {
	runCtx, err := config.NewRunContextFromEnv(context.Background())
	assert.NoError(t, err)

	loadResourcesOpts := LoadResourcesOptions{
		TerraformProject: project,
		RunCtx:           runCtx,
		UsageData:        opts.UsageData,
		RequiresInit:     opts.RequiresInit,
	}

	t.Run("HCL", func(t *testing.T) {
		loadResourcesOpts.ProviderType = "hcl"
		resourceTestsForTfProject(t, checks, loadResourcesOpts)
	})

	t.Run("Terraform Plan JSON", func(t *testing.T) {
		loadResourcesOpts.ProviderType = "terraform"
		resourceTestsForTfProject(t, checks, loadResourcesOpts)
	})
}

func resourceTestsForTfProject(t *testing.T, checks []testutil.ResourceCheck, loadResourceOpts LoadResourcesOptions) {
	t.Helper()

	runCtx := loadResourceOpts.RunCtx

	projects := loadResources(t, loadResourceOpts)

	projects, err := RunCostCalculations(runCtx, projects)
	assert.NoError(t, err)
	assert.Len(t, projects, 1)

	testutil.TestResources(t, projects[0].Resources, checks)
}

type GoldenFileOptions = struct {
	Currency       string
	CaptureLogs    bool
	RequiresInit   bool
	IgnorePlanJSON bool
}

func DefaultGoldenFileOptions() *GoldenFileOptions {
	return &GoldenFileOptions{
		Currency:    "USD",
		CaptureLogs: false,
	}
}

func GoldenFileResourceTests(t *testing.T, testName string) {
	GoldenFileResourceTestsWithOpts(t, testName, DefaultGoldenFileOptions())
}

func GoldenFileResourceTestsWithOpts(t *testing.T, testName string, options *GoldenFileOptions) {
	t.Run("HCL", func(t *testing.T) {
		goldenFileResourceTestWithOpts(t, "hcl", testName, options)
	})

	if options != nil && options.IgnorePlanJSON {
		return
	}

	t.Run("Terraform Plan JSON", func(t *testing.T) {
		goldenFileResourceTestWithOpts(t, "terraform", testName, options)
	})
}

func GoldenFileHCLResourceTestsWithOpts(t *testing.T, testName string, options *GoldenFileOptions) {
	goldenFileResourceTestWithOpts(t, "hcl", testName, options)
}

func goldenFileResourceTestWithOpts(t *testing.T, pName string, testName string, options *GoldenFileOptions) {
	t.Helper()

	runCtx, err := config.NewRunContextFromEnv(context.Background())

	var logBuf *bytes.Buffer
	if options != nil && options.CaptureLogs {
		logBuf = testutil.ConfigureTestToCaptureLogs(t, runCtx)
	} else {
		testutil.ConfigureTestToFailOnLogs(t, runCtx)
	}

	if options != nil && options.Currency != "" {
		runCtx.Config.Currency = options.Currency
	}

	require.NoError(t, err)

	// Load the terraform projects
	tfProjectData, err := os.ReadFile(filepath.Join("testdata", testName, testName+".tf"))
	require.NoError(t, err)
	tfProject := TerraformProject{
		Files: []File{
			{
				Path:     "main.tf",
				Contents: string(tfProjectData),
			},
		},
	}

	// Load the usage data, if any.
	var usageData schema.UsageMap
	usageFilePath := filepath.Join("testdata", testName, testName+".usage.yml")
	if _, err := os.Stat(usageFilePath); err == nil || !os.IsNotExist(err) {
		// usage file exists, load the data
		usageFile, err := usage.LoadUsageFile(usageFilePath)
		require.NoError(t, err)
		usageData = usageFile.ToUsageDataMap()
	}

	projects := loadResources(t, LoadResourcesOptions{
		ProviderType:     pName,
		TerraformProject: tfProject,
		RunCtx:           runCtx,
		UsageData:        usageData,
		RequiresInit:     options.RequiresInit,
	})

	// Generate the output
	projects, err = RunCostCalculations(runCtx, projects)
	require.NoError(t, err)

	r, err := output.ToOutputFormat(runCtx.Config, projects)
	if err != nil {
		require.NoError(t, err)
	}
	r.Currency = runCtx.Config.Currency

	opts := output.Options{
		ShowSkipped: true,
		NoColor:     true,
		Fields:      runCtx.Config.Fields,
	}

	actual, err := output.ToTable(r, opts)
	require.NoError(t, err)

	// strip the first line of output since it contains the temporary project path
	endOfFirstLine := bytes.Index(actual, []byte("\n"))
	if endOfFirstLine > 0 {
		actual = actual[endOfFirstLine+1:]
	}

	if logBuf != nil && logBuf.Len() > 0 {
		actual = append(actual, "\nLogs:\n"...)

		// need to sort the logs so they can be compared consistently
		logLines := strings.Split(logBuf.String(), "\n")
		sort.Strings(logLines)
		actual = append(actual, strings.Join(logLines, "\n")...)
	}

	goldenFilePath := filepath.Join("testdata", testName, testName+".golden")
	testutil.AssertGoldenFile(t, goldenFilePath, actual)
}

func loadResources(t *testing.T, opts LoadResourcesOptions) []*schema.Project {
	t.Helper()

	tfdir := createTerraformProject(t, opts.TerraformProject)
	path := tfdir
	runCtx := opts.RunCtx

	runCtx.Config.RootPath = tfdir
	var provider schema.Provider
	if opts.ProviderType == "hcl" {
		provider = newHCLProvider(t, runCtx, tfdir)
	} else {
		planJSON := generatePlanJSON(t, tfdir, opts.RequiresInit)
		defer os.Remove(planJSON)
		path = planJSON

		provider = terraform.NewPlanJSONProvider(config.NewProjectContext(runCtx, &config.Project{
			Path: planJSON,
		}, log.Fields{}), false)
	}

	projects, err := provider.LoadResources(opts.UsageData)
	require.NoError(t, err)

	for _, project := range projects {
		project.Name = strings.ReplaceAll(project.Name, path, t.Name())
		project.Name = strings.ReplaceAll(project.Name, "/Terraform_Plan_JSON", "")
		project.Name = strings.ReplaceAll(project.Name, "/HCL", "")
		project.BuildResources(schema.UsageMap{})
	}

	return projects
}

func RunCostCalculations(runCtx *config.RunContext, projects []*schema.Project) ([]*schema.Project, error) {
	for _, project := range projects {
		err := prices.PopulatePrices(runCtx, project)
		if err != nil {
			return projects, err
		}

		schema.CalculateCosts(project)
	}

	return projects, nil
}

func GoldenFileUsageSyncTest(t *testing.T, testName string) {
	t.Run("HCL", func(t *testing.T) {
		goldenFileSyncTest(t, "hcl", testName)
	})

	t.Run("Terraform Plan JSON", func(t *testing.T) {
		goldenFileSyncTest(t, "terraform", testName)
	})
}

func goldenFileSyncTest(t *testing.T, pName, testName string) {
	runCtx, err := config.NewRunContextFromEnv(context.Background())
	require.NoError(t, err)

	tfProjectData, err := os.ReadFile(filepath.Join("testdata", testName, testName+".tf"))
	require.NoError(t, err)
	tfProject := TerraformProject{
		Files: []File{
			{
				Path:     "main.tf",
				Contents: string(tfProjectData),
			},
		},
	}

	projectCtx := config.NewProjectContext(runCtx, &config.Project{}, log.Fields{})

	usageFilePath := filepath.Join("testdata", testName, testName+"_existing_usage.yml")
	projects := loadResources(t, LoadResourcesOptions{
		ProviderType:     pName,
		TerraformProject: tfProject,
		RunCtx:           runCtx,
		UsageData:        schema.UsageMap{},
	})

	actual := RunSyncUsage(t, projectCtx, projects, usageFilePath)
	require.NoError(t, err)

	goldenFilePath := filepath.Join("testdata", testName, testName+".golden")
	testutil.AssertGoldenFile(t, goldenFilePath, actual)
}

func RunSyncUsage(t *testing.T, projectCtx *config.ProjectContext, projects []*schema.Project, usageFilePath string) []byte {
	t.Helper()
	usageFile, err := usage.LoadUsageFile(usageFilePath)
	require.NoError(t, err)

	_, err = usage.SyncUsageData(projectCtx, usageFile, projects)
	require.NoError(t, err)

	out := filepath.Join(t.TempDir(), "actual_usage.yml")
	err = usageFile.WriteToPath(out)
	require.NoError(t, err)

	b, err := os.ReadFile(out)
	require.NoError(t, err)

	return b
}

func CreateTerraformProject(tmpDir string, tfProject TerraformProject) (string, error) {
	return writeToTmpDir(tmpDir, tfProject)
}

func newHCLProvider(t *testing.T, runCtx *config.RunContext, tfdir string) *terraform.HCLProvider {
	t.Helper()

	projectCtx := config.NewProjectContext(runCtx, &config.Project{
		Path: tfdir,
	}, log.Fields{})

	provider, err := terraform.NewHCLProvider(projectCtx, &terraform.HCLProviderConfig{SuppressLogging: true})
	require.NoError(t, err)

	return provider
}

func createTerraformProject(t *testing.T, tfProject TerraformProject) string {
	t.Helper()
	tmpDir := t.TempDir()

	_, err := os.ReadDir(initCache)
	if err == nil {
		err := copyInitCacheToPath(initCache, tmpDir)
		require.NoError(t, err)
	} else {
		t.Logf("Couldn't copy terraform init cache from %s", initCache)
	}

	tfdir, err := CreateTerraformProject(tmpDir, tfProject)
	require.NoError(t, err)

	return tfdir
}

func copyInitCacheToPath(source, destination string) error {
	files, err := os.ReadDir(source)
	if err != nil {
		return err
	}

	for _, file := range files {
		srcPath := filepath.Join(source, file.Name())
		destPath := filepath.Join(destination, file.Name())

		if file.IsDir() {
			if err := os.MkdirAll(destPath, os.ModePerm); err != nil {
				return err
			}
			if err := copyInitCacheToPath(srcPath, destPath); err != nil {
				return err
			}
		} else {
			info, err := file.Info()
			if err != nil {
				return err
			}
			if info.Mode()&os.ModeSymlink == os.ModeSymlink {
				if err := os.Symlink(srcPath, destPath); err != nil {
					return err
				}
			} else {
				if file.Name() != "init.tf" { // don't copy init.tf since the provider block will conflict with main.tf
					srcData, err := os.ReadFile(srcPath)
					if err != nil {
						return err
					}

					if err := os.WriteFile(destPath, srcData, os.ModePerm); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func writeToTmpDir(tmpDir string, tfProject TerraformProject) (string, error) {
	var err error

	for _, terraformFile := range tfProject.Files {
		fullPath := filepath.Join(tmpDir, terraformFile.Path)
		dir := filepath.Dir(fullPath)

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				return tmpDir, err
			}
		}

		err = os.WriteFile(fullPath, []byte(terraformFile.Contents), 0600)
		if err != nil {
			return tmpDir, err
		}
	}

	return tmpDir, err
}
