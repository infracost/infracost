package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

type S3OneZoneInfrequentAccessStorageClass struct {
	// "required" args that can't really be missing.
	Region string

	// "usage" args
	StorageGB                          *float64 `infracost_usage:"storage_gb"`
	MonthlyTier1Requests               *int64   `infracost_usage:"monthly_tier_1_requests"`
	MonthlyTier2Requests               *int64   `infracost_usage:"monthly_tier_2_requests"`
	MonthlyLifecycleTransitionRequests *int64   `infracost_usage:"monthly_lifecycle_transition_requests"`
	MonthlyDataRetrievalGB             *float64 `infracost_usage:"monthly_data_retrieval_gb"`
	MonthlySelectDataScannedGB         *float64 `infracost_usage:"monthly_select_data_scanned_gb"`
	MonthlySelectDataReturnedGB        *float64 `infracost_usage:"monthly_select_data_returned_gb"`
}

var S3OneZoneInfrequentAccessStorageClassUsageSchema = []*schema.UsageItem{
	{Key: "storage_gb", DefaultValue: 0.0, ValueType: schema.Float64},
	{Key: "monthly_tier_1_requests", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "monthly_tier_2_requests", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "monthly_lifecycle_transition_requests", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "monthly_data_retrieval_gb", DefaultValue: 0.0, ValueType: schema.Float64},
	{Key: "monthly_select_data_scanned_gb", DefaultValue: 0.0, ValueType: schema.Float64},
	{Key: "monthly_select_data_returned_gb", DefaultValue: 0.0, ValueType: schema.Float64},
}

func (a *S3OneZoneInfrequentAccessStorageClass) CoreType() string {
	return "S3OneZoneInfrequentAccessStorageClass"
}

func (a *S3OneZoneInfrequentAccessStorageClass) UsageSchema() []*schema.UsageItem {
	return S3OneZoneInfrequentAccessStorageClassUsageSchema
}

func (a *S3OneZoneInfrequentAccessStorageClass) UsageKey() string {
	return "one_zone_infrequent_access"
}

func (a *S3OneZoneInfrequentAccessStorageClass) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(a, u)
}

func (a *S3OneZoneInfrequentAccessStorageClass) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name:        "One zone - infrequent access",
		UsageSchema: a.UsageSchema(),
		CostComponents: []*schema.CostComponent{
			s3StorageCostComponent("Storage", "AmazonS3", a.Region, "TimedStorage-ZIA-ByteHrs", a.StorageGB),
			s3ApiCostComponent("PUT, COPY, POST, LIST requests", "AmazonS3", a.Region, "Requests-ZIA-Tier1", a.MonthlyTier1Requests),
			s3ApiCostComponent("GET, SELECT, and all other requests", "AmazonS3", a.Region, "Requests-ZIA-Tier2", a.MonthlyTier2Requests),
			s3LifecycleTransitionsCostComponent(a.Region, "Requests-Tier4", "", a.MonthlyLifecycleTransitionRequests),
			s3DataCostComponent("Retrievals", "AmazonS3", a.Region, "Retrieval-ZIA", a.MonthlyDataRetrievalGB),
			s3DataCostComponent("Select data scanned", "AmazonS3", a.Region, "Select-Scanned-ZIA-Bytes", a.MonthlySelectDataScannedGB),
			s3DataCostComponent("Select data returned", "AmazonS3", a.Region, "Select-Returned-ZIA-Bytes", a.MonthlySelectDataReturnedGB),
		},
	}
}
