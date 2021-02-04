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

	"github.com/urfave/cli/v2"
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
`

var (
	pluginCache = filepath.Join(config.RootDir(), ".test_cache/terraform_plugins")
	once        sync.Once
)

type Project struct {
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
	project := Project{
		Files: []File{
			{
				Path:     "init.tf",
				Contents: WithProviders(""),
			},
		},
	}

	tfdir, err := CreateProject(project)
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
		TerraformDir: tfdir,
	}

	_, err = terraform.Cmd(opts, "init", "-no-color")
	if err != nil {
		return errors.Wrap(err, "Error initializing Terraform working directory")
	}

	return nil
}

func ResourceTests(t *testing.T, tf string, usage map[string]*schema.UsageData, checks []testutil.ResourceCheck) {
	project := Project{
		Files: []File{
			{
				Path:     "main.tf",
				Contents: WithProviders(tf),
			},
		},
	}

	ResourceTestsForProject(t, project, usage, checks)
}

func ResourceTestsForProject(t *testing.T, project Project, usage map[string]*schema.UsageData, checks []testutil.ResourceCheck) {
	state, err := RunCostCalculations(project, usage)
	assert.NoError(t, err)

	testutil.TestResources(t, state.PlannedState.Resources, checks)
}

func RunCostCalculations(project Project, usage map[string]*schema.UsageData) (*schema.State, error) {
	state, err := loadResources(project, usage)
	if err != nil {
		return state, err
	}
	err = prices.PopulatePrices(state)
	if err != nil {
		return state, err
	}
	schema.CalculateCosts(state)
	return state, nil
}

func CreateProject(project Project) (string, error) {
	return writeToTmpDir(project)
}

func loadResources(project Project, usage map[string]*schema.UsageData) (*schema.State, error) {
	tfdir, err := CreateProject(project)
	if err != nil {
		return nil, err
	}

	flags := flag.NewFlagSet("test", 0)
	flags.String("tfdir", tfdir, "")
	flags.String("usage-file", filepath.Join(tfdir, "infracost-usage.yml"), "")
	c := cli.NewContext(nil, flags, nil)

	provider := terraform.New()
	err = provider.ProcessArgs(c)
	if err != nil {
		return nil, err
	}

	return provider.LoadResources(usage)
}

func writeToTmpDir(project Project) (string, error) {
	// Create temporary directory and output terraform code
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return tmpDir, err
	}

	for _, terraformFile := range project.Files {
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
