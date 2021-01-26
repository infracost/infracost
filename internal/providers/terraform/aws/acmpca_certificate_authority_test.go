package aws_test

import (
	"github.com/shopspring/decimal"
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestSACMPCACertificateAuthorityFunction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
        resource "aws_acmpca_certificate_authority" "private_ca" {
			certificate_authority_configuration {
				key_algorithm = "RSA_4096"
				signing_algorithm = "SHA512WITHRSA"

    			subject {
      				common_name = "private-ca.com"
    			}
  			}
        }`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_acmpca_certificate_authority.private_ca",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Private certificate authority",
					PriceHash:        "cabb33509c029f80e12140cc33872027-2cdaeeb8115b7046007118c018b9f493",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestSACMPCACertificateAuthority_1000(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
        resource "aws_acmpca_certificate_authority" "private_ca" {
			certificate_authority_configuration {
				key_algorithm = "RSA_4096"
				signing_algorithm = "SHA512WITHRSA"

    			subject {
      				common_name = "private-ca.com"
    			}
  			}
        }`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_acmpca_certificate_authority.private_ca": map[string]interface{}{
			"monthly_certificates_created": 1000,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_acmpca_certificate_authority.private_ca",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Private certificate authority",
					PriceHash:        "cabb33509c029f80e12140cc33872027-2cdaeeb8115b7046007118c018b9f493",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Certificates (1 - 1K)",
					PriceHash:        "58517ba36a89b107d4f5088c1e6cb3b8-3634aef65032f056acf2f6091e2c0022",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1000)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
