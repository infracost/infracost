package tftest

import (
	"flag"
	"fmt"
	"infracost/internal/providers/terraform"
	"infracost/pkg/prices"
	"infracost/pkg/schema"
	"infracost/pkg/testutil"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/urfave/cli/v2"
)

var tfProviders = `
	terraform {
		required_providers {
			aws = {
				source  = "hashicorp/aws"
			}
			infracost = {
				source = "infracost.io/infracost/infracost"
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

func ResourceTests(t *testing.T, tf string, resourceChecks []testutil.ResourceCheck) {
	project := Project{
		Files: []File{
			{
				Path:     "main.tf",
				Contents: WithProviders(tf),
			},
		},
	}

	ResourceTestsForProject(t, project, resourceChecks)
}

func ResourceTestsForProject(t *testing.T, project Project, resourceChecks []testutil.ResourceCheck) {
	resources, err := RunCostCalculations(project)
	if err != nil {
		t.Error(err)
	}

	for _, resourceCheck := range resourceChecks {
		testutil.TestResource(t, resources, resourceCheck)
	}
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

func loadResources(project Project) ([]*schema.Resource, error) {
	tfdir, err := writeToTmpDir(project)
	if err != nil {
		return nil, err
	}

	flags := flag.NewFlagSet("test", 0)
	flags.String("tfdir", tfdir, "")
	c := cli.NewContext(nil, flags, nil)

	provider := terraform.Provider()
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

		err = ioutil.WriteFile(fullPath, []byte(terraformFile.Contents), 0644)
		if err != nil {
			return tmpDir, err
		}
	}

	return tmpDir, err
}
