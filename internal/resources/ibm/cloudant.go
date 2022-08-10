package ibm

import (
	"fmt"
	"strconv"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Cloudant struct represents a Cloudant Instance
//
// Resource information: https://www.ibm.com/cloud/cloudant
// Pricing information: https://cloud.ibm.com/catalog/services/cloudant
type Cloudant struct {
	Address  string
	Region   string
	Plan     string
	Capacity string

	Storage *int64 `infracost_usage:"storage"`
}

// CloudantUsageSchema defines a list which represents the usage schema of Cloudant.
var CloudantUsageSchema = []*schema.UsageItem{
	{Key: "storage", ValueType: schema.Int64, DefaultValue: 0},
}

// PopulateUsage parses the u schema.UsageData into the Cloudant.
// It uses the `infracost_usage` struct tags to populate data into the Cloudant.
func (r *Cloudant) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func convertCapacity(capacity string) int {
	c, err := strconv.Atoi(capacity)

	if err != nil {
		fmt.Printf("Unable to convert capacity: %s. Using capacity: 1\n", capacity)
		c = 1
	}
	return c
}

func (r *Cloudant) cloudantFreeStorageCostComponent() *schema.CostComponent {
	var q *decimal.Decimal
	if r.Storage != nil {
		q = decimalPtr(decimal.NewFromInt(*r.Storage))
		if q.GreaterThan(decimal.NewFromInt(20)) {
			q = decimalPtr(decimal.NewFromInt(20))
		}
	}

	costComponent := schema.CostComponent{
		Name:            "Free Estimated Storage (first 20GB)",
		Unit:            "GB",
		MonthlyQuantity: q,
		UnitMultiplier:  decimal.NewFromInt(1),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Region),
			Service:       strPtr("cloudantnosqldb"),
			ProductFamily: strPtr("service"),
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("GB_STORAGE_ACCRUED_PER_MONTH"),
		},
	}

	costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))

	return &costComponent
}

func (r *Cloudant) cloudantStorageCostComponent() *schema.CostComponent {

	var q *decimal.Decimal
	if r.Storage != nil {
		q = decimalPtr(decimal.NewFromInt(*r.Storage))
		if q.LessThanOrEqual(decimal.NewFromInt(20)) {
			q = decimalPtr(decimal.NewFromInt(0))
		} else {
			q = decimalPtr(q.Sub(decimal.NewFromInt(20)))
		}
	}

	return &schema.CostComponent{
		Name:            "Estimated Storage",
		Unit:            "GB",
		MonthlyQuantity: q,
		UnitMultiplier:  decimal.NewFromInt(1),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Region),
			Service:       strPtr("cloudantnosqldb"),
			ProductFamily: strPtr("service"),
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("GB_STORAGE_ACCRUED_PER_MONTH"),
		},
	}
}

func (r *Cloudant) cloudantLiteCostComponent() *schema.CostComponent {
	costComponent := schema.CostComponent{
		Name:            "Lite Cloudant",
		Unit:            "instance",
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		UnitMultiplier:  decimal.NewFromInt(1),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Region),
			Service:       strPtr("cloudantnosqldb"),
			ProductFamily: strPtr("service"),
		},
		PriceFilter: &schema.PriceFilter{},
	}

	costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))

	return &costComponent
}

func (r *Cloudant) cloudantReadsCostComponent() *schema.CostComponent {
	c := convertCapacity(r.Capacity)

	monthlyReads := c * 100

	return &schema.CostComponent{
		Name:            "Monthly Reads",
		Unit:            "reads/second",
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(int64(monthlyReads))),
		UnitMultiplier:  decimal.NewFromInt(1),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Region),
			Service:       strPtr("cloudantnosqldb"),
			ProductFamily: strPtr("service"),
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("READ_CAPACITY_ACCRUED_PER_MONTH"),
		},
	}
}

func (r *Cloudant) cloudantWritesCostComponent() *schema.CostComponent {
	c := convertCapacity(r.Capacity)

	monthlyWrites := c * 50

	return &schema.CostComponent{
		Name:            "Monthly Writes",
		Unit:            "writes/second",
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(int64(monthlyWrites))),
		UnitMultiplier:  decimal.NewFromInt(1),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Region),
			Service:       strPtr("cloudantnosqldb"),
			ProductFamily: strPtr("service"),
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("WRITE_CAPACITY_ACCRUED_PER_MONTH"),
		},
	}
}

func (r *Cloudant) cloudantGlobalQueriesCostComponent() *schema.CostComponent {
	c := convertCapacity(r.Capacity)

	monthlyGlobalQueries := c * 5

	return &schema.CostComponent{
		Name:            "Monthly Global Queries",
		Unit:            "queries/second",
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(int64(monthlyGlobalQueries))),
		UnitMultiplier:  decimal.NewFromInt(1),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Region),
			Service:       strPtr("cloudantnosqldb"),
			ProductFamily: strPtr("service"),
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("GLOBAL_QUERY_CAPACITY_ACCRUED_PER_MONTH"),
		},
	}
}

func (r *Cloudant) cloudantDedicatedHardwareCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Dedicated Hardware",
		Unit:            "instance",
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(int64(1))),
		UnitMultiplier:  decimal.NewFromInt(1),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Region),
			Service:       strPtr("cloudantnosqldb"),
			ProductFamily: strPtr("service"),
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("INSTANCES_PER_MONTH"),
		},
	}
}

// BuildResource builds a schema.Resource from a valid Cloudant struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *Cloudant) BuildResource() *schema.Resource {

	costComponents := []*schema.CostComponent{}

	if r.Plan == "lite" {
		costComponents = append(
			costComponents,
			r.cloudantLiteCostComponent(),
		)
	} else if r.Plan == "standard" {
		costComponents = append(costComponents,
			r.cloudantReadsCostComponent(),
			r.cloudantWritesCostComponent(),
			r.cloudantGlobalQueriesCostComponent(),
			r.cloudantFreeStorageCostComponent(),
			r.cloudantStorageCostComponent())
	} else if r.Plan == "dedicated-hardware" {
		costComponents = append(costComponents, r.cloudantDedicatedHardwareCostComponent())
	}

	planName := cases.Title(language.Und).String(r.Plan)

	return &schema.Resource{
		Name:           fmt.Sprintf("Cloudant - %s", planName),
		UsageSchema:    CloudantUsageSchema,
		CostComponents: costComponents,
	}
}
