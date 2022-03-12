package aws

import (
	"context"
	"fmt"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
	"github.com/infracost/infracost/internal/usage/aws"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

type S3BucketLifecycleConfiguration struct {
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
	// S3StorageClass is defined in s3_bucket.go
	storageClasses    []S3StorageClass
	allStorageClasses []S3StorageClass
}

var S3BucketLifecycleConfigurationUsageSchema = []*schema.UsageItem{
	{Key: "object_tags", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "standard", DefaultValue: &usage.ResourceUsage{Name: "standard", Items: S3StandardStorageClassUsageSchema}, ValueType: schema.SubResourceUsage},
	{Key: "intelligent_tiering", DefaultValue: &usage.ResourceUsage{Name: "intelligent_tiering", Items: S3IntelligentTieringStorageClassUsageSchema}, ValueType: schema.SubResourceUsage},
	{Key: "standard_infrequent_access", DefaultValue: &usage.ResourceUsage{Name: "standard_infrequent_access", Items: S3StandardInfrequentAccessStorageClassUsageSchema}, ValueType: schema.SubResourceUsage},
	{Key: "one_zone_infrequent_access", DefaultValue: &usage.ResourceUsage{Name: "one_zone_infrequent_access", Items: S3OneZoneInfrequentAccessStorageClassUsageSchema}, ValueType: schema.SubResourceUsage},
	{Key: "glacier_flexible_retrieval", DefaultValue: &usage.ResourceUsage{Name: "glacier_flexible_retrieval", Items: S3GlacierFlexibleRetrievalStorageClassUsageSchema}, ValueType: schema.SubResourceUsage},
	{Key: "glacier_deep_archive", DefaultValue: &usage.ResourceUsage{Name: "glacier_deep_archive", Items: S3GlacierDeepArchiveStorageClassUsageSchema}, ValueType: schema.SubResourceUsage},
}

func (r *S3BucketLifecycleConfiguration) AllStorageClasses() []S3StorageClass {
	if r.allStorageClasses == nil {
		r.allStorageClasses = []S3StorageClass{
			&S3StandardStorageClass{Region: r.Region},
			&S3IntelligentTieringStorageClass{Region: r.Region},
			&S3StandardInfrequentAccessStorageClass{Region: r.Region},
			&S3OneZoneInfrequentAccessStorageClass{Region: r.Region},
			&S3GlacierFlexibleRetrievalStorageClass{Region: r.Region},
			&S3GlacierDeepArchiveStorageClass{Region: r.Region},
		}
	}

	return r.allStorageClasses
}

func (r *S3BucketLifecycleConfiguration) PopulateUsage(u *schema.UsageData) {
	// Add the storage classes based on what's based through in the usage
	// and any storage classes added in the lifecycle storage classes.
	for _, storageClass := range r.AllStorageClasses() {
		if stringInSlice(r.LifecycleStorageClasses, storageClass.UsageKey()) || (u != nil && !u.IsEmpty(storageClass.UsageKey())) {
			// Populate the storage class usage using the map in the usage data
			if u != nil {
				storageClass.PopulateUsage(&schema.UsageData{
					Address:    storageClass.UsageKey(),
					Attributes: u.Get(storageClass.UsageKey()).Map(),
				})
			}
			r.storageClasses = append(r.storageClasses, storageClass)
		}
	}

	resources.PopulateArgsWithUsage(r, u)
}

func (r *S3BucketLifecycleConfiguration) BuildResource() *schema.Resource {
	costComponents := make([]*schema.CostComponent, 0)
	if r.ObjectTagsEnabled {
		costComponents = append(costComponents, r.objectTagsCostComponent())
	}

	subResources := make([]*schema.Resource, 0, len(r.storageClasses))
	for _, storageClass := range r.storageClasses {
		subResources = append(subResources, storageClass.BuildResource())
	}

	estimate := func(ctx context.Context, u map[string]interface{}) error {
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
		for _, storageClass := range r.AllStorageClasses() {
			if _, ok := storageMetricsMap[storageClass.UsageKey()]; !ok {
				continue
			}

			storageClassUsage := make(map[string]interface{})
			if v, ok := u[storageClass.UsageKey()]; ok && v != nil {
				storageClassUsage = v.(map[string]interface{})
			}

			for usageKey, metric := range storageMetricsMap[storageClass.UsageKey()] {
				storageBytes, err := aws.S3GetBucketSizeBytes(ctx, r.Region, r.Name, metric)
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

		filter, err := aws.S3FindMetricsFilter(ctx, r.Region, r.Name)
		if err != nil || filter == "" {
			msg := "Unable to find matching metrics filter for S3 bucket, so unable to sync additional metrics"
			if err != nil {
				msg = fmt.Sprintf("%s: %s", msg, err)
			}
			log.Debugf(msg)
		} else {
			standardStorageClassUsage := u["standard"].(map[string]interface{})

			monthlyTier1Requests, err := aws.S3GetBucketRequests(ctx, r.Region, r.Name, filter, []string{"PutRequests", "PostRequests", "ListRequests"})
			if err != nil {
				return err
			}

			monthlyTier2Requests, err := aws.S3GetBucketRequests(ctx, r.Region, r.Name, filter, []string{"GetRequests", "HeadRequests", "SelectRequests"})
			if err != nil {
				return err
			}

			selectDataScannedBytes, err := aws.S3GetBucketDataBytes(ctx, r.Region, r.Name, filter, "SelectBytesScanned")
			if err != nil {
				return err
			}

			selectDataReturnedBytes, err := aws.S3GetBucketDataBytes(ctx, r.Region, r.Name, filter, "SelectBytesReturned")
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
		Name:           r.Address,
		UsageSchema:    S3BucketLifecycleConfigurationUsageSchema,
		EstimateUsage:  estimate,
		CostComponents: costComponents,
		SubResources:   subResources,
	}
}

func (r *S3BucketLifecycleConfiguration) objectTagsCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Object tagging",
		Unit:            "10k tags",
		UnitMultiplier:  decimal.NewFromInt(10000),
		MonthlyQuantity: intPtrToDecimalPtr(r.ObjectTags),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(r.Region),
			Service:    strPtr("AmazonS3"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("TagStorage-TagHrs")},
			},
		},
	}
}
