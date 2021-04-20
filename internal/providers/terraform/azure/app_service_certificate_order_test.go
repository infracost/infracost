package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestAzureRMAppServiceCertificateOrder(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "azurerm_app_service_certificate_order" "standard_cert" {
			name                = "example-cert-order"
			resource_group_name = "fake"
			location            = "global"
			distinguished_name  = "CN=example.com"
		}

		resource "azurerm_app_service_certificate_order" "wildcard_cert" {
			name                = "example-cert-order"
			resource_group_name = "fake"
			location            = "global"
			distinguished_name  = "CN=example.com"
			product_type        = "wildcard"
		}		
	`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "azurerm_app_service_certificate_order.standard_cert",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "SSL certificate (Standard)",
					PriceHash:        "4e8a819ae89e667ac4c8e5a86c983d49-e1f24f9fc7676b8cc310519e3f060f1d",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1).Div(decimal.NewFromInt(12))),
				},
			},
		},
		{
			Name: "azurerm_app_service_certificate_order.wildcard_cert",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "SSL certificate (wildcard)",
					PriceHash:        "6ec315d71563d4fc7369d88c8869765f-e1f24f9fc7676b8cc310519e3f060f1d",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1).Div(decimal.NewFromInt(12))),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
