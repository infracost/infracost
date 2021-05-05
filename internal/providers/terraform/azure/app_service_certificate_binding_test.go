package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestAzureRMAppServiceCertificateBinding(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "azurerm_app_service_certificate_binding" "ip_ssl" {
			hostname_binding_id = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Web/sites/mywebappfake/hostNameBindings/example.example.com"
			certificate_id      = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Web/certificates/example.example.com"
			ssl_state           = "IpBasedEnabled"
		}

		resource "azurerm_app_service_certificate_binding" "sni_ssl" {
			hostname_binding_id = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Web/sites/mywebappfake/hostNameBindings/example.example.com"
			certificate_id      = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Web/certificates/example.example.com"
			ssl_state           = "SniEnabled"
		}
	`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "azurerm_app_service_certificate_binding.ip_ssl",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "IP SSL certificate",
					PriceHash:        "9aaa15fe7dc8302f9046d95ba081c50a-e285791b6e6926c07354b58a33e7ecf4",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks, tmpDir)
}
