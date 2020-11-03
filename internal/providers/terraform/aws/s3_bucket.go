package aws

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
)

func GetS3BucketRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_s3_bucket",
		RFunc: NewS3Bucket,
	}
}

func NewS3Bucket(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	return &schema.Resource{
		Name:           d.Address,
		SubResources:   s3SubResources(d),
		CostComponents: s3CostComponents(d),
	}
}

func s3CostComponents(d *schema.ResourceData) []*schema.CostComponent {
	region := d.Get("region").String()

	costComponents := make([]*schema.CostComponent, 0)

	objTagsEnabled := false

	for _, rule := range d.Get("lifecycle_rule").Array() {
		if len(rule.Get("tags").Map()) > 0 {
			objTagsEnabled = true
		}
	}

	if objTagsEnabled {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:           "Object tagging",
			Unit:           "tags",
			UnitMultiplier: 10000,
			ProductFilter: &schema.ProductFilter{
				VendorName: strPtr("aws"),
				Region:     strPtr(region),
				Service:    strPtr("AmazonS3"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/TagStorage-TagHrs/")},
				},
			},
		})
	}

	return costComponents
}

func s3SubResources(d *schema.ResourceData) []*schema.Resource {
	region := d.Get("region").String()

	subResourceMap := make(map[string]*schema.Resource)

	subResourceMap["Standard"] = s3ResourceForStorageClass(region, "STANDARD")

	for _, rule := range d.Get("lifecycle_rule").Array() {
		if !rule.Get("enabled").Bool() {
			continue
		}
		for _, t := range rule.Get("transition").Array() {
			storageClass := t.Get("storage_class").String()
			if _, ok := subResourceMap[storageClass]; !ok {
				s := s3ResourceForStorageClass(region, storageClass)
				subResourceMap[s.Name] = s
			}
		}

		for _, t := range rule.Get("noncurrent_version_transition").Array() {
			if !rule.Get("enabled").Bool() {
				continue
			}
			storageClass := t.Get("storage_class").String()
			if _, ok := subResourceMap[storageClass]; !ok {
				s := s3ResourceForStorageClass(region, storageClass)
				if s != nil {
					subResourceMap[s.Name] = s
				}
			}
		}
	}

	subResources := make([]*schema.Resource, 0, len(subResourceMap))
	for _, s := range subResourceMap {
		subResources = append(subResources, s)
	}

	return subResources
}

func s3ResourceForStorageClass(region string, storageClass string) *schema.Resource {
	switch storageClass {
	case "STANDARD":
		return &schema.Resource{
			Name: "Standard",
			CostComponents: []*schema.CostComponent{
				s3StorageVolumeTypeCostComponent("Storage", "AmazonS3", region, "TimedStorage-ByteHrs", "Standard"),
				s3ApiCostComponent("PUT, COPY, POST, LIST requests", "AmazonS3", region, "Requests-Tier1"),
				s3ApiCostComponent("GET, SELECT, and all other requests", "AmazonS3", region, "Requests-Tier2"),
				s3DataGroupCostComponent("Select data scanned", "AmazonS3", region, "Select-Scanned-Bytes", "S3-API-Select-Scanned"),
				s3DataGroupCostComponent("Select data returned", "AmazonS3", region, "Select-Returned-Bytes", "S3-API-Select-Returned"),
			},
		}
	case "INTELLIGENT_TIERING":
		return &schema.Resource{
			Name: "Intelligent tiering",
			CostComponents: []*schema.CostComponent{
				s3StorageCostComponent("Storage (frequent access)", "AmazonS3", region, "TimedStorage-INT-FA-ByteHrs"),
				s3StorageCostComponent("Storage (infrequent access)", "AmazonS3", region, "TimedStorage-INT-IA-ByteHrs"),
				s3MonitoringCostComponent(region),
				s3ApiCostComponent("PUT, COPY, POST, LIST requests", "AmazonS3", region, "Requests-INT-Tier1"),
				s3ApiCostComponent("GET, SELECT, and all other requests", "AmazonS3", region, "Requests-INT-Tier2"),
				s3LifecycleTransitionsCostComponent(region, "Requests-Tier4", ""),
				s3DataCostComponent("Select data scanned", "AmazonS3", region, "Select-Scanned-INT-Bytes"),
				s3DataCostComponent("Select data returned", "AmazonS3", region, "Select-Returned-INT-Bytes"),
				s3DataCostComponent("Early delete (within 30 days)", "AmazonS3", region, "EarlyDelete-INT"),
			},
		}
	case "STANDARD_IA":
		return &schema.Resource{
			Name: "Standard - infrequent access",
			CostComponents: []*schema.CostComponent{
				s3StorageCostComponent("Storage", "AmazonS3", region, "TimedStorage-SIA-ByteHrs"),
				s3ApiCostComponent("PUT, COPY, POST, LIST requests", "AmazonS3", region, "Requests-SIA-Tier1"),
				s3ApiCostComponent("GET, SELECT, and all other requests", "AmazonS3", region, "Requests-SIA-Tier2"),
				s3LifecycleTransitionsCostComponent(region, "Requests-Tier4", ""),
				s3DataCostComponent("Retrievals", "AmazonS3", region, "Retrieval-SIA"),
				s3DataCostComponent("Select data scanned", "AmazonS3", region, "Select-Scanned-SIA-Bytes"),
				s3DataCostComponent("Select data returned", "AmazonS3", region, "Select-Returned-SIA-Bytes"),
			},
		}
	case "ONEZONE_IA":
		return &schema.Resource{
			Name: "One zone - infrequent access",
			CostComponents: []*schema.CostComponent{
				s3StorageCostComponent("Storage", "AmazonS3", region, "TimedStorage-ZIA-ByteHrs"),
				s3ApiCostComponent("PUT, COPY, POST, LIST requests", "AmazonS3", region, "Requests-ZIA-Tier1"),
				s3ApiCostComponent("GET, SELECT, and all other requests", "AmazonS3", region, "Requests-ZIA-Tier2"),
				s3LifecycleTransitionsCostComponent(region, "Requests-Tier4", ""),
				s3DataCostComponent("Retrievals", "AmazonS3", region, "Retrieval-ZIA"),
				s3DataCostComponent("Select data scanned", "AmazonS3", region, "Select-Scanned-ZIA-Bytes"),
				s3DataCostComponent("Select data returned", "AmazonS3", region, "Select-Returned-ZIA-Bytes"),
			},
		}
	case "GLACIER":
		return &schema.Resource{
			Name: "Glacier",
			CostComponents: []*schema.CostComponent{
				s3StorageCostComponent("Storage", "AmazonGlacier", region, "TimedStorage-ByteHrs"),
				s3ApiOperationCostComponent("PUT, COPY, POST, LIST requests", "AmazonS3", region, "Requests-GLACIER-Tier1", "PostObject"),
				s3ApiCostComponent("GET, SELECT, and all other requests", "AmazonS3", region, "Requests-GLACIER-Tier2"),
				s3LifecycleTransitionsCostComponent(region, "Requests-Tier3", "S3-GlacierTransition"),
				s3ApiCostComponent("Retrieval requests (standard)", "AmazonGlacier", region, "Requests-Tier1"),
				s3DataCostComponent("Retrievals (standard)", "AmazonGlacier", region, "Standard-Retrieval-Bytes"),
				s3DataCostComponent("Select data scanned (standard)", "AmazonGlacier", region, "Std-Select-Scanned-Bytes"),
				s3DataCostComponent("Select data returned (standard)", "AmazonGlacier", region, "Std-Select-Returned-Bytes"),
				s3ApiCostComponent("Retrieval requests (expedited)", "AmazonGlacier", region, "Requests-Tier3"),
				s3DataCostComponent("Retrievals (expedited)", "AmazonGlacier", region, "Expedited-Retrieval-Bytes"),
				s3DataCostComponent("Select data scanned (expedited)", "AmazonGlacier", region, "Exp-Select-Scanned-Bytes"),
				s3DataCostComponent("Select data returned (expedited)", "AmazonGlacier", region, "Exp-Select-Returned-Bytes"),
				s3ApiCostComponent("Retrieval requests (bulk)", "AmazonGlacier", region, "Requests-Tier2"),
				s3DataCostComponent("Retrievals (bulk)", "AmazonGlacier", region, "Bulk-Retrieval-Bytes"),
				s3DataCostComponent("Select data scanned (bulk)", "AmazonGlacier", region, "Bulk-Select-Scanned-Bytes"),
				s3DataCostComponent("Select data returned (bulk)", "AmazonGlacier", region, "Bulk-Select-Returned-Bytes"),
				s3DataCostComponent("Early delete (within 90 days)", "AmazonGlacier", region, "EarlyDelete-ByteHrs"),
			},
		}
	case "DEEP_ARCHIVE":
		return &schema.Resource{
			Name: "Glacier deep archive",
			CostComponents: []*schema.CostComponent{
				s3StorageCostComponent("Storage", "AmazonS3GlacierDeepArchive", region, "TimedStorage-GDA-ByteHrs"),
				s3ApiOperationCostComponent("PUT, COPY, POST, LIST requests", "AmazonS3GlacierDeepArchive", region, "Requests-GDA-Tier1", "PostObject"),
				s3ApiCostComponent("GET, SELECT, and all other requests", "AmazonS3", region, "Requests-GDA-Tier2"),
				s3LifecycleTransitionsCostComponent(region, "Requests-Tier3", "S3-GDATransition"),
				s3ApiOperationCostComponent("Retrieval requests (standard)", "AmazonS3GlacierDeepArchive", region, "Requests-GDA-Tier3", "DeepArchiveRestoreObject"),
				s3DataCostComponent("Retrievals (standard)", "AmazonS3GlacierDeepArchive", region, "Standard-Retrieval-Bytes"),
				s3ApiCostComponent("Retrieval requests (bulk)", "AmazonS3GlacierDeepArchive", region, "Requests-GDA-Tier5"),
				s3DataCostComponent("Retrievals (bulk)", "AmazonS3GlacierDeepArchive", region, "Bulk-Retrieval-Bytes"),
				s3DataCostComponent("Early delete (within 180 days)", "AmazonS3GlacierDeepArchive", region, "EarlyDelete-GDA"),
			},
		}
	default:
		return nil
	}
}

func s3StorageCostComponent(name string, service string, region string, usageType string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           name,
		Unit:           "GB",
		UnitMultiplier: 1,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr(service),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/", usageType))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("0"),
		},
	}
}

func s3StorageVolumeTypeCostComponent(name string, service string, region string, usageType string, volumeType string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           name,
		Unit:           "GB",
		UnitMultiplier: 1,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr(service),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/", usageType))},
				{Key: "volumeType", Value: strPtr(volumeType)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("0"),
		},
	}
}

func s3ApiCostComponent(name string, service string, region string, usageType string) *schema.CostComponent {
	return s3ApiOperationCostComponent(name, service, region, usageType, "")
}

func s3ApiOperationCostComponent(name string, service string, region string, usageType string, operation string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           name,
		Unit:           "requests",
		UnitMultiplier: 1000,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr(service),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/", usageType))},
				{Key: "operation", Value: strPtr(operation)},
			},
		},
	}
}

func s3DataCostComponent(name string, service string, region string, usageType string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           name,
		Unit:           "GB",
		UnitMultiplier: 1,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr(service),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/", usageType))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("0"),
		},
	}
}

func s3DataGroupCostComponent(name string, service string, region string, usageType string, group string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           name,
		Unit:           "GB",
		UnitMultiplier: 1,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr(service),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/", usageType))},
				{Key: "group", Value: strPtr(group)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("0"),
		},
	}
}

func s3LifecycleTransitionsCostComponent(region string, usageType string, operation string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Lifecycle transition",
		Unit:           "requests",
		UnitMultiplier: 1000,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr("AmazonS3"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/", usageType))},
				{Key: "operation", Value: strPtr(operation)},
			},
		},
	}
}

func s3MonitoringCostComponent(region string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Monitoring and automation",
		Unit:           "objects",
		UnitMultiplier: 1000,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr("AmazonS3"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/Monitoring-Automation-INT/")},
			},
		},
	}
}
