package aws

import (
	"github.com/infracost/infracost/internal/schema"
)

func getS3BucketVersioningRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_s3_bucket_versioning",
		RFunc:               NewS3BucketVersioning,
		ReferenceAttributes: []string{"bucket"},
	}
}

func NewS3BucketVersioning(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
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
