package terraform_test

import (
	"infracost/internal/providers/terraform/tftest"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLoadResources_rootModule(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	terraformFiles := []*tftest.TerraformFile{
		{
			Path: "main.tf",
			Contents: tftest.WithProviders(`
				resource "aws_nat_gateway" "nat1" {
					allocation_id = "eip-12345678"
					subnet_id     = "subnet-12345678"
				}
			`),
		},
	}

	resources, err := tftest.LoadResourcesForProject(terraformFiles)
	if err != nil {
		t.Error(err)
	}

	if len(resources) != 1 {
		t.Errorf("Unexpected number of resources returned: %d)", len(resources))
	}

	if !cmp.Equal(resources[0].Name, "aws_nat_gateway.nat1") {
		t.Errorf("Resource name is incorrect: %s", resources[0].Name)
	}
}

func TestLoadResources_nestedModule(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	terraformFiles := []*tftest.TerraformFile{
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
	}

	resources, err := tftest.LoadResourcesForProject(terraformFiles)
	if err != nil {
		t.Error(err)
	}

	if len(resources) != 2 {
		t.Errorf("Unexpected number of resources returned: %d)", len(resources))
	}

	resourceNames := make([]string, 0, len(resources))
	for _, resource := range resources {
		resourceNames = append(resourceNames, resource.Name)
	}
	sort.Strings(resourceNames)

	expectedResourceNames := []string{
		"module.module1.aws_nat_gateway.nat1",
		"module.module1.module.module2.aws_nat_gateway.nat2",
	}
	if !cmp.Equal(resourceNames, expectedResourceNames) {
		t.Errorf("Unexpected resource names: %v", resourceNames)
	}
}
