package aws

import (
	"sort"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
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
	storageClasses []S3StorageClass
}

type S3StorageClass interface {
	PopulateUsage(u *schema.UsageData)
	BuildResource() *schema.Resource
}

var S3BucketUsageSchema = []*schema.UsageSchemaItem{
	{Key: "object_tags", DefaultValue: 0, ValueType: schema.Int64},
}

func (a *S3Bucket) PopulateUsage(u *schema.UsageData) {
	// Add the storage classes based on what's based through in the usage
	// and any storage classes added in the lifecycle storage classes.
	storageClassMap := map[string]S3StorageClass{
		"standard":                   &S3StandardStorageClass{Region: a.Region},
		"intelligent_tiering":        &S3IntelligentTieringStorageClass{Region: a.Region},
		"standard_infrequent_access": &S3StandardInfrequentAccessStorageClass{Region: a.Region},
		"one_zone_infrequent_access": &S3OneZoneInfrequentAccessStorageClass{Region: a.Region},
		"glacier":                    &S3GlacierStorageClass{Region: a.Region},
		"glacier_deep_archive":       &S3GlacierDeepArchiveStorageClass{Region: a.Region},
	}

	for k, storageClass := range storageClassMap {
		if stringInSlice(a.LifecycleStorageClasses, k) || (u != nil && u.Get(k).Type != gjson.Null) {
			// Populate the storage class usage using the map in the usage data
			if u != nil {
				storageClass.PopulateUsage(&schema.UsageData{
					Address:    k,
					Attributes: u.Get(k).Map(),
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

	// Sort the subresources so that the output is deterministic
	sort.Slice(subResources, func(i, j int) bool {
		return subResources[i].Name < subResources[j].Name
	})

	return &schema.Resource{
		Name:           a.Address,
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
	}
}
