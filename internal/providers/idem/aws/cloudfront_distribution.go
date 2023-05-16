package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetCloudfrontDistributionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "states.aws.cloudfront.distribution.present",
		RFunc:               newCloudfrontDistribution,
		ReferenceAttributes: []string{},
	}
}
func newCloudfrontDistribution(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	isOriginShieldEnabled := d.Get("origins.Items.0.OriginShield.Enabled").Bool()
	isSSLSupportMethodVIP := d.Get("viewer_certificate.SSLSupportMethod").String() == "vip"
	hasLoggingConfigBucket := !d.IsEmpty("logging.Bucket")
	hasFieldLevelEncryptionID := !d.IsEmpty("default_cache_behaviour.FieldLevelEncryptionId")
	//originShieldRegion := d.Get("origin.0.origin_shield.0.origin_shield_region").String() # Currently we do not have origin shield region

	r := &aws.CloudfrontDistribution{
		Address:                   d.Address,
		Region:                    region,
		IsOriginShieldEnabled:     isOriginShieldEnabled,
		IsSSLSupportMethodVIP:     isSSLSupportMethodVIP,
		HasLoggingConfigBucket:    hasLoggingConfigBucket,
		HasFieldLevelEncryptionID: hasFieldLevelEncryptionID,
		//OriginShieldRegion:        originShieldRegion,
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
