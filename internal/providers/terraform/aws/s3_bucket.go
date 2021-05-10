package aws

import (
	"fmt"
	"sort"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
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
			Unit:            "tags",
			UnitMultiplier:  10000,
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
			if u.Get("intelligent_tiering.frequent_access_storage_gb").Exists() {
				subResourceMap["Intelligent tiering"] = s3ResourceForStorageClass(region, "INTELLIGENT_TIERING", u)
			}
		}

		if subResourceMap["Standard - infrequent access"] == nil {
			if u.Get("standard_infrequent_access.storage_gb").Exists() {
				subResourceMap["Standard - infrequent access"] = s3ResourceForStorageClass(region, "STANDARD_IA", u)
			}
		}

		if subResourceMap["One zone - infrequent access"] == nil {
			if u.Get("one_zone_infrequent_access.storage_gb").Exists() {
				subResourceMap["One zone - infrequent access"] = s3ResourceForStorageClass(region, "ONEZONE_IA", u)
			}
		}

		if subResourceMap["Glacier"] == nil {
			if u.Get("glacier.storage_gb").Exists() {
				subResourceMap["Glacier"] = s3ResourceForStorageClass(region, "GLACIER", u)
			}
		}

		if subResourceMap["Glacier deep archive"] == nil {
			if u.Get("glacier_deep_archive.storage_gb").Exists() {
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
		var dataStorage *decimal.Decimal
		if u != nil && u.Get("standard.storage_gb").Exists() {
			dataStorage = decimalPtr(decimal.NewFromInt(u.Get("standard.storage_gb").Int()))
		}

		var pcplRequests *decimal.Decimal
		if u != nil && u.Get("standard.monthly_tier_1_requests").Exists() {
			pcplRequests = decimalPtr(decimal.NewFromInt(u.Get("standard.monthly_tier_1_requests").Int()))
		}

		var allOtherRequests *decimal.Decimal
		if u != nil && u.Get("standard.monthly_tier_2_requests").Exists() {
			allOtherRequests = decimalPtr(decimal.NewFromInt(u.Get("standard.monthly_tier_2_requests").Int()))
		}

		var dataScanned *decimal.Decimal
		if u != nil && u.Get("standard.monthly_select_data_scanned_gb").Exists() {
			dataScanned = decimalPtr(decimal.NewFromInt(u.Get("standard.monthly_select_data_scanned_gb").Int()))
		}

		var dataReturned *decimal.Decimal
		if u != nil && u.Get("standard.monthly_select_data_returned_gb").Exists() {
			dataReturned = decimalPtr(decimal.NewFromInt(u.Get("standard.monthly_select_data_returned_gb").Int()))
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
		var frequentDataStorage *decimal.Decimal
		if u != nil && u.Get("intelligent_tiering.frequent_access_storage_gb").Exists() {
			frequentDataStorage = decimalPtr(decimal.NewFromInt(u.Get("intelligent_tiering.frequent_access_storage_gb").Int()))
		}

		var infrequentDataStorage *decimal.Decimal
		if u != nil && u.Get("intelligent_tiering.infrequent_access_storage_gb").Exists() {
			infrequentDataStorage = decimalPtr(decimal.NewFromInt(u.Get("intelligent_tiering.infrequent_access_storage_gb").Int()))
		}

		var monitAutoObg *decimal.Decimal
		if u != nil && u.Get("intelligent_tiering.monitored_objects").Exists() {
			monitAutoObg = decimalPtr(decimal.NewFromInt(u.Get("intelligent_tiering.monitored_objects").Int()))
		}

		var pcplRequests *decimal.Decimal
		if u != nil && u.Get("intelligent_tiering.monthly_tier_1_requests").Exists() {
			pcplRequests = decimalPtr(decimal.NewFromInt(u.Get("intelligent_tiering.monthly_tier_1_requests").Int()))
		}

		var allOtherRequests *decimal.Decimal
		if u != nil && u.Get("intelligent_tiering.monthly_tier_2_requests").Exists() {
			allOtherRequests = decimalPtr(decimal.NewFromInt(u.Get("intelligent_tiering.monthly_tier_2_requests").Int()))
		}

		var lifecycleRequests *decimal.Decimal
		if u != nil && u.Get("intelligent_tiering.monthly_lifecycle_transition_requests").Exists() {
			lifecycleRequests = decimalPtr(decimal.NewFromInt(u.Get("intelligent_tiering.monthly_lifecycle_transition_requests").Int()))
		}

		var dataScanned *decimal.Decimal
		if u != nil && u.Get("intelligent_tiering.monthly_select_data_scanned_gb").Exists() {
			dataScanned = decimalPtr(decimal.NewFromInt(u.Get("intelligent_tiering.monthly_select_data_scanned_gb").Int()))
		}

		var dataReturned *decimal.Decimal
		if u != nil && u.Get("intelligent_tiering.monthly_select_data_returned_gb").Exists() {
			dataReturned = decimalPtr(decimal.NewFromInt(u.Get("intelligent_tiering.monthly_select_data_returned_gb").Int()))
		}

		var earlyDeletedData *decimal.Decimal
		if u != nil && u.Get("intelligent_tiering.early_delete_gb").Exists() {
			earlyDeletedData = decimalPtr(decimal.NewFromInt(u.Get("intelligent_tiering.early_delete_gb").Int()))
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
		var dataStorage *decimal.Decimal
		if u != nil && u.Get("standard_infrequent_access.storage_gb").Exists() {
			dataStorage = decimalPtr(decimal.NewFromInt(u.Get("standard_infrequent_access.storage_gb").Int()))
		}

		var pcplRequests *decimal.Decimal
		if u != nil && u.Get("standard_infrequent_access.monthly_tier_1_requests").Exists() {
			pcplRequests = decimalPtr(decimal.NewFromInt(u.Get("standard_infrequent_access.monthly_tier_1_requests").Int()))
		}

		var allOtherRequests *decimal.Decimal
		if u != nil && u.Get("standard_infrequent_access.monthly_tier_2_requests").Exists() {
			allOtherRequests = decimalPtr(decimal.NewFromInt(u.Get("standard_infrequent_access.monthly_tier_2_requests").Int()))
		}

		var lifecycleRequests *decimal.Decimal
		if u != nil && u.Get("standard_infrequent_access.monthly_lifecycle_transition_requests").Exists() {
			lifecycleRequests = decimalPtr(decimal.NewFromInt(u.Get("standard_infrequent_access.monthly_lifecycle_transition_requests").Int()))
		}

		var retrievalData *decimal.Decimal
		if u != nil && u.Get("standard_infrequent_access.monthly_retrieval_gb").Exists() {
			retrievalData = decimalPtr(decimal.NewFromInt(u.Get("standard_infrequent_access.monthly_retrieval_gb").Int()))
		}

		var dataScanned *decimal.Decimal
		if u != nil && u.Get("standard_infrequent_access.monthly_select_data_scanned_gb").Exists() {
			dataScanned = decimalPtr(decimal.NewFromInt(u.Get("standard_infrequent_access.monthly_select_data_scanned_gb").Int()))
		}

		var dataReturned *decimal.Decimal
		if u != nil && u.Get("standard_infrequent_access.monthly_select_data_returned_gb").Exists() {
			dataReturned = decimalPtr(decimal.NewFromInt(u.Get("standard_infrequent_access.monthly_select_data_returned_gb").Int()))
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
		var dataStorage *decimal.Decimal
		if u != nil && u.Get("one_zone_infrequent_access.storage_gb").Exists() {
			dataStorage = decimalPtr(decimal.NewFromInt(u.Get("one_zone_infrequent_access.storage_gb").Int()))
		}

		var pcplRequests *decimal.Decimal
		if u != nil && u.Get("one_zone_infrequent_access.monthly_tier_1_requests").Exists() {
			pcplRequests = decimalPtr(decimal.NewFromInt(u.Get("one_zone_infrequent_access.monthly_tier_1_requests").Int()))
		}

		var allOtherRequests *decimal.Decimal
		if u != nil && u.Get("one_zone_infrequent_access.monthly_tier_2_requests").Exists() {
			allOtherRequests = decimalPtr(decimal.NewFromInt(u.Get("one_zone_infrequent_access.monthly_tier_2_requests").Int()))
		}

		var lifecycleRequests *decimal.Decimal
		if u != nil && u.Get("one_zone_infrequent_access.monthly_lifecycle_transition_requests").Exists() {
			lifecycleRequests = decimalPtr(decimal.NewFromInt(u.Get("one_zone_infrequent_access.monthly_lifecycle_transition_requests").Int()))
		}

		var retrievalData *decimal.Decimal
		if u != nil && u.Get("one_zone_infrequent_access.monthly_retrieval_gb").Exists() {
			retrievalData = decimalPtr(decimal.NewFromInt(u.Get("one_zone_infrequent_access.monthly_retrieval_gb").Int()))
		}

		var dataScanned *decimal.Decimal
		if u != nil && u.Get("one_zone_infrequent_access.monthly_select_data_scanned_gb").Exists() {
			dataScanned = decimalPtr(decimal.NewFromInt(u.Get("one_zone_infrequent_access.monthly_select_data_scanned_gb").Int()))
		}

		var dataReturned *decimal.Decimal
		if u != nil && u.Get("one_zone_infrequent_access.monthly_select_data_returned_gb").Exists() {
			dataReturned = decimalPtr(decimal.NewFromInt(u.Get("one_zone_infrequent_access.monthly_select_data_returned_gb").Int()))
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
		var dataStorage *decimal.Decimal
		if u != nil && u.Get("glacier.storage_gb").Exists() {
			dataStorage = decimalPtr(decimal.NewFromInt(u.Get("glacier.storage_gb").Int()))
		}

		var pcplRequests *decimal.Decimal
		if u != nil && u.Get("glacier.monthly_tier_1_requests").Exists() {
			pcplRequests = decimalPtr(decimal.NewFromInt(u.Get("glacier.monthly_tier_1_requests").Int()))
		}

		var allOtherRequests *decimal.Decimal
		if u != nil && u.Get("glacier.monthly_tier_2_requests").Exists() {
			allOtherRequests = decimalPtr(decimal.NewFromInt(u.Get("glacier.monthly_tier_2_requests").Int()))
		}

		var lifecycleRequests *decimal.Decimal
		if u != nil && u.Get("glacier.monthly_lifecycle_transition_requests").Exists() {
			lifecycleRequests = decimalPtr(decimal.NewFromInt(u.Get("glacier.monthly_lifecycle_transition_requests").Int()))
		}

		var earlyDeletedData *decimal.Decimal
		if u != nil && u.Get("glacier.early_delete_gb").Exists() {
			earlyDeletedData = decimalPtr(decimal.NewFromInt(u.Get("glacier.early_delete_gb").Int()))
		}

		var stdDataScanned *decimal.Decimal
		if u != nil && u.Get("glacier.monthly_standard_select_data_scanned_gb").Exists() {
			stdDataScanned = decimalPtr(decimal.NewFromInt(u.Get("glacier.monthly_standard_select_data_scanned_gb").Int()))
		}

		var stdDataReturned *decimal.Decimal
		if u != nil && u.Get("glacier.monthly_standard_select_data_returned_gb").Exists() {
			stdDataReturned = decimalPtr(decimal.NewFromInt(u.Get("glacier.monthly_standard_select_data_returned_gb").Int()))
		}

		var bulkDataScanned *decimal.Decimal
		if u != nil && u.Get("glacier.monthly_bulk_select_data_scanned_gb").Exists() {
			bulkDataScanned = decimalPtr(decimal.NewFromInt(u.Get("glacier.monthly_bulk_select_data_scanned_gb").Int()))
		}

		var bulkDataReturned *decimal.Decimal
		if u != nil && u.Get("glacier.monthly_bulk_select_data_returned_gb").Exists() {
			bulkDataReturned = decimalPtr(decimal.NewFromInt(u.Get("glacier.monthly_bulk_select_data_returned_gb").Int()))
		}

		var expDataScanned *decimal.Decimal
		if u != nil && u.Get("glacier.monthly_expedited_select_data_scanned_gb").Exists() {
			expDataScanned = decimalPtr(decimal.NewFromInt(u.Get("glacier.monthly_expedited_select_data_scanned_gb").Int()))
		}

		var expDataReturned *decimal.Decimal
		if u != nil && u.Get("glacier.monthly_expedited_select_data_returned_gb").Exists() {
			expDataReturned = decimalPtr(decimal.NewFromInt(u.Get("glacier.monthly_expedited_select_data_returned_gb").Int()))
		}

		var stdRetrievalData *decimal.Decimal
		if u != nil && u.Get("glacier.monthly_standard_data_retrieval_gb").Exists() {
			stdRetrievalData = decimalPtr(decimal.NewFromInt(u.Get("glacier.monthly_standard_data_retrieval_gb").Int()))
		}

		var stdRetrievalReq *decimal.Decimal
		if u != nil && u.Get("glacier.monthly_standard_data_retrieval_requests").Exists() {
			stdRetrievalReq = decimalPtr(decimal.NewFromInt(u.Get("glacier.monthly_standard_data_retrieval_requests").Int()))
		}

		var bulkRetrievalData *decimal.Decimal
		if u != nil && u.Get("glacier.monthly_bulk_data_retrieval_gb").Exists() {
			bulkRetrievalData = decimalPtr(decimal.NewFromInt(u.Get("glacier.monthly_bulk_data_retrieval_gb").Int()))
		}

		var bulkRetrievalReq *decimal.Decimal
		if u != nil && u.Get("glacier.monthly_bulk_data_retrieval_requests").Exists() {
			bulkRetrievalReq = decimalPtr(decimal.NewFromInt(u.Get("glacier.monthly_bulk_data_retrieval_requests").Int()))
		}

		var expRetrievalData *decimal.Decimal
		if u != nil && u.Get("glacier.monthly_expedited_data_retrieval_gb").Exists() {
			expRetrievalData = decimalPtr(decimal.NewFromInt(u.Get("glacier.monthly_expedited_data_retrieval_gb").Int()))
		}

		var expRetrievalReq *decimal.Decimal
		if u != nil && u.Get("glacier.monthly_expedited_data_retrieval_requests").Exists() {
			expRetrievalReq = decimalPtr(decimal.NewFromInt(u.Get("glacier.monthly_expedited_data_retrieval_requests").Int()))
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
		var dataStorage *decimal.Decimal
		if u != nil && u.Get("glacier_deep_archive.storage_gb").Exists() {
			dataStorage = decimalPtr(decimal.NewFromInt(u.Get("glacier_deep_archive.storage_gb").Int()))
		}

		var pcplRequests *decimal.Decimal
		if u != nil && u.Get("glacier_deep_archive.monthly_tier_1_requests").Exists() {
			pcplRequests = decimalPtr(decimal.NewFromInt(u.Get("glacier_deep_archive.monthly_tier_1_requests").Int()))
		}

		var allOtherRequests *decimal.Decimal
		if u != nil && u.Get("glacier_deep_archive.monthly_tier_2_requests").Exists() {
			allOtherRequests = decimalPtr(decimal.NewFromInt(u.Get("glacier_deep_archive.monthly_tier_2_requests").Int()))
		}

		var lifecycleRequests *decimal.Decimal
		if u != nil && u.Get("glacier_deep_archive.monthly_lifecycle_transition_requests").Exists() {
			lifecycleRequests = decimalPtr(decimal.NewFromInt(u.Get("glacier_deep_archive.monthly_lifecycle_transition_requests").Int()))
		}

		var stdRetrievalData *decimal.Decimal
		if u != nil && u.Get("glacier_deep_archive.monthly_standard_data_retrieval_gb").Exists() {
			stdRetrievalData = decimalPtr(decimal.NewFromInt(u.Get("glacier_deep_archive.monthly_standard_data_retrieval_gb").Int()))
		}

		var stdRetrievalReq *decimal.Decimal
		if u != nil && u.Get("glacier_deep_archive.monthly_standard_data_retrieval_requests").Exists() {
			stdRetrievalReq = decimalPtr(decimal.NewFromInt(u.Get("glacier_deep_archive.monthly_standard_data_retrieval_requests").Int()))
		}

		var bulkRetrievalData *decimal.Decimal
		if u != nil && u.Get("glacier_deep_archive.monthly_bulk_data_retrieval_gb").Exists() {
			bulkRetrievalData = decimalPtr(decimal.NewFromInt(u.Get("glacier_deep_archive.monthly_bulk_data_retrieval_gb").Int()))
		}

		var bulkRetrievalReq *decimal.Decimal
		if u != nil && u.Get("glacier_deep_archive.monthly_bulk_data_retrieval_requests").Exists() {
			bulkRetrievalReq = decimalPtr(decimal.NewFromInt(u.Get("glacier_deep_archive.monthly_bulk_data_retrieval_requests").Int()))
		}

		var earlyDeletedData *decimal.Decimal
		if u != nil && u.Get("glacier_deep_archive.early_delete_gb").Exists() {
			earlyDeletedData = decimalPtr(decimal.NewFromInt(u.Get("glacier_deep_archive.early_delete_gb").Int()))
		}

		return &schema.Resource{
			Name: "Glacier deep archive",
			CostComponents: []*schema.CostComponent{
				s3StorageCostComponent("Storage", "AmazonS3GlacierDeepArchive", region, "TimedStorage-GDA-ByteHrs", dataStorage),
				s3ApiOperationCostComponent("PUT, COPY, POST, LIST requests", "AmazonS3GlacierDeepArchive", region, "Requests-GDA-Tier1", "PostObject", pcplRequests),
				s3ApiCostComponent("GET, SELECT, and all other requests", "AmazonS3", region, "Requests-GDA-Tier2", allOtherRequests),
				s3LifecycleTransitionsCostComponent(region, "Requests-Tier3", "S3-GDATransition", lifecycleRequests),
				s3ApiOperationCostComponent("Retrieval requests (standard)", "AmazonS3GlacierDeepArchive", region, "Requests-GDA-Tier3", "DeepArchiveRestoreObject", stdRetrievalReq),
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
		Unit:            "GB-months",
		UnitMultiplier:  1,
		MonthlyQuantity: dataStorage,
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

func s3StorageVolumeTypeCostComponent(name string, service string, region string, usageType string, volumeType string, dataStorage *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "GB-months",
		UnitMultiplier:  1,
		MonthlyQuantity: dataStorage,
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

func s3ApiCostComponent(name string, service string, region string, usageType string, requests *decimal.Decimal) *schema.CostComponent {
	return s3ApiOperationCostComponent(name, service, region, usageType, "", requests)
}

func s3ApiOperationCostComponent(name string, service string, region string, usageType string, operation string, requests *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "requests",
		UnitMultiplier:  1000,
		MonthlyQuantity: requests,
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

func s3DataCostComponent(name string, service string, region string, usageType string, data *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "GB-months",
		UnitMultiplier:  1,
		MonthlyQuantity: data,
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

func s3DataGroupCostComponent(name string, service string, region string, usageType string, group string, data *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "GB-months",
		UnitMultiplier:  1,
		MonthlyQuantity: data,
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

func s3LifecycleTransitionsCostComponent(region string, usageType string, operation string, requests *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Lifecycle transition",
		Unit:            "requests",
		UnitMultiplier:  1000,
		MonthlyQuantity: requests,
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

func s3MonitoringCostComponent(region string, objects *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Monitoring and automation",
		Unit:            "objects",
		UnitMultiplier:  1000,
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
