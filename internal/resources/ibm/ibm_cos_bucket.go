package ibm

import (
	"strings"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// IbmCosBucket struct represents IBM Cloud Object Storage instance
//
// Resource information: https://cloud.ibm.com/objectstorage
// Pricing information: https://cloud.ibm.com/objectstorage/create#pricing

type IbmCosBucket struct {
	Address            string
	Location           string
	LocationIdentifier string
	StorageClass       string
	Archive            bool
	ArchiveType        string
	Plan               string

	MonthlyAverageCapacity     *float64 `infracost_usage:"monthly_average_capacity"`
	PublicStandardEgress       *float64 `infracost_usage:"public_standard_egress"`
	AsperaIngress              *float64 `infracost_usage:"aspera_ingress"`
	AsperaEgress               *float64 `infracost_usage:"aspera_egress"`
	ArchiveCapacity            *float64 `infracost_usage:"archive_capacity"`
	ArchiveRestore             *float64 `infracost_usage:"archive_restore"`
	AcceleratedArchiveCapacity *float64 `infracost_usage:"accelerated_archive_capacity"`
	AcceleratedArchiveRestore  *float64 `infracost_usage:"accelerated_archive_restore"`
	MonthlyDataRetrieval       *float64 `infracost_usage:"monthly_data_retrieval"`
	ClassARequestCount         *int64   `infracost_usage:"class_a_request_count"`
	ClassBRequestCount         *int64   `infracost_usage:"class_b_request_count"`
}

// IbmCosBucketUsageSchema defines a list which represents the usage schema of IbmCosBucket.
var IbmCosBucketUsageSchema = []*schema.UsageItem{
	{Key: "monthly_average_capacity", ValueType: schema.Float64, DefaultValue: 0},
	{Key: "public_standard_egress", ValueType: schema.Float64, DefaultValue: 0},
	{Key: "aspera_ingress", ValueType: schema.Float64, DefaultValue: 0},
	{Key: "aspera_egress", ValueType: schema.Float64, DefaultValue: 0},
	{Key: "archive_capacity", ValueType: schema.Float64, DefaultValue: 0},
	{Key: "archive_restore", ValueType: schema.Float64, DefaultValue: 0},
	{Key: "accelerated_archive_capacity", ValueType: schema.Float64, DefaultValue: 0},
	{Key: "accelerated_archive_restore", ValueType: schema.Float64, DefaultValue: 0},
	{Key: "class_a_request_count", ValueType: schema.Int64, DefaultValue: 0},
	{Key: "class_b_request_count", ValueType: schema.Int64, DefaultValue: 0},
	{Key: "monthly_data_retrieval", ValueType: schema.Int64, DefaultValue: 0},
}

var oneRateMapping = map[string]string{
	"us-south": "na",
	"us-east":  "na",
	"ca-tor":   "na",
	"mon01":    "na",
	"sjc04":    "na",
	"eu-gb":    "eu",
	"eu-de":    "eu",
	"ams03":    "eu",
	"mil01":    "eu",
	"par01":    "eu",
	"br-sao":   "sa",
	"au-syd":   "ap",
	"jp-osa":   "ap",
	"jp-tok":   "ap",
	"che01":    "ap",
	"sng01":    "ap",
}

// PopulateUsage parses the u schema.UsageData into the IbmCosBucket.
// It uses the `infracost_usage` struct tags to populate data into the IbmCosBucket.
func (r *IbmCosBucket) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *IbmCosBucket) StandardMonthlyAverageCapacityCostComponent() *schema.CostComponent {
	var q *decimal.Decimal

	if r.MonthlyAverageCapacity != nil {
		q = decimalPtr(decimal.NewFromFloat((*r.MonthlyAverageCapacity)))
	}

	u := ""

	switch r.StorageClass {
	case "vault":
		u = "STORAGE_VAULT"
	case "standard":
		u = "STORAGE_STANDARD"
	case "cold":
		u = "STORAGE_COLD"
	case "smart":
		u = "SMARTSTORAGE_SMART"
	}

	return &schema.CostComponent{
		Name:            "Storage Capacity",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Location),
			Service:       strPtr(("cloud-object-storage")),
			ProductFamily: strPtr("iaas"),
			AttributeFilters: []*schema.AttributeFilter{
				{
					Key: "planName", Value: strPtr(r.Plan),
				},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr(u),
		},
	}
}

func (r *IbmCosBucket) StandardClassARequestCountCostComponent() *schema.CostComponent {
	var q *decimal.Decimal

	if r.ClassARequestCount != nil {
		q = decimalPtr(decimal.NewFromInt(*r.ClassARequestCount))
	}

	u := ""

	switch r.StorageClass {
	case "vault":
		u = "CLASSA_VAULT"
	case "standard":
		u = "CLASSA_STANDARD"
	case "cold":
		u = "CLASSA_COLD"
	case "smart":
		u = "CLASSA_SMART"
	}

	return &schema.CostComponent{
		Name:            "Class A requests",
		Unit:            "1k API calls",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Location),
			Service:       strPtr(("cloud-object-storage")),
			ProductFamily: strPtr("iaas"),
			AttributeFilters: []*schema.AttributeFilter{
				{
					Key: "planName", Value: strPtr(r.Plan),
				},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr(u),
		},
	}
}

func (r *IbmCosBucket) StandardClassBRequestCountCostComponent() *schema.CostComponent {
	var q *decimal.Decimal

	if r.ClassBRequestCount != nil {
		q = decimalPtr(decimal.NewFromInt(*r.ClassBRequestCount))
	}

	u := ""

	switch r.StorageClass {
	case "vault":
		u = "CLASSB_VAULT"
	case "standard":
		u = "CLASSB_STANDARD"
	case "cold":
		u = "CLASSB_COLD"
	case "smart":
		u = "CLASSB_SMART"
	}

	return &schema.CostComponent{
		Name:            "Class B requests",
		Unit:            "10k API calls",
		UnitMultiplier:  decimal.NewFromFloat(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Location),
			Service:       strPtr(("cloud-object-storage")),
			ProductFamily: strPtr("iaas"),
			AttributeFilters: []*schema.AttributeFilter{
				{
					Key: "planName", Value: strPtr(r.Plan),
				},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr(u),
		},
	}
}

func (r *IbmCosBucket) StandardEgressCostComponent() *schema.CostComponent {
	var q *decimal.Decimal

	if r.PublicStandardEgress != nil {
		q = decimalPtr(decimal.NewFromFloat(*r.PublicStandardEgress))
	}

	// using bandwidth for egress
	// https://github.ibm.com/ibmcloud/estimator/blob/f9dfa477c27bbf7570d296816bdc07b706646572/__tests__/client/fixtures/callback-estimate.json#L41
	u := ""

	switch r.StorageClass {
	case "vault":
		u = "BANDWIDTH_VAULT"
	case "standard":
		u = "BANDWIDTH_STANDARD"
	case "cold":
		u = "BANDWIDTH_COLD"
	case "smart":
		u = "BANDWIDTH_SMART"
	}

	return &schema.CostComponent{
		Name:            "Public Standard Egress",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Location),
			Service:       strPtr(("cloud-object-storage")),
			ProductFamily: strPtr("iaas"),
			AttributeFilters: []*schema.AttributeFilter{
				{
					Key: "planName", Value: strPtr(r.Plan),
				},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr(u),
		},
	}
}

func (r *IbmCosBucket) AsperaIngressCostComponent() *schema.CostComponent {
	var q *decimal.Decimal

	if r.AsperaIngress != nil {
		q = decimalPtr(decimal.NewFromInt(int64(*r.AsperaIngress)))
	}

	costComponent := schema.CostComponent{
		Name:            "Aspera Ingress Free",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Location),
			Service:       strPtr(("cloud-object-storage")),
			ProductFamily: strPtr("iaas"),
			AttributeFilters: []*schema.AttributeFilter{
				{
					Key: "planName", Value: strPtr("standard"),
				},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("ASPERA_INGRESS"),
		},
	}

	// regardless the quantity, ingress is free
	costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))

	return &costComponent
}

func (r *IbmCosBucket) AsperaEgressCostComponent() *schema.CostComponent {
	var q *decimal.Decimal

	if r.AsperaEgress != nil {
		q = decimalPtr(decimal.NewFromInt(int64(*r.AsperaEgress)))
	}

	return &schema.CostComponent{
		Name:            "Aspera Egress",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Location),
			Service:       strPtr(("cloud-object-storage")),
			ProductFamily: strPtr("iaas"),
			AttributeFilters: []*schema.AttributeFilter{
				{
					Key: "planName", Value: strPtr("standard"),
				},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("ASPERA_EGRESS"),
		},
	}
}

func (r *IbmCosBucket) ArchiveCapacityCostComponent() *schema.CostComponent {
	var q *decimal.Decimal

	if r.ArchiveCapacity != nil {
		q = decimalPtr(decimal.NewFromInt(int64(*r.ArchiveCapacity)))
	}

	var region string

	switch r.Plan {
	case "standard":
		region = r.Location
	case "cos-one-rate-plan":
		region = oneRateMapping[r.Location]
	}

	return &schema.CostComponent{
		Name:            "Archive Capacity",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(region),
			Service:       strPtr(("cloud-object-storage")),
			ProductFamily: strPtr("iaas"),
			AttributeFilters: []*schema.AttributeFilter{
				{
					Key: "planName", Value: strPtr(r.Plan),
				},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("ARCHIVE_STORAGE"),
		},
	}
}

func (r *IbmCosBucket) ArchiveRestoreCostComponent() *schema.CostComponent {
	var q *decimal.Decimal

	if r.ArchiveRestore != nil {
		q = decimalPtr(decimal.NewFromInt(int64(*r.ArchiveRestore)))
	}

	var region string

	switch r.Plan {
	case "standard":
		region = r.Location
	case "cos-one-rate-plan":
		region = oneRateMapping[r.Location]
	}

	return &schema.CostComponent{
		Name:            "Archive Restore",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(region),
			Service:       strPtr(("cloud-object-storage")),
			ProductFamily: strPtr("iaas"),
			AttributeFilters: []*schema.AttributeFilter{
				{
					Key: "planName", Value: strPtr(r.Plan),
				},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("ARCHIVE_RESTORE"),
		},
	}
}

func (r *IbmCosBucket) AcceleratedArchiveCapacityCostComponent() *schema.CostComponent {
	var q *decimal.Decimal

	if r.ArchiveCapacity != nil {
		q = decimalPtr(decimal.NewFromInt(int64(*r.ArchiveCapacity)))
	}

	var region string

	switch r.Plan {
	case "standard":
		region = r.Location
	case "cos-one-rate-plan":
		region = oneRateMapping[r.Location]
	}

	return &schema.CostComponent{
		Name:            "Accelerated Archive Capacity",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(region),
			Service:       strPtr(("cloud-object-storage")),
			ProductFamily: strPtr("iaas"),
			AttributeFilters: []*schema.AttributeFilter{
				{
					Key: "planName", Value: strPtr(r.Plan),
				},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("ACCELERATEDARCHIVE_STORAGE"),
		},
	}
}

func (r *IbmCosBucket) AcceleratedArchiveRestoreCostComponent() *schema.CostComponent {
	var q *decimal.Decimal

	if r.ArchiveRestore != nil {
		q = decimalPtr(decimal.NewFromInt(int64(*r.ArchiveRestore)))
	}

	var region string

	switch r.Plan {
	case "standard":
		region = r.Location
	case "cos-one-rate-plan":
		region = oneRateMapping[r.Location]
	}

	return &schema.CostComponent{
		Name:            "Accelerated Archive Restore",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(region),
			Service:       strPtr(("cloud-object-storage")),
			ProductFamily: strPtr("iaas"),
			AttributeFilters: []*schema.AttributeFilter{
				{
					Key: "planName", Value: strPtr(r.Plan),
				},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("ACCELERATEDARCHIVE_RESTORE"),
		},
	}
}

func (r *IbmCosBucket) MonthlyDataRetrievalCostComponent() *schema.CostComponent {
	var q *decimal.Decimal

	if r.MonthlyDataRetrieval != nil {
		q = decimalPtr(decimal.NewFromInt(int64(*r.MonthlyDataRetrieval)))
	}

	u := ""

	switch r.StorageClass {
	case "cold":
		u = "RETRIEVAL_COLD"
	case "vault":
		u = "RETRIEVAL_VAULT"
	}

	return &schema.CostComponent{
		Name:            "Data Retrieval",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Location),
			Service:       strPtr(("cloud-object-storage")),
			ProductFamily: strPtr("iaas"),
			AttributeFilters: []*schema.AttributeFilter{
				{
					Key: "planName", Value: strPtr(r.Plan),
				},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr(u),
		},
	}
}

func (r *IbmCosBucket) OneRateMonthlyAverageCapacityCostComponent() *schema.CostComponent {
	var q *decimal.Decimal

	if r.MonthlyAverageCapacity != nil {
		q = decimalPtr(decimal.NewFromFloat((*r.MonthlyAverageCapacity)))
	}

	return &schema.CostComponent{
		Name:            "Storage Capacity",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(oneRateMapping[r.Location]),
			Service:       strPtr(("cloud-object-storage")),
			ProductFamily: strPtr("iaas"),
			AttributeFilters: []*schema.AttributeFilter{
				{
					Key: "planName", Value: strPtr(r.Plan),
				},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("STORAGE_ACTIVE"),
		},
	}
}

func (r *IbmCosBucket) OneRateClassARequestCountCostComponent() []*schema.CostComponent {
	var freeQuantity *decimal.Decimal
	var pricedQuantity *decimal.Decimal

	if r.ClassARequestCount != nil && r.MonthlyAverageCapacity != nil {
		lowerBound := decimal.NewFromFloat(*r.MonthlyAverageCapacity).Mul(decimal.NewFromInt(100))
		requestCount := decimal.NewFromInt(*r.ClassARequestCount).Mul(decimal.NewFromInt(1000))
		q := decimalPtr(decimal.NewFromInt(*r.ClassARequestCount))
		if requestCount.LessThanOrEqual(lowerBound) {
			freeQuantity = q
			pricedQuantity = decimalPtr(decimal.NewFromInt(0))
		} else {
			freeQuantity = decimalPtr(lowerBound.Div(decimal.NewFromInt(1000)))
			pricedQuantity = decimalPtr(q.Sub(*freeQuantity))
		}
	}

	var costComponents []*schema.CostComponent

	freeCostComponent := &schema.CostComponent{
		Name:            "Free Class A requests",
		Unit:            "1k API calls",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: freeQuantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(oneRateMapping[r.Location]),
			Service:       strPtr(("cloud-object-storage")),
			ProductFamily: strPtr("iaas"),
			AttributeFilters: []*schema.AttributeFilter{
				{
					Key: "planName", Value: strPtr(r.Plan),
				},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("CLASSA_ACTIVE"),
		},
	}
	freeCostComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
	costComponents = append(costComponents, freeCostComponent)

	pricedCostComponent := &schema.CostComponent{
		Name:            "Class A requests",
		Unit:            "1k API calls",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: pricedQuantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(oneRateMapping[r.Location]),
			Service:       strPtr(("cloud-object-storage")),
			ProductFamily: strPtr("iaas"),
			AttributeFilters: []*schema.AttributeFilter{
				{
					Key: "planName", Value: strPtr(r.Plan),
				},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("CLASSA_ACTIVE"),
		},
	}
	costComponents = append(costComponents, pricedCostComponent)

	return costComponents
}

func (r *IbmCosBucket) OneRateClassBRequestCountCostComponent() []*schema.CostComponent {
	var freeQuantity *decimal.Decimal
	var pricedQuantity *decimal.Decimal

	if r.ClassBRequestCount != nil && r.MonthlyAverageCapacity != nil {
		lowerBound := decimal.NewFromFloat(*r.MonthlyAverageCapacity).Mul(decimal.NewFromInt(100))
		requestCount := decimal.NewFromInt(*r.ClassBRequestCount).Mul(decimal.NewFromInt(10000))
		q := decimalPtr(decimal.NewFromInt(*r.ClassBRequestCount))
		if requestCount.LessThanOrEqual(lowerBound) {
			freeQuantity = q
			pricedQuantity = decimalPtr(decimal.NewFromInt(0))
		} else {
			freeQuantity = decimalPtr(lowerBound.Div(decimal.NewFromInt(10000)))
			pricedQuantity = decimalPtr(q.Sub(*freeQuantity))
		}
	}

	var costComponents []*schema.CostComponent

	freeCostComponent := &schema.CostComponent{
		Name:            "Free Class B requests",
		Unit:            "10k API calls",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: freeQuantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(oneRateMapping[r.Location]),
			Service:       strPtr(("cloud-object-storage")),
			ProductFamily: strPtr("iaas"),
			AttributeFilters: []*schema.AttributeFilter{
				{
					Key: "planName", Value: strPtr(r.Plan),
				},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("CLASSB_ACTIVE"),
		},
	}
	freeCostComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
	costComponents = append(costComponents, freeCostComponent)

	pricedCostComponent := &schema.CostComponent{
		Name:            "Class B requests",
		Unit:            "10k API calls",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: pricedQuantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(oneRateMapping[r.Location]),
			Service:       strPtr(("cloud-object-storage")),
			ProductFamily: strPtr("iaas"),
			AttributeFilters: []*schema.AttributeFilter{
				{
					Key: "planName", Value: strPtr(r.Plan),
				},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("CLASSB_ACTIVE"),
		},
	}
	costComponents = append(costComponents, pricedCostComponent)

	return costComponents
}

func (r *IbmCosBucket) OneRateEgressCostComponent() []*schema.CostComponent {
	var freeQuantity *decimal.Decimal
	var pricedQuantity *decimal.Decimal

	if r.PublicStandardEgress != nil && r.MonthlyAverageCapacity != nil {
		lowerBound := decimal.NewFromFloat(*r.MonthlyAverageCapacity)
		q := decimalPtr(decimal.NewFromFloat(*r.PublicStandardEgress))
		if q.LessThanOrEqual(lowerBound) {
			freeQuantity = q
			pricedQuantity = decimalPtr(decimal.NewFromInt(0))
		} else {
			freeQuantity = decimalPtr(lowerBound)
			pricedQuantity = decimalPtr(q.Sub(lowerBound))
		}
	}

	var costComponents []*schema.CostComponent

	freeCostComponent := &schema.CostComponent{
		Name:            "Free Public Egress",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: freeQuantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(oneRateMapping[r.Location]),
			Service:       strPtr(("cloud-object-storage")),
			ProductFamily: strPtr("iaas"),
			AttributeFilters: []*schema.AttributeFilter{
				{
					Key: "planName", Value: strPtr(r.Plan),
				},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("CLASSB_ACTIVE"),
		},
	}
	freeCostComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
	costComponents = append(costComponents, freeCostComponent)

	pricedCostComponent := &schema.CostComponent{
		Name:            "Public Egress",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: pricedQuantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(oneRateMapping[r.Location]),
			Service:       strPtr(("cloud-object-storage")),
			ProductFamily: strPtr("iaas"),
			AttributeFilters: []*schema.AttributeFilter{
				{
					Key: "planName", Value: strPtr(r.Plan),
				},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("BANDWIDTH_ACTIVE"),
		},
	}
	costComponents = append(costComponents, pricedCostComponent)

	return costComponents
}

func (r *IbmCosBucket) LitePlanCostComponent() *schema.CostComponent {
	costComponent := schema.CostComponent{
		Name:            "Lite Plan Bucket",
		Unit:            "Instance",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
	}
	costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
	return &costComponent
}

// BuildResource builds a schema.Resource from a valid IbmCosBucket struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *IbmCosBucket) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}

	if r.Plan == "lite" {
		costComponents = append(costComponents, r.LitePlanCostComponent())
	} else if r.Plan == "standard" {
		if r.LocationIdentifier == "region_location" && r.Archive {
			if strings.ToLower(r.ArchiveType) == "glacier" {
				costComponents = append(costComponents, r.ArchiveCapacityCostComponent(), r.ArchiveRestoreCostComponent())
			} else if strings.ToLower(r.ArchiveType) == "accelerated" {
				costComponents = append(costComponents, r.AcceleratedArchiveCapacityCostComponent(), r.AcceleratedArchiveRestoreCostComponent())
			}
		} else if (r.LocationIdentifier == "region_location" || r.LocationIdentifier == "cross_region_location") && (r.AsperaEgress != nil || r.AsperaIngress != nil) {
			costComponents = append(costComponents, r.AsperaEgressCostComponent(), r.AsperaIngressCostComponent())
		}

		costComponents = append(
			costComponents,
			r.StandardMonthlyAverageCapacityCostComponent(),
			r.StandardClassARequestCountCostComponent(),
			r.StandardClassBRequestCountCostComponent(),
			r.StandardEgressCostComponent(),
		)

		if r.StorageClass == "vault" || r.StorageClass == "cold" {
			costComponents = append(costComponents, r.MonthlyDataRetrievalCostComponent())
		}
	} else if r.Plan == "cos-one-rate-plan" {
		costComponents = append(costComponents, r.OneRateMonthlyAverageCapacityCostComponent())
		costComponents = append(
			costComponents,
			r.OneRateClassARequestCountCostComponent()...,
		)
		costComponents = append(
			costComponents,
			r.OneRateClassBRequestCountCostComponent()...,
		)
		costComponents = append(
			costComponents,
			r.OneRateEgressCostComponent()...,
		)
		if r.LocationIdentifier == "region_location" && r.Archive {
			if strings.ToLower(r.ArchiveType) == "glacier" {
				costComponents = append(costComponents, r.ArchiveCapacityCostComponent(), r.ArchiveRestoreCostComponent())
			} else if strings.ToLower(r.ArchiveType) == "accelerated" {
				costComponents = append(costComponents, r.AcceleratedArchiveCapacityCostComponent(), r.AcceleratedArchiveRestoreCostComponent())
			}
		} else if (r.LocationIdentifier == "region_location" || r.LocationIdentifier == "cross_region_location") && (r.AsperaEgress != nil || r.AsperaIngress != nil) {
			costComponents = append(costComponents, r.AsperaEgressCostComponent(), r.AsperaIngressCostComponent())
		}
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    IbmCosBucketUsageSchema,
		CostComponents: costComponents,
	}
}
