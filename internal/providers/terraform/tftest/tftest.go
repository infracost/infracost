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
		s3_force_path_style         = true
		skip_credentials_validation = true
		skip_metadata_api_check     = true
		skip_requesting_account_id  = true
		skip_get_ec2_platforms      = true
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

func ResourceTests(t *testing.T, tf string, usage map[string]*schema.UsageData, checks []testutil.ResourceCheck) {
	project := TerraformProject{
		Files: []File{
			{
				Path:     "main.tf",
				Contents: WithProviders(tf),
			},
		},
	}

	ResourceTestsForTerraformProject(t, project, usage, checks)
}

func ResourceTestsForTerraformProject(t *testing.T, tfProject TerraformProject, usage map[string]*schema.UsageData, checks []testutil.ResourceCheck) {
	runCtx, err := config.NewRunContextFromEnv(context.Background())
	assert.NoError(t, err)

	projects, err := RunCostCalculations(t, runCtx, tfProject, usage)
	assert.NoError(t, err)
	assert.Len(t, projects, 1)

	testutil.TestResources(t, projects[0].Resources, checks)
}

type GoldenFileOptions = struct {
	Currency    string
	CaptureLogs bool
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
	var usageData map[string]*schema.UsageData
	usageFilePath := filepath.Join("testdata", testName, testName+".usage.yml")
	if _, err := os.Stat(usageFilePath); err == nil || !os.IsNotExist(err) {
		// usage file exists, load the data
		usageFile, err := usage.LoadUsageFile(usageFilePath)
		require.NoError(t, err)
		usageData = usageFile.ToUsageDataMap()
	}

	// Generate the output
	projects, err := RunCostCalculations(t, runCtx, tfProject, usageData)
	require.NoError(t, err)

	r, err := output.ToOutputFormat(projects)
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

func RunCostCalculations(t *testing.T, runCtx *config.RunContext, tfProject TerraformProject, usage map[string]*schema.UsageData) ([]*schema.Project, error) {
	projects, err := loadResources(t, runCtx, tfProject, usage)
	if err != nil {
		return projects, err
	}

	for _, project := range projects {
		err = prices.PopulatePrices(runCtx, project)
		if err != nil {
			return projects, err
		}
		schema.CalculateCosts(project)
	}

	return projects, nil
}

func GoldenFileUsageSyncTest(t *testing.T, testName string) {
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

	usageFilePath := filepath.Join("testdata", testName, testName+"_existing_usage.yml")
	actual, err := RunSyncUsage(t, runCtx, tfProject, usageFilePath)
	require.NoError(t, err)

	goldenFilePath := filepath.Join("testdata", testName, testName+".golden")
	testutil.AssertGoldenFile(t, goldenFilePath, actual)
}

func RunSyncUsage(t *testing.T, runCtx *config.RunContext, tfProject TerraformProject, usageFilePath string) ([]byte, error) {
	tmpDir := t.TempDir()
	projects, err := loadResources(t, runCtx, tfProject, map[string]*schema.UsageData{})
	if err != nil {
		return nil, err
	}

	usageFile, err := usage.LoadUsageFile(usageFilePath)
	if err != nil {
		return nil, err
	}

	_, err = usage.SyncUsageData(usageFile, projects)
	if err != nil {
		return nil, err
	}

	out := filepath.Join(tmpDir, "actual_usage.yml")
	err = usageFile.WriteToPath(out)
	if err != nil {
		return nil, err
	}

	return os.ReadFile(out)
}

func CreateTerraformProject(tmpDir string, tfProject TerraformProject) (string, error) {
	return writeToTmpDir(tmpDir, tfProject)
}

func loadResources(t *testing.T, runCtx *config.RunContext, tfProject TerraformProject, usage map[string]*schema.UsageData) ([]*schema.Project, error) {
	tmpDir := t.TempDir()

	_, err := os.ReadDir(initCache)
	if err == nil {
		if err := copyInitCacheToPath(initCache, tmpDir); err != nil {
			return nil, err
		}
	} else {
		t.Log(fmt.Sprintf("Couldn't copy terraform init cache from %s", initCache))
	}

	tfdir, err := CreateTerraformProject(tmpDir, tfProject)
	if err != nil {
		return nil, err
	}

	//provider := terraform.NewDirProvider(config.NewProjectContext(
	//	runCtx,
	//	&config.Project{
	//		Path: tfdir,
	//	},
	//))
	projectCtx := config.NewProjectContext(
		runCtx,
		&config.Project{
			Path: tfdir,
		},
	)

	provider, err := terraform.NewHCLProvider(projectCtx, terraform.NewPlanJSONProvider(projectCtx))
	if err != nil {
		return nil, err
	}

	return provider.LoadResources(usage)
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
