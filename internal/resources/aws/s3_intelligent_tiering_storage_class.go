package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

type S3IntelligentTieringStorageClass struct {
	// "required" args that can't really be missing.
	Region string

	// "usage" args
	FrequentAccessStorageGB            *float64 `infracost_usage:"frequent_access_storage_gb"`
	InfrequentAccessStorageGB          *float64 `infracost_usage:"infrequent_access_storage_gb"`
	ArchiveAccessStorageGB             *float64 `infracost_usage:"archive_access_storage_gb"`
	DeepArchiveAccessStorageGB         *float64 `infracost_usage:"deep_archive_access_storage_gb"`
	MonitoredObjects                   *int64   `infracost_usage:"monitored_objects"`
	MonthlyTier1Requests               *int64   `infracost_usage:"monthly_tier_1_requests"`
	MonthlyTier2Requests               *int64   `infracost_usage:"monthly_tier_2_requests"`
	MonthlyLifecycleTransitionRequests *int64   `infracost_usage:"monthly_lifecycle_transition_requests"`
	MonthlySelectDataScannedGB         *float64 `infracost_usage:"monthly_select_data_scanned_gb"`
	MonthlySelectDataReturnedGB        *float64 `infracost_usage:"monthly_select_data_returned_gb"`
	EarlyDeleteGB                      *float64 `infracost_usage:"early_delete_gb"`
}

var S3IntelligentTieringStorageClassUsageSchema = []*schema.UsageItem{
	{Key: "frequent_access_storage_gb", DefaultValue: 0.0, ValueType: schema.Float64},
	{Key: "infrequent_access_storage_gb", DefaultValue: 0.0, ValueType: schema.Float64},
	{Key: "archive_access_storage_gb", DefaultValue: 0.0, ValueType: schema.Float64},
	{Key: "deep_archive_access_storage_gb", DefaultValue: 0.0, ValueType: schema.Float64},
	{Key: "monitored_objects", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "monthly_tier_1_requests", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "monthly_tier_2_requests", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "monthly_lifecycle_transition_requests", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "monthly_select_data_scanned_gb", DefaultValue: 0.0, ValueType: schema.Float64},
	{Key: "monthly_select_data_returned_gb", DefaultValue: 0.0, ValueType: schema.Float64},
	{Key: "early_delete_gb", DefaultValue: 0.0, ValueType: schema.Float64},
}

func (a *S3IntelligentTieringStorageClass) CoreType() string {
	return "S3IntelligentTieringStorageClass"
}

func (a *S3IntelligentTieringStorageClass) UsageSchema() []*schema.UsageItem {
	return S3IntelligentTieringStorageClassUsageSchema
}

func (a *S3IntelligentTieringStorageClass) UsageKey() string {
	return "intelligent_tiering"
}

func (a *S3IntelligentTieringStorageClass) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(a, u)
}

func (a *S3IntelligentTieringStorageClass) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name:        "Intelligent tiering",
		UsageSchema: a.UsageSchema(),
		CostComponents: []*schema.CostComponent{
			s3StorageCostComponent("Storage (frequent access)", "AmazonS3", a.Region, "TimedStorage-INT-FA-ByteHrs", a.FrequentAccessStorageGB),
			s3StorageCostComponent("Storage (infrequent access)", "AmazonS3", a.Region, "TimedStorage-INT-IA-ByteHrs", a.InfrequentAccessStorageGB),
			s3StorageVolumeTypeCostComponent("Storage (archive access)", "AmazonS3", a.Region, "TimedStorage-INT-AA-ByteHrs", "IntelligentTieringArchive", a.FrequentAccessStorageGB),
			s3StorageVolumeTypeCostComponent("Storage (deep archive access)", "AmazonS3", a.Region, "TimedStorage-INT-DAA-ByteHrs", "IntelligentTieringDeepArchive", a.InfrequentAccessStorageGB),
			s3MonitoringCostComponent(a.Region, a.MonitoredObjects),
			s3ApiCostComponent("PUT, COPY, POST, LIST requests", "AmazonS3", a.Region, "Requests-INT-Tier1", a.MonthlyTier1Requests),
			s3ApiCostComponent("GET, SELECT, and all other requests", "AmazonS3", a.Region, "Requests-INT-Tier2", a.MonthlyTier2Requests),
			s3LifecycleTransitionsCostComponent(a.Region, "Requests-Tier4", "", a.MonthlyLifecycleTransitionRequests),
			s3DataCostComponent("Select data scanned", "AmazonS3", a.Region, "Select-Scanned-INT-Bytes", a.MonthlySelectDataScannedGB),
			s3DataCostComponent("Select data returned", "AmazonS3", a.Region, "Select-Returned-INT-Bytes", a.MonthlySelectDataReturnedGB),
			s3DataCostComponent("Early delete (within 30 days)", "AmazonS3", a.Region, "EarlyDelete-INT", a.EarlyDeleteGB),
		},
	}
}
