package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type ACMCertificate struct {
	Address                 *string
	Region                  *string
	CertificateAuthorityArn *string
}

var ACMCertificateUsageSchema = []*schema.UsageItem{}

func (r *ACMCertificate) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ACMCertificate) BuildResource() *schema.Resource {
	region := *r.Region

	if r.CertificateAuthorityArn != nil {
		one := decimal.NewFromInt(1)
		return &schema.Resource{
			Name: *r.Address,
			CostComponents: []*schema.CostComponent{
				certificateCostComponent(region, "Certificate", "0", &one),
			}, UsageSchema: ACMCertificateUsageSchema,
		}
	}

	return &schema.Resource{
		Name:      *r.Address,
		NoPrice:   true,
		IsSkipped: true, UsageSchema: ACMCertificateUsageSchema,
	}
}
