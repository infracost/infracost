package tftest

import (
	"flag"
	"fmt"
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

func EnsurePluginsInstalled(tmpDir string) {
	flag.Parse()
	if !testing.Short() {
		once.Do(func() {
			// Ensure plugins are installed and cached
			err := installPlugins(tmpDir)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		})
	}
}

func installPlugins(tmpDir string) error {
	project := TerraformProject{
		Files: []File{
			{
				Path:     "init.tf",
				Contents: WithProviders(""),
			},
		},
	}

	tfdir, err := CreateTerraformProject(project, tmpDir)
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

func ResourceTests(t *testing.T, tf string, usage map[string]*schema.UsageData, checks []testutil.ResourceCheck, tmpDir string) {
	project := TerraformProject{
		Files: []File{
			{
				Path:     "main.tf",
				Contents: WithProviders(tf),
			},
		},
	}

	ResourceTestsForTerraformProject(t, project, usage, checks, tmpDir)
}

func ResourceTestsForTerraformProject(t *testing.T, tfProject TerraformProject, usage map[string]*schema.UsageData, checks []testutil.ResourceCheck, tmpDir string) {
	cfg := config.DefaultConfig()
	err := cfg.LoadFromEnv()
	assert.NoError(t, err)

	project, err := RunCostCalculations(cfg, tfProject, usage, tmpDir)
	assert.NoError(t, err)

	testutil.TestResources(t, project.Resources, checks)
}

func RunCostCalculations(cfg *config.Config, tfProject TerraformProject, usage map[string]*schema.UsageData, tmpDir string) (*schema.Project, error) {
	project, err := loadResources(cfg, tfProject, usage, tmpDir)
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

func CreateTerraformProject(tfProject TerraformProject, tmpDir string) (string, error) {
	return writeToTmpDir(tfProject, tmpDir)
}

func loadResources(cfg *config.Config, tfProject TerraformProject, usage map[string]*schema.UsageData, tmpDir string) (*schema.Project, error) {
	tfdir, err := CreateTerraformProject(tfProject, tmpDir)
	if err != nil {
		return nil, err
	}

	provider := terraform.NewDirProvider(cfg, &config.Project{
		Path: tfdir,
	})

	return provider.LoadResources(usage)
}

func writeToTmpDir(tfProject TerraformProject, tmpDir string) (string, error) {
	os.Remove(filepath.Join(tmpDir, "init.tf"))
	os.Remove(filepath.Join(tmpDir, "main.tf"))

	for _, terraformFile := range tfProject.Files {
		fullPath := filepath.Join(tmpDir, terraformFile.Path)
		dir := filepath.Dir(fullPath)

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				return tmpDir, err
			}
		}

		err := ioutil.WriteFile(fullPath, []byte(terraformFile.Contents), 0600)
		if err != nil {
			return tmpDir, err
		}
	}

	return tmpDir, nil
}
