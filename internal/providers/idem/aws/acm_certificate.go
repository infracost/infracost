package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetACMCertificate() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "states.aws.acm.certificate_manager.present",
		RFunc: NewACMCertificate,
	}
}
func NewACMCertificate(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.ACMCertificate{
		Address:                 d.Address,
		Region:                  d.Get("region").String(),
		CertificateAuthorityARN: d.Get("certificate_authority_arn").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
