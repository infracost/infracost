package tftest

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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
			infracost = {
				source = "infracost/infracost"
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

	provider "infracost" {}
`

var pluginCache = filepath.Join(config.RootDir(), ".test_cache/terraform_plugins")

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

func InstallPlugins() error {
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

	fmt.Println(tfdir)
	fmt.Println(pluginCache)

	_, err = terraform.Cmd(opts, "init", "-no-color")
	if err != nil {
		return errors.Wrap(err, "Error initializing Terraform working directory")
	}

	return nil
}

func ResourceTests(t *testing.T, tf string, checks []testutil.ResourceCheck) {
	project := Project{
		Files: []File{
			{
				Path:     "main.tf",
				Contents: WithProviders(tf),
			},
		},
	}

	ResourceTestsForProject(t, project, checks)
}

func ResourceTestsForProject(t *testing.T, project Project, checks []testutil.ResourceCheck) {
	resources, err := RunCostCalculations(project)
	assert.NoError(t, err)

	testutil.TestResources(t, resources, checks)
}

func RunCostCalculations(project Project) ([]*schema.Resource, error) {
	resources, err := loadResources(project)
	if err != nil {
		return resources, err
	}
	err = prices.PopulatePrices(resources)
	if err != nil {
		return resources, err
	}
	schema.CalculateCosts(resources)
	return resources, nil
}

func CreateProject(project Project) (string, error) {
	return writeToTmpDir(project)
}

func loadResources(project Project) ([]*schema.Resource, error) {
	tfdir, err := CreateProject(project)
	if err != nil {
		return nil, err
	}

	flags := flag.NewFlagSet("test", 0)
	flags.String("tfdir", tfdir, "")
	c := cli.NewContext(nil, flags, nil)

	provider := terraform.New()
	err = provider.ProcessArgs(c)
	if err != nil {
		return nil, err
	}

	return provider.LoadResources()
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
