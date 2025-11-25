package aws

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
	"github.com/infracost/infracost/internal/usage/aws"
)

type S3Bucket struct {
	// "required" args that can't really be missing.
	Address           string
	Region            string
	Name              string
	ObjectTagsEnabled bool

	// "optional" args, that may be empty depending on the resource config
	LifecycleStorageClasses []string

	// "usage" args
	ObjectTags *int64 `infracost_usage:"object_tags"`

	// "derived" attributes, that are constructed from the other arguments
	storageClasses    []S3StorageClass
	allStorageClasses []S3StorageClass
}

func (a *S3Bucket) CoreType() string {
	return "S3Bucket"
}

func (a *S3Bucket) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "object_tags", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "standard", DefaultValue: &usage.ResourceUsage{Name: "standard", Items: S3StandardStorageClassUsageSchema}, ValueType: schema.SubResourceUsage},
		{Key: "intelligent_tiering", DefaultValue: &usage.ResourceUsage{Name: "intelligent_tiering", Items: S3IntelligentTieringStorageClassUsageSchema}, ValueType: schema.SubResourceUsage},
		{Key: "standard_infrequent_access", DefaultValue: &usage.ResourceUsage{Name: "standard_infrequent_access", Items: S3StandardInfrequentAccessStorageClassUsageSchema}, ValueType: schema.SubResourceUsage},
		{Key: "one_zone_infrequent_access", DefaultValue: &usage.ResourceUsage{Name: "one_zone_infrequent_access", Items: S3OneZoneInfrequentAccessStorageClassUsageSchema}, ValueType: schema.SubResourceUsage},
		{Key: "glacier_flexible_retrieval", DefaultValue: &usage.ResourceUsage{Name: "glacier_flexible_retrieval", Items: S3GlacierFlexibleRetrievalStorageClassUsageSchema}, ValueType: schema.SubResourceUsage},
		{Key: "glacier_deep_archive", DefaultValue: &usage.ResourceUsage{Name: "glacier_deep_archive", Items: S3GlacierDeepArchiveStorageClassUsageSchema}, ValueType: schema.SubResourceUsage},
	}
}

type S3StorageClass interface {
	UsageKey() string
	PopulateUsage(u *schema.UsageData)
	BuildResource() *schema.Resource
}

func (a *S3Bucket) AllStorageClasses() []S3StorageClass {
	if a.allStorageClasses == nil {
		a.allStorageClasses = []S3StorageClass{
			&S3StandardStorageClass{Region: a.Region},
			&S3IntelligentTieringStorageClass{Region: a.Region},
			&S3StandardInfrequentAccessStorageClass{Region: a.Region},
			&S3OneZoneInfrequentAccessStorageClass{Region: a.Region},
			&S3GlacierFlexibleRetrievalStorageClass{Region: a.Region},
			&S3GlacierDeepArchiveStorageClass{Region: a.Region},
		}
	}

	return a.allStorageClasses
}

func (a *S3Bucket) PopulateUsage(u *schema.UsageData) {
	// Add the storage classes based on what's based through in the usage
	// and any storage classes added in the lifecycle storage classes.
	for _, storageClass := range a.AllStorageClasses() {
		if stringInSlice(a.LifecycleStorageClasses, storageClass.UsageKey()) || (u != nil && !u.IsEmpty(storageClass.UsageKey())) {
			// Populate the storage class usage using the map in the usage data
			if u != nil {
				storageClass.PopulateUsage(&schema.UsageData{
					Address:    storageClass.UsageKey(),
					Attributes: u.Get(storageClass.UsageKey()).Map(),
				})
			}
			a.storageClasses = append(a.storageClasses, storageClass)
		}
	}

	resources.PopulateArgsWithUsage(a, u)
}

func (a *S3Bucket) BuildResource() *schema.Resource {
	costComponents := make([]*schema.CostComponent, 0)
	if a.ObjectTagsEnabled {
		costComponents = append(costComponents, a.objectTagsCostComponent())
	}

	subResources := make([]*schema.Resource, 0, len(a.storageClasses))
	for _, storageClass := range a.storageClasses {
		subResources = append(subResources, storageClass.BuildResource())
	}

	estimate := func(ctx context.Context, u map[string]any) error {
		// https://docs.aws.amazon.com/AmazonS3/latest/userguide/metrics-dimensions.html

		storageMetricsMap := map[string]map[string]string{
			"standard": {
				"storage_gb": "StandardStorage",
			},
			"intelligent_tiering": {
				"frequent_access_storage_gb":     "IntelligentTieringFAStorage",
				"infrequent_access_storage_gb":   "IntelligentTieringIAStorage",
				"archive_access_storage_gb":      "IntelligentTieringAAStorage",
				"deep_archive_access_storage_gb": "IntelligentTieringDAAStorage",
			},
			"standard_infrequent_access": {
				"storage_gb": "StandardIAStorage",
			},
			"one_zone_infrequent_access": {
				"storage_gb": "OneZoneIAStorage",
			},
			"glacier_flexible_retrieval": {
				"storage_gb": "GlacierStorage",
			},
			"glacier_deep_archive": {
				"storage_gb": "DeepArchiveStorage",
			},
		}

		// We want to check all storage classes, not just the ones that have been added by the lifecycle policy or previous
		// usage data, so that any additional storage classes that have estimated data will be added when we reload the resources.
		for _, storageClass := range a.AllStorageClasses() {
			if _, ok := storageMetricsMap[storageClass.UsageKey()]; !ok {
				continue
			}

			storageClassUsage := make(map[string]any)
			if v, ok := u[storageClass.UsageKey()]; ok && v != nil {
				storageClassUsage = v.(map[string]any)
			}

			for usageKey, metric := range storageMetricsMap[storageClass.UsageKey()] {
				storageBytes, err := aws.S3GetBucketSizeBytes(ctx, a.Region, a.Name, metric)
				if err != nil {
					return err
				}

				// Always add usage for the Standard storage class, but skip others that have no data.
				if storageBytes > 0 || storageClass.UsageKey() == "standard" {
					storageClassUsage[usageKey] = storageBytes / 1000 / 1000 / 1000
				}
			}

			if len(storageClassUsage) > 0 {
				u[storageClass.UsageKey()] = storageClassUsage
			}
		}

		filter, err := aws.S3FindMetricsFilter(ctx, a.Region, a.Name)
		if err != nil || filter == "" {
			msg := "Unable to find matching metrics filter for S3 bucket, so unable to sync additional metrics"
			if err != nil {
				msg = fmt.Sprintf("%s: %s", msg, err)
			}
			logging.Logger.Debug().Msg(msg)
		} else {
			standardStorageClassUsage := u["standard"].(map[string]any)

			monthlyTier1Requests, err := aws.S3GetBucketRequests(ctx, a.Region, a.Name, filter, []string{"PutRequests", "PostRequests", "ListRequests"})
			if err != nil {
				return err
			}

			monthlyTier2Requests, err := aws.S3GetBucketRequests(ctx, a.Region, a.Name, filter, []string{"GetRequests", "HeadRequests", "SelectRequests"})
			if err != nil {
				return err
			}

			selectDataScannedBytes, err := aws.S3GetBucketDataBytes(ctx, a.Region, a.Name, filter, "SelectBytesScanned")
			if err != nil {
				return err
			}

			selectDataReturnedBytes, err := aws.S3GetBucketDataBytes(ctx, a.Region, a.Name, filter, "SelectBytesReturned")
			if err != nil {
				return err
			}

			standardStorageClassUsage["monthly_tier_1_requests"] = monthlyTier1Requests
			standardStorageClassUsage["monthly_tier_2_requests"] = monthlyTier2Requests
			standardStorageClassUsage["monthly_select_data_scanned_gb"] = selectDataScannedBytes / 1000 / 1000 / 1000
			standardStorageClassUsage["monthly_select_data_returned_gb"] = selectDataReturnedBytes / 1000 / 1000 / 1000
		}

		return nil
	}

	return &schema.Resource{
		Name:           a.Address,
		UsageSchema:    a.UsageSchema(),
		EstimateUsage:  estimate,
		CostComponents: costComponents,
		SubResources:   subResources,
	}
}

func (a *S3Bucket) objectTagsCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Object tagging",
		Unit:            "10k tags",
		UnitMultiplier:  decimal.NewFromInt(10000),
		MonthlyQuantity: intPtrToDecimalPtr(a.ObjectTags),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(a.Region),
			Service:    strPtr("AmazonS3"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/TagStorage-TagHrs/")},
			},
		},
		UsageBased: true,
	}
}
