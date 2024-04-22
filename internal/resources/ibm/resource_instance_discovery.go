package ibm

import (
	"fmt"
	"math"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

const ENTERPRISE_PLAN_PROGRAMMATIC_NAME string = "enterprise"
const PLUS_PLAN_PROGRAMMATIC_NAME string = "plus"

/*
 * Plus = 'plus'
 * Enterprise = 'enterprise'
 * Premium = 'premium' (not applicable, need to call for pricing)
 */
func GetWDCostComponents(r *ResourceInstance) []*schema.CostComponent {
	if r.Plan == PLUS_PLAN_PROGRAMMATIC_NAME {
		return []*schema.CostComponent{
			WDInstanceCostComponent(r),
			WDMonthlyDocumentsCostComponent(r),
			WDMonthlyQueriesCostComponent(r),
		}
	} else if r.Plan == ENTERPRISE_PLAN_PROGRAMMATIC_NAME {
		return []*schema.CostComponent{
			WDInstanceCostComponent(r),
			WDMonthlyDocumentsCostComponent(r),
			WDMonthlyQueriesCostComponent(r),
			WDMonthlyCustomModelsCostComponent(r),
			WDMonthlyCollectionsCostComponent(r),
		}
	} else {
		costComponent := schema.CostComponent{
			Name:            fmt.Sprintf("Plan %s with customized pricing", r.Plan),
			UnitMultiplier:  decimal.NewFromInt(1), // Final quantity for this cost component will be divided by this amount
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		}
		costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
		return []*schema.CostComponent{
			&costComponent,
		}
	}
}

/*
 * Instance
 * - Plus: $USD/instance/month
 * - Enterprise: $USD/instance/month
 */
func WDInstanceCostComponent(r *ResourceInstance) *schema.CostComponent {

	var instances_unit_name string
	var quantity *decimal.Decimal

	if r.Plan == PLUS_PLAN_PROGRAMMATIC_NAME {
		instances_unit_name = "PLUS_SERVICE_INSTANCES_PER_MONTH"
	} else {
		instances_unit_name = "ENTERPRISE_SERVICE_INSTANCES_PER_MONTH"
	}

	if r.WD_Instance != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.WD_Instance))
	} else {
		quantity = decimalPtr(decimal.NewFromInt(1))
	}

	return &schema.CostComponent{
		Name:            "Instance",
		Unit:            "Instance",
		UnitMultiplier:  decimal.NewFromInt(1), // Final quantity for this cost component will be divided by this amount
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr(instances_unit_name),
		},
	}
}

/*
 * Documents:
 * - Plus: $USD/documents/month. Includes 10,000 documents per month; $USD for every additional 1,000 documents.
 * - Enterprise: $USD/documents/month. Includes 100,000 documents per month; $USD for every additional 1,000 documents.
 */
func WDMonthlyDocumentsCostComponent(r *ResourceInstance) *schema.CostComponent {

	var documents_included int     // Base number of documents that are included with an instance and do not have a cost
	var documents_unit_name string // Unit to display
	var quantity *decimal.Decimal  // Quantity of current cost component (e.g. number of additional 1000 "blocks" over the base number of documents included)

	if r.Plan == PLUS_PLAN_PROGRAMMATIC_NAME {
		documents_unit_name = "PLUS_DOCUMENTS_TOTAL"
		documents_included = 10000
	} else {
		documents_unit_name = "ENTERPRISE_DOCUMENTS_TOTAL"
		documents_included = 100000
	}

	if r.WD_Documents != nil {

		additional_documents := *r.WD_Documents - float64(documents_included)

		if additional_documents > 0 {
			quantity = decimalPtr(decimal.NewFromFloat(additional_documents))
		}
	} else {
		quantity = decimalPtr(decimal.NewFromInt(0))
	}

	return &schema.CostComponent{
		Name:            "Additional Monthly Documents",
		Unit:            "Documents",
		UnitMultiplier:  decimal.NewFromFloat(1), // Final quantity for this cost component will be divided by this amount
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr(documents_unit_name),
		},
	}
}

/*
 * Queries:
 * - Plus: $USD/queries/month. Includes 10,000 queries per month; $USD for every additional 1,000 queries.
 * - Enterprise: $USD/queries/month. Includes 100,000 queries per month; $USD for every additional 1,000 queries.
 */
func WDMonthlyQueriesCostComponent(r *ResourceInstance) *schema.CostComponent {

	var quantity *decimal.Decimal
	var queries_included int // Base number of queries that are included with an instance and do not have a cost

	if r.Plan == PLUS_PLAN_PROGRAMMATIC_NAME {
		queries_included = 10000
	} else {
		queries_included = 100000
	}

	if r.WD_Queries != nil {

		additional_queries := *r.WD_Queries - float64(queries_included)

		if additional_queries > 0 {
			quantity = decimalPtr(decimal.NewFromFloat(additional_queries))
		}
	} else {
		quantity = decimalPtr(decimal.NewFromInt(0))
	}

	return &schema.CostComponent{
		Name:            "Additional Monthly Queries",
		Unit:            "Queries",
		UnitMultiplier:  decimal.NewFromInt(1), // Final quantity for this cost component will be divided by this amount
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("GRADUATED_PRICE_QUERIES_PER_MONTH"),
		},
	}
}

/*
 * Custom Models:
 * - Enterprise: $USD/additional custom models/month. Includes 3 custom models per month.
 */
func WDMonthlyCustomModelsCostComponent(r *ResourceInstance) *schema.CostComponent {

	var quantity *decimal.Decimal
	var custom_models_included int = 3 // Base number of custom models that are included with an instance and do not have a cost

	if r.WD_CustomModels != nil {
		// Determine number of custom models that go over the base number of custom models included
		quantity = decimalPtr(decimal.NewFromFloat(*r.WD_CustomModels - float64(custom_models_included)))
	} else {
		quantity = decimalPtr(decimal.NewFromInt(0))
	}

	return &schema.CostComponent{
		Name:            "Additional Monthly Custom Models",
		Unit:            "Custom Models",
		UnitMultiplier:  decimal.NewFromInt(1), // Final quantity for this cost component will be divided by this amount
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("CUSTOM_MODELS_PER_MONTH"),
		},
	}
}

/*
 * Collections
 * - Enterprise: $USD/additional collections/month. Includes 300 collections per month; $USD for every additional 100 collections.
 */
func WDMonthlyCollectionsCostComponent(r *ResourceInstance) *schema.CostComponent {

	var quantity *decimal.Decimal
	var collections_additional_range int = 100 // Additional cost for every 100 over the included amount of collections
	var collections_included int = 300         // Base number of collections that are included with an instance and do not have a cost

	if r.WD_Collections != nil {

		additional_collections := *r.WD_Collections - float64(collections_included)

		if additional_collections > 0 {
			// Determine number of 100 "blocks" of collections go over the base number of collections included
			quantity = decimalPtr(decimal.NewFromFloat(math.Ceil(additional_collections / float64(collections_additional_range))))
		}

	} else {
		quantity = decimalPtr(decimal.NewFromInt(0))
	}

	return &schema.CostComponent{
		Name:            "Additional Monthly Collections",
		Unit:            "Hundred Collections",
		UnitMultiplier:  decimal.NewFromFloat(1), // Final quantity for this cost component will be divided by this amount
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("ENTERPRISE_COLLECTIONS_TOTAL"),
		},
	}
}
