package terraform_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

var tmpDir string

func TestMain(m *testing.M) {
	var err error
	tmpDir, err = ioutil.TempDir("", "tmp_terraform_test_*")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	tftest.EnsurePluginsInstalled(tmpDir)
	code := m.Run()

	os.RemoveAll(tmpDir) // clean up
	os.Exit(code)
}

func TestLoadResources_rootModule(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	project := tftest.TerraformProject{
		Files: []tftest.File{
			{
				Path: "main.tf",
				Contents: tftest.WithProviders(`
					resource "aws_nat_gateway" "nat1" {
						allocation_id = "eip-12345678"
						subnet_id     = "subnet-12345678"
					}
				`),
			},
		},
	}

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "aws_nat_gateway.nat1",
			SkipCheck: true,
		},
	}

	tftest.ResourceTestsForTerraformProject(t, project, schema.NewEmptyUsageMap(), resourceChecks, tmpDir)
}

func TestLoadResources_nestedModule(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	project := tftest.TerraformProject{
		Files: []tftest.File{
			{
				Path: "main.tf",
				Contents: tftest.WithProviders(`
					module "module1" {
						source   = "./module1"
					}
				`),
			},
			{
				Path: "module1/main.tf",
				Contents: tftest.WithProviders(`
					module "module2" {
						source   = "./module2"
					}

					resource "aws_nat_gateway" "nat1" {
						allocation_id = "eip-12345678"
						subnet_id     = "subnet-12345678"
					}
				`),
			},
			{
				Path: "module1/module2/main.tf",
				Contents: tftest.WithProviders(`
					resource "aws_nat_gateway" "nat2" {
						allocation_id = "eip-12345678"
						subnet_id     = "subnet-12345678"
					}
				`),
			},
		},
	}

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "module.module1.aws_nat_gateway.nat1",
			SkipCheck: true,
		},
		{
			Name:      "module.module1.module.module2.aws_nat_gateway.nat2",
			SkipCheck: true,
		},
	}

	tftest.ResourceTestsForTerraformProject(t, project, schema.NewEmptyUsageMap(), resourceChecks, tmpDir)
}
