package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getS3BucketAnalyticsConfigurationRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_s3_bucket_analytics_configuration",
		CoreRFunc: NewS3BucketAnalyticsConfiguration,
	}
}

func NewS3BucketAnalyticsConfiguration(d *schema.ResourceData) schema.CoreResource {
	r := &aws.S3BucketAnalyticsConfiguration{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}
	return r
}
