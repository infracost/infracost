package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type ACMCertificate struct {
	Address                 string
	Region                  string
	CertificateAuthorityARN string
}

var ACMCertificateUsageSchema = []*schema.UsageItem{}

func (r *ACMCertificate) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ACMCertificate) BuildResource() *schema.Resource {
	if r.CertificateAuthorityARN == "" {
		return &schema.Resource{
			Name:        r.Address,
			NoPrice:     true,
			IsSkipped:   true,
			UsageSchema: ACMCertificateUsageSchema,
		}
	}

	certAuthority := &ACMPCACertificateAuthority{
		Region: r.Region,
	}

	certCostComponent := certAuthority.certificateCostComponent("Certificate", "0", decimalPtr(decimal.NewFromInt(1)))

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: []*schema.CostComponent{certCostComponent},
		UsageSchema:    ACMCertificateUsageSchema,
	}
}
