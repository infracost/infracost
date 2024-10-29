package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getACMPCACertificateAuthorityRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_acmpca_certificate_authority",
		CoreRFunc: NewACMPCACertificateAuthority,
	}
}
func NewACMPCACertificateAuthority(d *schema.ResourceData) schema.CoreResource {
	r := &aws.ACMPCACertificateAuthority{
		Address:   d.Address,
		Region:    d.Get("region").String(),
		UsageMode: d.Get("usage_mode").String(),
	}
	return r
}
