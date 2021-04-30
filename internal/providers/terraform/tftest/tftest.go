package tftest

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/usage"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/prices"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/infracost/infracost/internal/providers/terraform"
)

var update = flag.Bool("update", false, "update .golden files")

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

	tfdir, err := CreateTerraformProject(project)
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
	cfg := config.DefaultConfig()
	err := cfg.LoadFromEnv()
	assert.NoError(t, err)

	project, err := RunCostCalculations(cfg, tfProject, usage)
	assert.NoError(t, err)

	testutil.TestResources(t, project.Resources, checks)
}

func GoldenFileResourceTests(t *testing.T, testName string) {
	cfg := config.DefaultConfig()
	err := cfg.LoadFromEnv()
	require.NoError(t, err)

	// Load the terraform projects
	tfProjectData, err := ioutil.ReadFile(filepath.Join("testdata", testName, testName+".tf"))
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
		usageData, err = usage.LoadFromFile(usageFilePath, false)
		require.NoError(t, err)
	}

	// Generate the output
	project, err := RunCostCalculations(cfg, tfProject, usageData)
	require.NoError(t, err)

	r := output.ToOutputFormat([]*schema.Project{project})

	opts := output.Options{
		ShowSkipped: true,
		NoColor:     true,
		Fields:      cfg.Fields,
	}

	actual, err := output.ToTable(r, opts)
	require.NoError(t, err)

	// strip the first line of output since it contains the temporary project path
	endOfFirstLine := bytes.Index(actual, []byte("\n"))
	if endOfFirstLine > 0 {
		actual = actual[endOfFirstLine+1:]
	}

	// Load the snapshot result
	expected := []byte("")
	goldenFilePath := filepath.Join("testdata", testName, testName+".golden")
	if _, err := os.Stat(goldenFilePath); err == nil || !os.IsNotExist(err) {
		// golden file exists, load the data
		expected, err = ioutil.ReadFile(goldenFilePath)
		assert.NoError(t, err)
	}

	if !bytes.Equal(expected, actual) {
		if *update {
			err = ioutil.WriteFile(goldenFilePath, actual, 0600)
			assert.NoError(t, err)
			t.Logf(fmt.Sprintf("Wrote golden file %s", goldenFilePath))
		} else {
			// Generate the diff and error message.  We don't call assert.Equal because it escapes
			// newlines (\n) and the output looks terrible.
			diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
				A:        difflib.SplitLines(string(expected)),
				B:        difflib.SplitLines(string(actual)),
				FromFile: "Expected",
				FromDate: "",
				ToFile:   "Actual",
				ToDate:   "",
				Context:  1,
			})

			t.Errorf(fmt.Sprintf("\nOutput does not match golden file: \n\n%s\n", diff))
		}
	}
}

func RunCostCalculations(cfg *config.Config, tfProject TerraformProject, usage map[string]*schema.UsageData) (*schema.Project, error) {
	project, err := loadResources(cfg, tfProject, usage)
	if err != nil {
		return project, err
	}
	err = prices.PopulatePrices(cfg, project)
	if err != nil {
		return project, err
	}
	schema.CalculateCosts(project)
	return project, nil
}

func CreateTerraformProject(tfProject TerraformProject) (string, error) {
	return writeToTmpDir(tfProject)
}

func loadResources(cfg *config.Config, tfProject TerraformProject, usage map[string]*schema.UsageData) (*schema.Project, error) {
	tfdir, err := CreateTerraformProject(tfProject)
	if err != nil {
		return nil, err
	}

	provider := terraform.NewDirProvider(cfg, &config.Project{
		Path: tfdir,
	})

	return provider.LoadResources(usage)
}

func writeToTmpDir(tfProject TerraformProject) (string, error) {
	// Create temporary directory and output terraform code
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return tmpDir, err
	}

	for _, terraformFile := range tfProject.Files {
		fullPath := filepath.Join(tmpDir, terraformFile.Path)
		dir := filepath.Dir(fullPath)

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				return tmpDir, err
			}
		}

		err = ioutil.WriteFile(fullPath, []byte(terraformFile.Contents), 0600)
		if err != nil {
			return tmpDir, err
		}
	}

	return tmpDir, err
}
