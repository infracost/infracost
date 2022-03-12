package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"

	"github.com/infracost/infracost/internal/schema"
)

func getS3BucketLifecycleConfigurationRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_s3_bucket_lifecycle_configuration",
		RFunc: newS3BucketLifecycleConfigurationResource,
	}
}

func newS3BucketLifecycleConfigurationResource(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	storageClassNames := map[string]string{
		"STANDARD":            "standard",
		"INTELLIGENT_TIERING": "intelligent_tiering",
		"STANDARD_IA":         "standard_infrequent_access",
		"ONEZONE_IA":          "one_zone_infrequent_access",
		"GLACIER":             "glacier_flexible_retrieval",
		"DEEP_ARCHIVE":        "glacier_deep_archive",
	}

	objTagsEnabled := false

	// Always add the standard storage class
	lifecycleStorageClassMap := map[string]bool{
		"standard": true,
	}

	for _, rule := range d.Get("rule").Array() {
		if rule.Get("status").String() != "Enabled" {
			continue
		}

		if len(rule.Get("filter.#.tag").Array()) > 0 || len(rule.Get("filter.#.and.#.tag").Array()) > 0 {
			objTagsEnabled = true
		}

		for _, t := range rule.Get("transition").Array() {
			storageClass := storageClassNames[t.Get("storage_class").String()]
			if _, ok := lifecycleStorageClassMap[storageClass]; !ok && storageClass != "" {
				lifecycleStorageClassMap[storageClass] = true
			}
		}

		for _, t := range rule.Get("noncurrent_version_transition").Array() {
			storageClass := storageClassNames[t.Get("storage_class").String()]
			if _, ok := lifecycleStorageClassMap[storageClass]; !ok && storageClass != "" {
				lifecycleStorageClassMap[storageClass] = true
			}
		}
	}

	lifecycleStorageClasses := make([]string, 0, len(lifecycleStorageClassMap))
	for storageClass := range lifecycleStorageClassMap {
		lifecycleStorageClasses = append(lifecycleStorageClasses, storageClass)
	}

	r := &aws.S3BucketLifecycleConfiguration{
		Address:                 d.Address,
		Region:                  d.Get("region").String(),
		Name:                    d.Get("bucket").String(),
		ObjectTagsEnabled:       objTagsEnabled,
		LifecycleStorageClasses: lifecycleStorageClasses,
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}
