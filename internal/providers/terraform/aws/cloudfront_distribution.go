package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getCloudfrontDistributionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_cloudfront_distribution",
		RFunc: newCloudfrontDistribution,
	}
}
func newCloudfrontDistribution(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	isOriginShieldEnabled := d.Get("origin.0.origin_shield.0.enabled").Bool()
	isSSLSupportMethodVIP := d.Get("viewer_certificate.0.ssl_support_method").String() == "vip"
	hasLoggingConfigBucket := !d.IsEmpty("logging_config.0.bucket")
	hasFieldLevelEncryptionID := !d.IsEmpty("default_cache_behavior.0.field_level_encryption_id")
	originShieldRegion := d.Get("origin.0.origin_shield.0.origin_shield_region").String()

	r := &aws.CloudfrontDistribution{
		Address:                   d.Address,
		Region:                    region,
		IsOriginShieldEnabled:     isOriginShieldEnabled,
		IsSSLSupportMethodVIP:     isSSLSupportMethodVIP,
		HasLoggingConfigBucket:    hasLoggingConfigBucket,
		HasFieldLevelEncryptionID: hasFieldLevelEncryptionID,
		OriginShieldRegion:        originShieldRegion,
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
