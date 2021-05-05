package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestACMCertificateFunction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_acm_certificate" "private_cert" {
			domain_name = "private-ca.com"
  		certificate_authority_arn = "arn:aws:acm-pca:us-east-1:123456789012:certificate-authority/12345678-1234-1234-1234-123456789012"
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_acm_certificate.private_cert",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Certificate",
					PriceHash:        "58517ba36a89b107d4f5088c1e6cb3b8-3634aef65032f056acf2f6091e2c0022",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks, tmpDir)
}
