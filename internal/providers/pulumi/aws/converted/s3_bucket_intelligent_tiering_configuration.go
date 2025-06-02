package aws

import (
	"github.com/infracost/infracost/internal/schema"
)

func getS3BucketIntelligentTieringConfigurationRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_s3_bucket_intelligent_tiering_configuration",
		RFunc:               NewS3BucketIntelligentTieringConfiguration,
		ReferenceAttributes: []string{"bucket"},
	}
}

func NewS3BucketIntelligentTieringConfiguration(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:         d.Address,
		ResourceType: d.Type,
		Tags:         d.Tags,
		DefaultTags:  d.DefaultTags,
		IsSkipped:    true,
		NoPrice:      true,
		SkipMessage:  "Free resource.",
	}
}
