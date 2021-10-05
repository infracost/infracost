package aws

import (
	"fmt"
	"sort"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/schema"
)

func GetS3BucketRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name: "aws_s3_bucket",
		Notes: []string{
			"S3 replication time control data transfer, and batch operations are not supported by Terraform.",
		},
		RFunc: NewS3Bucket,
	}
}

func NewS3Bucket(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:           d.Address,
		SubResources:   s3SubResources(d, u),
		CostComponents: s3CostComponents(d, u),
	}
}

func s3CostComponents(d *schema.ResourceData, u *schema.UsageData) []*schema.CostComponent {
	region := d.Get("region").String()

	costComponents := make([]*schema.CostComponent, 0)

	objTagsEnabled := false

	for _, rule := range d.Get("lifecycle_rule").Array() {
		if len(rule.Get("tags").Map()) > 0 {
			objTagsEnabled = true
		}
	}

	var objTags *decimal.Decimal
	if u != nil && u.Get("object_tags").Exists() {
		objTags = decimalPtr(decimal.NewFromInt(u.Get("object_tags").Int()))
	}

	if objTagsEnabled {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Object tagging",
			Unit:            "10k tags",
			UnitMultiplier:  decimal.NewFromInt(10000),
			MonthlyQuantity: objTags,
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

func s3SubResources(d *schema.ResourceData, u *schema.UsageData) []*schema.Resource {
	region := d.Get("region").String()

	subResourceMap := make(map[string]*schema.Resource)

	subResourceMap["Standard"] = s3ResourceForStorageClass(region, "STANDARD", u)

	for _, rule := range d.Get("lifecycle_rule").Array() {
		if !rule.Get("enabled").Bool() {
			continue
		}
		for _, t := range rule.Get("transition").Array() {
			storageClass := t.Get("storage_class").String()
			if _, ok := subResourceMap[storageClass]; !ok {
				s := s3ResourceForStorageClass(region, storageClass, u)
				subResourceMap[s.Name] = s
			}
		}

		for _, t := range rule.Get("noncurrent_version_transition").Array() {
			if !rule.Get("enabled").Bool() {
				continue
			}
			storageClass := t.Get("storage_class").String()
			if _, ok := subResourceMap[storageClass]; !ok {
				s := s3ResourceForStorageClass(region, storageClass, u)
				if s != nil {
					subResourceMap[s.Name] = s
				}
			}
		}
	}

	if u != nil {
		if subResourceMap["Intelligent tiering"] == nil {
			if _, ok := u.Get("intelligent_tiering").Map()["frequent_access_storage_gb"]; ok {
				subResourceMap["Intelligent tiering"] = s3ResourceForStorageClass(region, "INTELLIGENT_TIERING", u)
			}
		}

		if subResourceMap["Standard - infrequent access"] == nil {
			if _, ok := u.Get("standard_infrequent_access").Map()["storage_gb"]; ok {
				subResourceMap["Standard - infrequent access"] = s3ResourceForStorageClass(region, "STANDARD_IA", u)
			}
		}

		if subResourceMap["One zone - infrequent access"] == nil {
			if _, ok := u.Get("one_zone_infrequent_access").Map()["storage_gb"]; ok {
				subResourceMap["One zone - infrequent access"] = s3ResourceForStorageClass(region, "ONEZONE_IA", u)
			}
		}

		if subResourceMap["Glacier"] == nil {
			if _, ok := u.Get("glacier").Map()["storage_gb"]; ok {
				subResourceMap["Glacier"] = s3ResourceForStorageClass(region, "GLACIER", u)
			}
		}

		if subResourceMap["Glacier deep archive"] == nil {
			if _, ok := u.Get("glacier_deep_archive").Map()["storage_gb"]; ok {
				subResourceMap["Glacier deep archive"] = s3ResourceForStorageClass(region, "DEEP_ARCHIVE", u)
			}
		}
	}

	subResources := make([]*schema.Resource, 0, len(subResourceMap))
	for _, s := range subResourceMap {
		subResources = append(subResources, s)
	}

	// Sort so we get consistent output (map iteration returns things in random order)
	sort.Slice(subResources, func(i, j int) bool {
		return subResources[i].Name < subResources[j].Name
	})

	return subResources
}

func s3ResourceForStorageClass(region string, storageClass string, u *schema.UsageData) *schema.Resource {
	switch storageClass {
	case "STANDARD":
		standardUsage := map[string]gjson.Result{}
		if u != nil && u.Get("standard").Exists() {
			standardUsage = u.Get("standard").Map()
		}

		var dataStorage *decimal.Decimal
		if v, ok := standardUsage["storage_gb"]; ok {
			dataStorage = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var pcplRequests *decimal.Decimal
		if v, ok := standardUsage["monthly_tier_1_requests"]; ok {
			pcplRequests = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var allOtherRequests *decimal.Decimal
		if v, ok := standardUsage["monthly_tier_2_requests"]; ok {
			allOtherRequests = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var dataScanned *decimal.Decimal
		if v, ok := standardUsage["monthly_select_data_scanned_gb"]; ok {
			dataScanned = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var dataReturned *decimal.Decimal
		if v, ok := standardUsage["monthly_select_data_returned_gb"]; ok {
			dataReturned = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		return &schema.Resource{
			Name: "Standard",
			CostComponents: []*schema.CostComponent{
				s3StorageVolumeTypeCostComponent("Storage", "AmazonS3", region, "TimedStorage-ByteHrs", "Standard", dataStorage),
				s3ApiCostComponent("PUT, COPY, POST, LIST requests", "AmazonS3", region, "Requests-Tier1", pcplRequests),
				s3ApiCostComponent("GET, SELECT, and all other requests", "AmazonS3", region, "Requests-Tier2", allOtherRequests),
				s3DataGroupCostComponent("Select data scanned", "AmazonS3", region, "Select-Scanned-Bytes", "S3-API-Select-Scanned", dataScanned),
				s3DataGroupCostComponent("Select data returned", "AmazonS3", region, "Select-Returned-Bytes", "S3-API-Select-Returned", dataReturned),
			},
		}
	case "INTELLIGENT_TIERING":
		inteliUsage := map[string]gjson.Result{}
		if u != nil && u.Get("intelligent_tiering").Exists() {
			inteliUsage = u.Get("intelligent_tiering").Map()
		}

		var frequentDataStorage *decimal.Decimal
		if v, ok := inteliUsage["frequent_access_storage_gb"]; ok {
			frequentDataStorage = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var infrequentDataStorage *decimal.Decimal
		if v, ok := inteliUsage["infrequent_access_storage_gb"]; ok {
			infrequentDataStorage = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var monitAutoObg *decimal.Decimal
		if v, ok := inteliUsage["monitored_objects"]; ok {
			monitAutoObg = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var pcplRequests *decimal.Decimal
		if v, ok := inteliUsage["monthly_tier_1_requests"]; ok {
			pcplRequests = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var allOtherRequests *decimal.Decimal
		if v, ok := inteliUsage["monthly_tier_2_requests"]; ok {
			allOtherRequests = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var lifecycleRequests *decimal.Decimal
		if v, ok := inteliUsage["monthly_lifecycle_transition_requests"]; ok {
			lifecycleRequests = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var dataScanned *decimal.Decimal
		if v, ok := inteliUsage["monthly_select_data_scanned_gb"]; ok {
			dataScanned = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var dataReturned *decimal.Decimal
		if v, ok := inteliUsage["monthly_select_data_returned_gb"]; ok {
			dataReturned = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var earlyDeletedData *decimal.Decimal
		if v, ok := inteliUsage["early_delete_gb"]; ok {
			earlyDeletedData = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		return &schema.Resource{
			Name: "Intelligent tiering",
			CostComponents: []*schema.CostComponent{
				s3StorageCostComponent("Storage (frequent access)", "AmazonS3", region, "TimedStorage-INT-FA-ByteHrs", frequentDataStorage),
				s3StorageCostComponent("Storage (infrequent access)", "AmazonS3", region, "TimedStorage-INT-IA-ByteHrs", infrequentDataStorage),
				s3MonitoringCostComponent(region, monitAutoObg),
				s3ApiCostComponent("PUT, COPY, POST, LIST requests", "AmazonS3", region, "Requests-INT-Tier1", pcplRequests),
				s3ApiCostComponent("GET, SELECT, and all other requests", "AmazonS3", region, "Requests-INT-Tier2", allOtherRequests),
				s3LifecycleTransitionsCostComponent(region, "Requests-Tier4", "", lifecycleRequests),
				s3DataCostComponent("Select data scanned", "AmazonS3", region, "Select-Scanned-INT-Bytes", dataScanned),
				s3DataCostComponent("Select data returned", "AmazonS3", region, "Select-Returned-INT-Bytes", dataReturned),
				s3DataCostComponent("Early delete (within 30 days)", "AmazonS3", region, "EarlyDelete-INT", earlyDeletedData),
			},
		}
	case "STANDARD_IA":
		standardIAUsage := map[string]gjson.Result{}
		if u != nil && u.Get("standard_infrequent_access").Exists() {
			standardIAUsage = u.Get("standard_infrequent_access").Map()
		}

		var dataStorage *decimal.Decimal
		if v, ok := standardIAUsage["storage_gb"]; ok {
			dataStorage = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var pcplRequests *decimal.Decimal
		if v, ok := standardIAUsage["monthly_tier_1_requests"]; ok {
			pcplRequests = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var allOtherRequests *decimal.Decimal
		if v, ok := standardIAUsage["monthly_tier_2_requests"]; ok {
			allOtherRequests = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var lifecycleRequests *decimal.Decimal
		if v, ok := standardIAUsage["monthly_lifecycle_transition_requests"]; ok {
			lifecycleRequests = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var retrievalData *decimal.Decimal
		if v, ok := standardIAUsage["monthly_retrieval_gb"]; ok {
			retrievalData = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var dataScanned *decimal.Decimal
		if v, ok := standardIAUsage["monthly_select_data_scanned_gb"]; ok {
			dataScanned = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var dataReturned *decimal.Decimal
		if v, ok := standardIAUsage["monthly_select_data_returned_gb"]; ok {
			dataReturned = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		return &schema.Resource{
			Name: "Standard - infrequent access",
			CostComponents: []*schema.CostComponent{
				s3StorageCostComponent("Storage", "AmazonS3", region, "TimedStorage-SIA-ByteHrs", dataStorage),
				s3ApiCostComponent("PUT, COPY, POST, LIST requests", "AmazonS3", region, "Requests-SIA-Tier1", pcplRequests),
				s3ApiCostComponent("GET, SELECT, and all other requests", "AmazonS3", region, "Requests-SIA-Tier2", allOtherRequests),
				s3LifecycleTransitionsCostComponent(region, "Requests-Tier4", "", lifecycleRequests),
				s3DataCostComponent("Retrievals", "AmazonS3", region, "Retrieval-SIA", retrievalData),
				s3DataCostComponent("Select data scanned", "AmazonS3", region, "Select-Scanned-SIA-Bytes", dataScanned),
				s3DataCostComponent("Select data returned", "AmazonS3", region, "Select-Returned-SIA-Bytes", dataReturned),
			},
		}
	case "ONEZONE_IA":
		onezoneIAUsage := map[string]gjson.Result{}
		if u != nil && u.Get("one_zone_infrequent_access").Exists() {
			onezoneIAUsage = u.Get("one_zone_infrequent_access").Map()
		}

		var dataStorage *decimal.Decimal
		if v, ok := onezoneIAUsage["storage_gb"]; ok {
			dataStorage = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var pcplRequests *decimal.Decimal
		if v, ok := onezoneIAUsage["monthly_tier_1_requests"]; ok {
			pcplRequests = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var allOtherRequests *decimal.Decimal
		if v, ok := onezoneIAUsage["monthly_tier_2_requests"]; ok {
			allOtherRequests = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var lifecycleRequests *decimal.Decimal
		if v, ok := onezoneIAUsage["monthly_lifecycle_transition_requests"]; ok {
			lifecycleRequests = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var retrievalData *decimal.Decimal
		if v, ok := onezoneIAUsage["monthly_retrieval_gb"]; ok {
			retrievalData = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var dataScanned *decimal.Decimal
		if v, ok := onezoneIAUsage["monthly_select_data_scanned_gb"]; ok {
			dataScanned = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var dataReturned *decimal.Decimal
		if v, ok := onezoneIAUsage["monthly_select_data_returned_gb"]; ok {
			dataReturned = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		return &schema.Resource{
			Name: "One zone - infrequent access",
			CostComponents: []*schema.CostComponent{
				s3StorageCostComponent("Storage", "AmazonS3", region, "TimedStorage-ZIA-ByteHrs", dataStorage),
				s3ApiCostComponent("PUT, COPY, POST, LIST requests", "AmazonS3", region, "Requests-ZIA-Tier1", pcplRequests),
				s3ApiCostComponent("GET, SELECT, and all other requests", "AmazonS3", region, "Requests-ZIA-Tier2", allOtherRequests),
				s3LifecycleTransitionsCostComponent(region, "Requests-Tier4", "", lifecycleRequests),
				s3DataCostComponent("Retrievals", "AmazonS3", region, "Retrieval-ZIA", retrievalData),
				s3DataCostComponent("Select data scanned", "AmazonS3", region, "Select-Scanned-ZIA-Bytes", dataScanned),
				s3DataCostComponent("Select data returned", "AmazonS3", region, "Select-Returned-ZIA-Bytes", dataReturned),
			},
		}
	case "GLACIER":
		glacierUsage := map[string]gjson.Result{}
		if u != nil && u.Get("glacier").Exists() {
			glacierUsage = u.Get("glacier").Map()
		}

		var dataStorage *decimal.Decimal
		if v, ok := glacierUsage["storage_gb"]; ok {
			dataStorage = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var pcplRequests *decimal.Decimal
		if v, ok := glacierUsage["monthly_tier_1_requests"]; ok {
			pcplRequests = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var allOtherRequests *decimal.Decimal
		if v, ok := glacierUsage["monthly_tier_2_requests"]; ok {
			allOtherRequests = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var lifecycleRequests *decimal.Decimal
		if v, ok := glacierUsage["monthly_lifecycle_transition_requests"]; ok {
			lifecycleRequests = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var earlyDeletedData *decimal.Decimal
		if v, ok := glacierUsage["early_delete_gb"]; ok {
			earlyDeletedData = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var stdDataScanned *decimal.Decimal
		if v, ok := glacierUsage["monthly_standard_select_data_scanned_gb"]; ok {
			stdDataScanned = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var stdDataReturned *decimal.Decimal
		if v, ok := glacierUsage["monthly_standard_select_data_returned_gb"]; ok {
			stdDataReturned = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var bulkDataScanned *decimal.Decimal
		if v, ok := glacierUsage["monthly_bulk_select_data_scanned_gb"]; ok {
			bulkDataScanned = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var bulkDataReturned *decimal.Decimal
		if v, ok := glacierUsage["monthly_bulk_select_data_returned_gb"]; ok {
			bulkDataReturned = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var expDataScanned *decimal.Decimal
		if v, ok := glacierUsage["monthly_expedited_select_data_scanned_gb"]; ok {
			expDataScanned = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var expDataReturned *decimal.Decimal
		if v, ok := glacierUsage["monthly_expedited_select_data_returned_gb"]; ok {
			expDataReturned = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var stdRetrievalData *decimal.Decimal
		if v, ok := glacierUsage["monthly_standard_data_retrieval_gb"]; ok {
			stdRetrievalData = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var stdRetrievalReq *decimal.Decimal
		if v, ok := glacierUsage["monthly_standard_data_retrieval_requests"]; ok {
			stdRetrievalReq = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var bulkRetrievalData *decimal.Decimal
		if v, ok := glacierUsage["monthly_bulk_data_retrieval_gb"]; ok {
			bulkRetrievalData = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var bulkRetrievalReq *decimal.Decimal
		if v, ok := glacierUsage["monthly_bulk_data_retrieval_requests"]; ok {
			bulkRetrievalReq = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var expRetrievalData *decimal.Decimal
		if v, ok := glacierUsage["monthly_expedited_data_retrieval_gb"]; ok {
			expRetrievalData = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var expRetrievalReq *decimal.Decimal
		if v, ok := glacierUsage["monthly_expedited_data_retrieval_requests"]; ok {
			expRetrievalReq = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		return &schema.Resource{
			Name: "Glacier",
			CostComponents: []*schema.CostComponent{
				s3StorageCostComponent("Storage", "AmazonGlacier", region, "TimedStorage-ByteHrs", dataStorage),
				s3ApiOperationCostComponent("PUT, COPY, POST, LIST requests", "AmazonS3", region, "Requests-GLACIER-Tier1", "PostObject", pcplRequests),
				s3ApiCostComponent("GET, SELECT, and all other requests", "AmazonS3", region, "Requests-GLACIER-Tier2", allOtherRequests),
				s3LifecycleTransitionsCostComponent(region, "Requests-Tier3", "S3-GlacierTransition", lifecycleRequests),
				s3ApiCostComponent("Retrieval requests (standard)", "AmazonGlacier", region, "Requests-Tier1", stdRetrievalReq),
				s3DataCostComponent("Retrievals (standard)", "AmazonGlacier", region, "Standard-Retrieval-Bytes", stdRetrievalData),
				s3DataCostComponent("Select data scanned (standard)", "AmazonGlacier", region, "Std-Select-Scanned-Bytes", stdDataScanned),
				s3DataCostComponent("Select data returned (standard)", "AmazonGlacier", region, "Std-Select-Returned-Bytes", stdDataReturned),
				s3ApiCostComponent("Retrieval requests (expedited)", "AmazonGlacier", region, "Requests-Tier3", expRetrievalReq),
				s3DataCostComponent("Retrievals (expedited)", "AmazonGlacier", region, "Expedited-Retrieval-Bytes", expRetrievalData),
				s3DataCostComponent("Select data scanned (expedited)", "AmazonGlacier", region, "Exp-Select-Scanned-Bytes", expDataScanned),
				s3DataCostComponent("Select data returned (expedited)", "AmazonGlacier", region, "Exp-Select-Returned-Bytes", expDataReturned),
				s3ApiCostComponent("Retrieval requests (bulk)", "AmazonGlacier", region, "Requests-Tier2", bulkRetrievalReq),
				s3DataCostComponent("Retrievals (bulk)", "AmazonGlacier", region, "Bulk-Retrieval-Bytes", bulkRetrievalData),
				s3DataCostComponent("Select data scanned (bulk)", "AmazonGlacier", region, "Bulk-Select-Scanned-Bytes", bulkDataScanned),
				s3DataCostComponent("Select data returned (bulk)", "AmazonGlacier", region, "Bulk-Select-Returned-Bytes", bulkDataReturned),
				s3DataCostComponent("Early delete (within 90 days)", "AmazonGlacier", region, "EarlyDelete-ByteHrs", earlyDeletedData),
			},
		}
	case "DEEP_ARCHIVE":
		daUsage := map[string]gjson.Result{}
		if u != nil && u.Get("glacier_deep_archive").Exists() {
			daUsage = u.Get("glacier_deep_archive").Map()
		}

		var dataStorage *decimal.Decimal
		if v, ok := daUsage["storage_gb"]; ok {
			dataStorage = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var pcplRequests *decimal.Decimal
		if v, ok := daUsage["monthly_tier_1_requests"]; ok {
			pcplRequests = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var allOtherRequests *decimal.Decimal
		if v, ok := daUsage["monthly_tier_2_requests"]; ok {
			allOtherRequests = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var lifecycleRequests *decimal.Decimal
		if v, ok := daUsage["monthly_lifecycle_transition_requests"]; ok {
			lifecycleRequests = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var stdRetrievalData *decimal.Decimal
		if v, ok := daUsage["monthly_standard_data_retrieval_gb"]; ok {
			stdRetrievalData = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var stdRetrievalReq *decimal.Decimal
		if v, ok := daUsage["monthly_standard_data_retrieval_requests"]; ok {
			stdRetrievalReq = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var bulkRetrievalData *decimal.Decimal
		if v, ok := daUsage["monthly_bulk_data_retrieval_gb"]; ok {
			bulkRetrievalData = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var bulkRetrievalReq *decimal.Decimal
		if v, ok := daUsage["monthly_bulk_data_retrieval_requests"]; ok {
			bulkRetrievalReq = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		var earlyDeletedData *decimal.Decimal
		if v, ok := daUsage["early_delete_gb"]; ok {
			earlyDeletedData = decimalPtr(decimal.NewFromInt(v.Int()))
		}

		return &schema.Resource{
			Name: "Glacier deep archive",
			CostComponents: []*schema.CostComponent{
				s3StorageCostComponent("Storage", "AmazonS3GlacierDeepArchive", region, "TimedStorage-GDA-ByteHrs", dataStorage),
				s3ApiOperationCostComponent("PUT, COPY, POST, LIST requests", "AmazonS3GlacierDeepArchive", region, "Requests-GDA-Tier1", "PostObject", pcplRequests),
				s3ApiCostComponent("GET, SELECT, and all other requests", "AmazonS3", region, "Requests-GDA-Tier2", allOtherRequests),
				s3LifecycleTransitionsCostComponent(region, "Requests-Tier3", "S3-GDATransition", lifecycleRequests),
				s3ApiOperationCostComponent("Retrieval requests (standard)", "AmazonS3GlacierDeepArchive", region, "Requests-GDA-Tier3", "", stdRetrievalReq),
				s3DataCostComponent("Retrievals (standard)", "AmazonS3GlacierDeepArchive", region, "Standard-Retrieval-Bytes", stdRetrievalData),
				s3ApiCostComponent("Retrieval requests (bulk)", "AmazonS3GlacierDeepArchive", region, "Requests-GDA-Tier5", bulkRetrievalReq),
				s3DataCostComponent("Retrievals (bulk)", "AmazonS3GlacierDeepArchive", region, "Bulk-Retrieval-Bytes", bulkRetrievalData),
				s3DataCostComponent("Early delete (within 180 days)", "AmazonS3GlacierDeepArchive", region, "EarlyDelete-GDA", earlyDeletedData),
			},
		}
	default:
		return nil
	}
}

func s3StorageCostComponent(name string, service string, region string, usageType string, dataStorage *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: dataStorage,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr(service),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", usageType))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("0"),
		},
	}
}

func s3StorageVolumeTypeCostComponent(name string, service string, region string, usageType string, volumeType string, dataStorage *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: dataStorage,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr(service),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", usageType))},
				{Key: "volumeType", ValueRegex: strPtr(fmt.Sprintf("/%s/i", volumeType))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("0"),
		},
	}
}

func s3ApiCostComponent(name string, service string, region string, usageType string, requests *decimal.Decimal) *schema.CostComponent {
	return s3ApiOperationCostComponent(name, service, region, usageType, "", requests)
}

func s3ApiOperationCostComponent(name string, service string, region string, usageType string, operation string, requests *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "1k requests",
		UnitMultiplier:  decimal.NewFromInt(1000),
		MonthlyQuantity: requests,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr(service),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", usageType))},
				{Key: "operation", ValueRegex: strPtr(fmt.Sprintf("/%s/i", operation))},
			},
		},
	}
}

func s3DataCostComponent(name string, service string, region string, usageType string, data *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: data,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr(service),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", usageType))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("0"),
		},
	}
}

func s3DataGroupCostComponent(name string, service string, region string, usageType string, group string, data *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: data,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr(service),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", usageType))},
				{Key: "group", ValueRegex: strPtr(fmt.Sprintf("/%s/i", group))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("0"),
		},
	}
}

func s3LifecycleTransitionsCostComponent(region string, usageType string, operation string, requests *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Lifecycle transition",
		Unit:            "1k requests",
		UnitMultiplier:  decimal.NewFromInt(1000),
		MonthlyQuantity: requests,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr("AmazonS3"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", usageType))},
				{Key: "operation", ValueRegex: strPtr(fmt.Sprintf("/%s/i", operation))},
			},
		},
	}
}

func s3MonitoringCostComponent(region string, objects *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Monitoring and automation",
		Unit:            "1k objects",
		UnitMultiplier:  decimal.NewFromInt(1000),
		MonthlyQuantity: objects,
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
