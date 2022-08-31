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
	Region             string
	Location           string
	LocationIdentifier string
	StorageClass       string
	Archive            bool
	ArchiveType        string

	MonthlyAverageCapacity     *float64 `infracost_usage:"monthly_average_capacity"`
	PublicStandardEgress       *float64 `infracost_usage:"public_standard_egress"`
	AsperaIngress              *float64 `infracost_usage:"aspera_ingress"`
	AsperaEgress               *float64 `infracost_usage:"aspera_egress"`
	ArchiveCapacity            *float64 `infracost_usage:"archive_capacity"`
	ArchiveRestore             *float64 `infracost_usage:"archive_restore"`
	AcceleratedArchiveCapacity *float64 `infracost_usage:"accelerated_archive_capacity"`
	AcceleratedArchiveRestore  *float64 `infracost_usage:"accelerated_archive_restore"`
	ClassARequestCount         *int64   `infracost_usage:"class_a_request_count"`
	ClassBRequestCount         *int64   `infracost_usage:"class_b_request_count"`
	MonthlyDataRetrieval       *float64 `infracost_usage:"monthly_data_retrieval"`
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

// PopulateUsage parses the u schema.UsageData into the IbmCosBucket.
// It uses the `infracost_usage` struct tags to populate data into the IbmCosBucket.
func (r *IbmCosBucket) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *IbmCosBucket) MonthlyAverageCapacityCostComponent() *schema.CostComponent {

	var q *decimal.Decimal

	if r.MonthlyAverageCapacity != nil {
		q = decimalPtr(decimal.NewFromInt(int64(*r.MonthlyAverageCapacity)))
	}

	s := r.StorageClass
	u := ""

	switch s {
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
		Name:            "Monthly Average Capacity",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:       strPtr("ibm"),
			Region:           strPtr(r.Location),
			Service:          strPtr(("cloud-object-storage")),
			ProductFamily:    strPtr("iaas"),
			AttributeFilters: []*schema.AttributeFilter{},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr(u),
		},
	}
}

func (r *IbmCosBucket) ClassARequestCountCostComponent() *schema.CostComponent {

	var q *decimal.Decimal

	if r.ClassARequestCount != nil {
		q = decimalPtr(decimal.NewFromInt(*r.ClassARequestCount))
	}

	s := r.StorageClass
	u := ""

	switch s {
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
			VendorName:       strPtr("ibm"),
			Region:           strPtr(r.Location),
			Service:          strPtr(("cloud-object-storage")),
			ProductFamily:    strPtr("iaas"),
			AttributeFilters: []*schema.AttributeFilter{},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr(u),
		},
	}
}

func (r *IbmCosBucket) ClassBRequestCountCostComponent() *schema.CostComponent {

	var q *decimal.Decimal

	if r.ClassBRequestCount != nil {
		q = decimalPtr(decimal.NewFromInt(*r.ClassBRequestCount))
	}

	s := r.StorageClass
	u := ""

	switch s {
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
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:       strPtr("ibm"),
			Region:           strPtr(r.Location),
			Service:          strPtr(("cloud-object-storage")),
			ProductFamily:    strPtr("iaas"),
			AttributeFilters: []*schema.AttributeFilter{},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr(u),
		},
	}
}

func (r *IbmCosBucket) PublicStandardEgressCostComponent() *schema.CostComponent {

	var q *decimal.Decimal

	if r.PublicStandardEgress != nil {
		q = decimalPtr(decimal.NewFromInt(int64(*r.PublicStandardEgress)))
	}

	// using bandwidth for egress
	// https://github.ibm.com/ibmcloud/estimator/blob/f9dfa477c27bbf7570d296816bdc07b706646572/__tests__/client/fixtures/callback-estimate.json#L41
	s := r.StorageClass
	u := ""

	switch s {
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
			VendorName:       strPtr("ibm"),
			Region:           strPtr(r.Location),
			Service:          strPtr(("cloud-object-storage")),
			ProductFamily:    strPtr("iaas"),
			AttributeFilters: []*schema.AttributeFilter{},
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
			VendorName:       strPtr("ibm"),
			Region:           strPtr(r.Location),
			Service:          strPtr(("cloud-object-storage")),
			ProductFamily:    strPtr("iaas"),
			AttributeFilters: []*schema.AttributeFilter{},
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
			VendorName:       strPtr("ibm"),
			Region:           strPtr(r.Location),
			Service:          strPtr(("cloud-object-storage")),
			ProductFamily:    strPtr("iaas"),
			AttributeFilters: []*schema.AttributeFilter{},
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

	return &schema.CostComponent{
		Name:            "Archive Capacity",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:       strPtr("ibm"),
			Region:           strPtr(r.Location),
			Service:          strPtr(("cloud-object-storage")),
			ProductFamily:    strPtr("iaas"),
			AttributeFilters: []*schema.AttributeFilter{},
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

	return &schema.CostComponent{
		Name:            "Archive Restore",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:       strPtr("ibm"),
			Region:           strPtr(r.Location),
			Service:          strPtr(("cloud-object-storage")),
			ProductFamily:    strPtr("iaas"),
			AttributeFilters: []*schema.AttributeFilter{},
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

	return &schema.CostComponent{
		Name:            "Accelerated Archive Capacity",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:       strPtr("ibm"),
			Region:           strPtr(r.Location),
			Service:          strPtr(("cloud-object-storage")),
			ProductFamily:    strPtr("iaas"),
			AttributeFilters: []*schema.AttributeFilter{},
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

	return &schema.CostComponent{
		Name:            "Accelerated Archive Restore",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:       strPtr("ibm"),
			Region:           strPtr(r.Location),
			Service:          strPtr(("cloud-object-storage")),
			ProductFamily:    strPtr("iaas"),
			AttributeFilters: []*schema.AttributeFilter{},
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

	if r.StorageClass == "cold" {
		u = "RETRIEVAL_COLD"
	}

	if r.StorageClass == "vault" {
		u = "RETRIEVAL_VAULT"
	}

	return &schema.CostComponent{
		Name:            "Data Retrieval",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:       strPtr("ibm"),
			Region:           strPtr(r.Location),
			Service:          strPtr(("cloud-object-storage")),
			ProductFamily:    strPtr("iaas"),
			AttributeFilters: []*schema.AttributeFilter{},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr(u),
		},
	}
}

// BuildResource builds a schema.Resource from a valid IbmCosBucket struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *IbmCosBucket) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}

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
		r.MonthlyAverageCapacityCostComponent(),
		r.ClassARequestCountCostComponent(),
		r.ClassBRequestCountCostComponent(),
		r.PublicStandardEgressCostComponent(),
	)

	if r.StorageClass == "vault" || r.StorageClass == "cold" {
		costComponents = append(costComponents, r.MonthlyDataRetrievalCostComponent())
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    IbmCosBucketUsageSchema,
		CostComponents: costComponents,
	}
}
