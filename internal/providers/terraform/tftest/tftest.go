package tftest

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/usage"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/stretchr/testify/require"

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
	cfg := config.DefaultConfig()
	err := cfg.LoadFromEnv()
	assert.NoError(t, err)

	project, err := RunCostCalculations(t, cfg, tfProject, usage)
	assert.NoError(t, err)

	testutil.TestResources(t, project.Resources, checks)
}

type GoldenFileOptions = struct {
	Currency string
}

func GoldenFileResourceTests(t *testing.T, testName string) {
	GoldenFileResourceTestsWithOpts(t, testName,
		&GoldenFileOptions{
			Currency: "USD",
		})
}

func GoldenFileResourceTestsWithOpts(t *testing.T, testName string, options *GoldenFileOptions) {
	cfg := config.DefaultConfig()
	err := cfg.LoadFromEnv()

	if options != nil && options.Currency != "" {
		cfg.Currency = options.Currency
	}

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
	project, err := RunCostCalculations(t, cfg, tfProject, usageData)
	require.NoError(t, err)

	r := output.ToOutputFormat([]*schema.Project{project})
	r.Currency = cfg.Currency

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

func RunCostCalculations(t *testing.T, cfg *config.Config, tfProject TerraformProject, usage map[string]*schema.UsageData) (*schema.Project, error) {
	project, err := loadResources(t, cfg, tfProject, usage)
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

func CreateTerraformProject(tmpDir string, tfProject TerraformProject) (string, error) {
	return writeToTmpDir(tmpDir, tfProject)
}

func loadResources(t *testing.T, cfg *config.Config, tfProject TerraformProject, usage map[string]*schema.UsageData) (*schema.Project, error) {
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

	runCtx, err := config.NewRunContextFromEnv(context.Background())
	if err != nil {
		return nil, err
	}

	provider := terraform.NewDirProvider(config.NewProjectContext(
		runCtx,
		&config.Project{
			Path: tfdir,
		},
	))

	project := schema.NewProject("tftest", &schema.ProjectMetadata{})

	return project, provider.LoadResources(project, usage)
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
					srcData, err := ioutil.ReadFile(srcPath)
					if err != nil {
						return err
					}

					if err := ioutil.WriteFile(destPath, srcData, os.ModePerm); err != nil {
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

		err = ioutil.WriteFile(fullPath, []byte(terraformFile.Contents), 0600)
		if err != nil {
			return tmpDir, err
		}
	}

	return tmpDir, err
}
