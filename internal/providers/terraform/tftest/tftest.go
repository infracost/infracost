package tftest

import (
	"flag"
	"fmt"
	"infracost/internal/providers/terraform"
	"infracost/pkg/prices"
	"infracost/pkg/schema"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

var tfProviders = `
	provider "aws" {
		region                      = "us-east-1"
		s3_force_path_style         = true
		skip_credentials_validation = true
		skip_metadata_api_check     = true
		skip_requesting_account_id  = true
		access_key                  = "mock_access_key"
		secret_key                  = "mock_secret_key"
	}
`

type TerraformFile struct {
	Path string
	Contents string
}

func WithProviders(tf string) string {
	return fmt.Sprintf("%s%s", tfProviders, tf)
}

func RunCostCalculation(tf string) ([]*schema.Resource, error) {
	resources, err := LoadResources(tf)
	if err != nil {
		return resources, err
	}
	prices.PopulatePrices(resources)
	schema.CalculateCosts(resources)
	return resources, nil
}

func LoadResources(tf string) ([]*schema.Resource, error) {
	terraformFiles := []*TerraformFile{
		{
			Path: "main.tf",
			Contents: WithProviders(tf),
		},
	}

	return LoadResourcesForProject(terraformFiles)
}

func LoadResourcesForProject(terraformFiles []*TerraformFile) ([]*schema.Resource, error) {
	tfdir, err := writeToTmpDir(terraformFiles)
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


func writeToTmpDir(terraformFiles []*TerraformFile) (string, error) {
	// Create temporary directory and output terraform code
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return tmpDir, err
	}

	for _, terraformFile := range terraformFiles {
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
