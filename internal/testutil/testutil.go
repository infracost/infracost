package testutil

import (
	"fmt"
	"infracost/pkg/config"
	"infracost/pkg/costs"
	"infracost/pkg/terraform"
	"io/ioutil"
	"path/filepath"

	"github.com/google/go-cmp/cmp/cmpopts"
)

var tfPrefix = `
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

func RunTFCostBreakdown(resourceTf string) ([]costs.ResourceCostBreakdown, error) {
	tf := fmt.Sprintf("%s\n%s", tfPrefix, resourceTf)

	// Create temporary directory and output terraform code
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return []costs.ResourceCostBreakdown{}, err
	}

	err = ioutil.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tf), 0644)
	if err != nil {
		return []costs.ResourceCostBreakdown{}, err
	}

	planJSON, err := terraform.GeneratePlanJSON(tmpDir, "")
	if err != nil {
		return []costs.ResourceCostBreakdown{}, err
	}

	resources, err := terraform.ParsePlanJSON(planJSON)
	if err != nil {
		return []costs.ResourceCostBreakdown{}, err
	}

	q := costs.NewGraphQLQueryRunner(fmt.Sprintf("%s/graphql", config.Config.ApiUrl))
	return costs.GenerateCostBreakdowns(q, resources)
}

var PriceHashResultSort = cmpopts.SortSlices(func(x, y []string) bool {
	return fmt.Sprintf("%s %s", x[0], x[1]) < fmt.Sprintf("%s %s", y[0], y[1])
})

func ExtractPriceHashes(resourceCostBreakdowns []costs.ResourceCostBreakdown) [][]string {
	priceHashes := [][]string{}

	for _, resourceCostBreakdown := range resourceCostBreakdowns {
		for _, priceComponentCost := range resourceCostBreakdown.PriceComponentCosts {
			priceHashes = append(priceHashes, []string{resourceCostBreakdown.Resource.Address(), priceComponentCost.PriceComponent.Name(), priceComponentCost.PriceHash})
		}

		priceHashes = append(priceHashes, ExtractPriceHashes(resourceCostBreakdown.SubResourceCosts)...)
	}

	return priceHashes
}

func PriceComponentCostFor(resourceCostBreakdowns []costs.ResourceCostBreakdown, resourceAddress string, priceComponentName string) *costs.PriceComponentCost {
	for _, resourceCostBreakdown := range resourceCostBreakdowns {
		for _, priceComponentCost := range resourceCostBreakdown.PriceComponentCosts {
			if resourceCostBreakdown.Resource.Address() == resourceAddress && priceComponentCost.PriceComponent.Name() == priceComponentName {
				return &priceComponentCost
			}
		}

		priceComponentCost := PriceComponentCostFor(resourceCostBreakdown.SubResourceCosts, resourceAddress, priceComponentName)
		if priceComponentCost != nil {
			return priceComponentCost
		}
	}

	return nil
}
