package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

type S3GlacierStorageClass struct {
	// "required" args that can't really be missing.
	Region string

	// "usage" args
	StorageGB                             *int64 `infracost_usage:"storage_gb"`
	MonthlyTier1Requests                  *int64 `infracost_usage:"monthly_tier_1_requests"`
	MonthlyTier2Requests                  *int64 `infracost_usage:"monthly_tier_2_requests"`
	MonthlyLifecycleTransitionRequests    *int64 `infracost_usage:"monthly_lifecycle_transition_requests"`
	MonthlyStandardDataRetrievalRequests  *int64 `infracost_usage:"monthly_standard_data_retrieval_requests"`
	MonthlyStandardDataRetrievalGB        *int64 `infracost_usage:"monthly_standard_data_retrieval_gb"`
	MonthlyStandardSelectDataScannedGB    *int64 `infracost_usage:"monthly_standard_select_data_scanned_gb"`
	MonthlyStandardSelectDataReturnedGB   *int64 `infracost_usage:"monthly_standard_select_data_returned_gb"`
	MonthlyExpeditedDataRetrievalRequests *int64 `infracost_usage:"monthly_expedited_data_retrieval_requests"`
	MonthlyExpeditedDataRetrievalGB       *int64 `infracost_usage:"monthly_expedited_data_retrieval_gb"`
	MonthlyExpeditedSelectDataScannedGB   *int64 `infracost_usage:"monthly_expedited_select_data_scanned_gb"`
	MonthlyExpeditedSelectDataReturnedGB  *int64 `infracost_usage:"monthly_expedited_select_data_returned_gb"`
	MonthlyBulkDataRetrievalRequests      *int64 `infracost_usage:"monthly_bulk_data_retrieval_requests"`
	MonthlyBulkDataRetrievalGB            *int64 `infracost_usage:"monthly_bulk_data_retrieval_gb"`
	MonthlyBulkSelectDataScannedGB        *int64 `infracost_usage:"monthly_bulk_select_data_scanned_gb"`
	MonthlyBulkSelectDataReturnedGB       *int64 `infracost_usage:"monthly_bulk_select_data_returned_gb"`
	EarlyDeleteGB                         *int64 `infracost_usage:"early_delete_gb"`
}

var S3GlacierStorageClassUsageSchema = []*schema.UsageSchemaItem{
	{Key: "storage_gb", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "monthly_tier_1_requests", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "monthly_tier_2_requests", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "monthly_lifecycle_transition_requests", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "monthly_standard_data_retrieval_requests", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "monthly_standard_data_retrieval_gb", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "monthly_standard_select_data_scanned_gb", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "monthly_standard_select_data_returned_gb", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "monthly_expedited_data_retrieval_requests", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "monthly_expedited_data_retrieval_gb", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "monthly_expedited_select_data_scanned_gb", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "monthly_expedited_select_data_returned_gb", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "monthly_bulk_data_retrieval_requests", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "monthly_bulk_data_retrieval_gb", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "monthly_bulk_select_data_scanned_gb", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "monthly_bulk_select_data_returned_gb", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "early_delete_gb", DefaultValue: 0, ValueType: schema.Int64},
}

func (a *S3GlacierStorageClass) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(a, u)
}

func (a *S3GlacierStorageClass) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name:        "Glacier",
		UsageSchema: S3BucketUsageSchema,
		CostComponents: []*schema.CostComponent{
			s3StorageCostComponent("Storage", "AmazonGlacier", a.Region, "TimedStorage-ByteHrs", a.StorageGB),
			s3ApiOperationCostComponent("PUT, COPY, POST, LIST requests", "AmazonS3", a.Region, "Requests-GLACIER-Tier1", "PostObject", a.MonthlyTier1Requests),
			s3ApiCostComponent("GET, SELECT, and all other requests", "AmazonS3", a.Region, "Requests-GLACIER-Tier2", a.MonthlyTier2Requests),
			s3LifecycleTransitionsCostComponent(a.Region, "Requests-Tier3", "S3-GlacierTransition", a.MonthlyLifecycleTransitionRequests),
			s3ApiCostComponent("Retrieval requests (standard)", "AmazonGlacier", a.Region, "Requests-Tier1", a.MonthlyStandardDataRetrievalRequests),
			s3DataCostComponent("Retrievals (standard)", "AmazonGlacier", a.Region, "Standard-Retrieval-Bytes", a.MonthlyStandardDataRetrievalGB),
			s3DataCostComponent("Select data scanned (standard)", "AmazonGlacier", a.Region, "Std-Select-Scanned-Bytes", a.MonthlyStandardSelectDataScannedGB),
			s3DataCostComponent("Select data returned (standard)", "AmazonGlacier", a.Region, "Std-Select-Returned-Bytes", a.MonthlyStandardSelectDataReturnedGB),
			s3ApiCostComponent("Retrieval requests (expedited)", "AmazonGlacier", a.Region, "Requests-Tier3", a.MonthlyExpeditedDataRetrievalRequests),
			s3DataCostComponent("Retrievals (expedited)", "AmazonGlacier", a.Region, "Expedited-Retrieval-Bytes", a.MonthlyExpeditedDataRetrievalGB),
			s3DataCostComponent("Select data scanned (expedited)", "AmazonGlacier", a.Region, "Exp-Select-Scanned-Bytes", a.MonthlyExpeditedSelectDataScannedGB),
			s3DataCostComponent("Select data returned (expedited)", "AmazonGlacier", a.Region, "Exp-Select-Returned-Bytes", a.MonthlyExpeditedSelectDataReturnedGB),
			s3ApiCostComponent("Retrieval requests (bulk)", "AmazonGlacier", a.Region, "Requests-Tier2", a.MonthlyBulkDataRetrievalRequests),
			s3DataCostComponent("Retrievals (bulk)", "AmazonGlacier", a.Region, "Bulk-Retrieval-Bytes", a.MonthlyBulkDataRetrievalGB),
			s3DataCostComponent("Select data scanned (bulk)", "AmazonGlacier", a.Region, "Bulk-Select-Scanned-Bytes", a.MonthlyBulkSelectDataScannedGB),
			s3DataCostComponent("Select data returned (bulk)", "AmazonGlacier", a.Region, "Bulk-Select-Returned-Bytes", a.MonthlyBulkSelectDataReturnedGB),
			s3DataCostComponent("Early delete (within 90 days)", "AmazonGlacier", a.Region, "EarlyDelete-ByteHrs", a.EarlyDeleteGB),
		},
	}
}
