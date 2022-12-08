package azure

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

const (
	sqlServiceName   = "SQL Database"
	sqlProductFamily = "Databases"

	sqlServerlessTier = "general purpose - serverless"
	sqlHyperscaleTier = "hyperscale"
)

var (
	tierMapping = map[string]string{
		"p": "Premium",
		"s": "Standard",
	}
)

// SQLDatabase represents an azure sql database instance.
//
// More resource information here: https://azure.microsoft.com/en-gb/products/azure-sql/database/
// Pricing information here: https://azure.microsoft.com/en-gb/pricing/details/azure-sql-database/single/
type SQLDatabase struct {
	Address          string
	Region           string
	SKU              string
	LicenceType      string
	Tier             string
	Family           string
	Cores            *int64
	MaxSizeGB        *float64
	ReadReplicaCount *int64
	ZoneRedundant    bool

	// ExtraDataStorageGB represents a usage cost of additional backup storage used by the sql database.
	ExtraDataStorageGB *int64 `infracost_usage:"extra_data_storage_gb"`
	// MonthlyVCoreHours represents a usage param that allows users to define how many hours of usage a serverless sql database instance uses.
	MonthlyVCoreHours *int64 `infracost_usage:"monthly_vcore_hours"`
	// LongTermRetentionStorageGB defines a usage param that allows users to define how many gb of cold storage the database uses.
	// This is storage that can be kept for up to 10 years.
	LongTermRetentionStorageGB *int64 `infracost_usage:"long_term_retention_storage_gb"`
}

// PopulateUsage parses the u schema.UsageData into the SQLDatabase.
func (r *SQLDatabase) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid SQLDatabase.
// It returns a SQLDatabase as a *schema.Resource with cost components initialized.
//
// SQLDatabase splits pricing into two different models. DTU & vCores.
//
//	Database Transaction Unit (DTU) is made a performance metric representing a mixture of performance metrics
//	in azure sql. Some include: CPU, I/O, Memory. DTU is used as Azure tries to simplify billing by using a single metric.
//
//	Virtual Core (vCore) pricing is designed to translate from on premise hardware metrics (cores) into the cloud
//	sql instance. vCore is designed to allow users to better estimate their resource limits, e.g. RAM.
//
// SQL databases that follow a DTU pricing model have the following costs associated with them:
//
//  1. Costs based on the number of DTUs that the sql database has
//  2. Extra backup data costs - this is configured using SQLDatabase.ExtraDataStorageGB
//  3. Long term data backup costs - this is configured using SQLDatabase.LongTermRetentionStorageGB
//
// SQL databases that follow a vCore pricing model have the following costs associated with them:
//
//  1. Costs based on the number of vCores the resource has
//  2. Extra pricing if any database read replicas have been provisioned
//  3. Additional charge for sql server licencing based on vCores amount
//  4. Charges for storage used
//  5. Charges for long term data backup - this is configured using SQLDatabase.LongTermRetentionStorageGB
//
// This method is called after the resource is initialized by an IaC provider. SQLDatabase is used by both mssql_database
// and sql_database terraform resources to build a sql database costing.
func (r *SQLDatabase) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name: r.Address,
		UsageSchema: []*schema.UsageItem{
			{Key: "extra_data_storage_gb", DefaultValue: 0, ValueType: schema.Int64},
			{Key: "monthly_vcore_hours", DefaultValue: 0, ValueType: schema.Int64},
			{Key: "long_term_retention_storage_gb", DefaultValue: 0, ValueType: schema.Int64},
		},
		CostComponents: r.costComponents(),
	}
}

func (r *SQLDatabase) costComponents() []*schema.CostComponent {
	if r.Cores != nil {
		return r.vCorePurchaseCostComponents()
	}

	return r.dtuPurchaseCostComponents()
}

func (r *SQLDatabase) dtuPurchaseCostComponents() []*schema.CostComponent {
	skuName := strings.ToLower(r.SKU)
	if skuName == "basic" {
		skuName = "b"
	}

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("Compute (%s)", strings.ToTitle(r.SKU)),
			Unit:           "days",
			UnitMultiplier: decimal.NewFromInt(1),
			// This is not the same as the 730h/month value we use elsewhere but it looks more understandable than seeing `30.4166` in the output
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(30)),
			ProductFilter: r.productFilter([]*schema.AttributeFilter{
				{Key: "productName", ValueRegex: regexPtr("^SQL Database Single")},
				{Key: "skuName", ValueRegex: regexPtr(fmt.Sprintf("^%s$", skuName))},
				{Key: "meterName", ValueRegex: regexPtr("DTU(s)?$")},
			}),
			PriceFilter: priceFilterConsumption,
		},
	}

	if skuName != "b" {
		costComponents = append(costComponents, r.extraDataStorageCostComponent())
	}

	costComponents = append(costComponents, r.longTermRetentionMSSQLCostComponent())

	return costComponents
}

func (r *SQLDatabase) extraDataStorageCostComponent() *schema.CostComponent {
	sn := tierMapping[strings.ToLower(r.SKU)[:1]]

	var storageGB *decimal.Decimal
	if r.MaxSizeGB != nil {
		storageGB = decimalPtr(decimal.NewFromFloat(*r.MaxSizeGB))

		if strings.ToLower(sn) == "premium" {
			storageGB = decimalPtr(storageGB.Sub(decimal.NewFromInt(500)))
		} else {
			storageGB = decimalPtr(storageGB.Sub(decimal.NewFromInt(250)))
		}

		if storageGB.IsNegative() {
			storageGB = nil
		}
	}

	if r.ExtraDataStorageGB != nil {
		storageGB = decimalPtr(decimal.NewFromInt(*r.ExtraDataStorageGB))
	}

	return &schema.CostComponent{
		Name:            "Extra data storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: storageGB,
		ProductFilter: r.productFilter([]*schema.AttributeFilter{
			{Key: "productName", ValueRegex: strPtr(fmt.Sprintf("/SQL Database %s - Storage/i", sn))},
			{Key: "skuName", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", sn))},
			{Key: "meterName", Value: strPtr("Data Stored")},
		}),
		PriceFilter: priceFilterConsumption,
	}
}

func (r *SQLDatabase) vCorePurchaseCostComponents() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{
		r.computeHoursCostComponent(),
	}

	if strings.ToLower(r.Tier) == sqlHyperscaleTier {
		costComponents = append(costComponents, r.readReplicaCostComponent())
	}

	if strings.ToLower(r.Tier) != sqlServerlessTier && strings.ToLower(r.LicenceType) == "licenseincluded" {
		costComponents = append(costComponents, r.sqlLicenseCostComponent())
	}

	costComponents = append(costComponents, r.mssqlStorageComponent())

	if strings.ToLower(r.Tier) != sqlHyperscaleTier {
		costComponents = append(costComponents, r.longTermRetentionMSSQLCostComponent())
	}

	return costComponents
}

func (r *SQLDatabase) computeHoursCostComponent() *schema.CostComponent {
	if strings.ToLower(r.Tier) == sqlServerlessTier {
		return r.serverlessComputeHoursCostComponent()
	}

	return r.provisionedComputeCostComponent()
}

func (r *SQLDatabase) serverlessComputeHoursCostComponent() *schema.CostComponent {
	productNameRegex := fmt.Sprintf("/%s - %s/", r.Tier, r.Family)

	var vCoreHours *decimal.Decimal
	if r.MonthlyVCoreHours != nil {
		vCoreHours = decimalPtr(decimal.NewFromInt(*r.MonthlyVCoreHours))
	}

	serverlessSkuName := r.mssqlSkuName(1)
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Compute (serverless, %s)", r.SKU),
		Unit:            "vCore-hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: vCoreHours,
		ProductFilter: r.productFilter([]*schema.AttributeFilter{
			{Key: "productName", ValueRegex: strPtr(productNameRegex)},
			{Key: "skuName", Value: strPtr(serverlessSkuName)},
		}),
		PriceFilter: priceFilterConsumption,
	}
}

func (r *SQLDatabase) provisionedComputeCostComponent() *schema.CostComponent {
	skuName := r.mssqlSkuName(*r.Cores)
	productNameRegex := fmt.Sprintf("/%s - %s/", r.Tier, r.Family)
	name := fmt.Sprintf("Compute (provisioned, %s)", r.SKU)

	log.Warnf("'Multiple products found' are safe to ignore for '%s' due to limitations in the Azure API.", name)

	return &schema.CostComponent{
		Name:           name,
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: r.productFilter([]*schema.AttributeFilter{
			{Key: "productName", ValueRegex: strPtr(productNameRegex)},
			{Key: "skuName", Value: strPtr(skuName)},
		}),
		PriceFilter: priceFilterConsumption,
	}
}

func (r *SQLDatabase) readReplicaCostComponent() *schema.CostComponent {
	productNameRegex := fmt.Sprintf("/%s - %s/", r.Tier, r.Family)
	skuName := r.mssqlSkuName(*r.Cores)

	var replicaCount *decimal.Decimal
	if r.ReadReplicaCount != nil {
		replicaCount = decimalPtr(decimal.NewFromInt(*r.ReadReplicaCount))
	}

	return &schema.CostComponent{
		Name:           "Read replicas",
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: replicaCount,
		ProductFilter: r.productFilter([]*schema.AttributeFilter{
			{Key: "productName", ValueRegex: strPtr(productNameRegex)},
			{Key: "skuName", Value: strPtr(skuName)},
		}),
		PriceFilter: priceFilterConsumption,
	}
}

func (r *SQLDatabase) sqlLicenseCostComponent() *schema.CostComponent {
	licenseRegion := "Global"
	if strings.Contains(r.Region, "usgov") {
		licenseRegion = "US Gov"
	}

	if strings.Contains(r.Region, "china") {
		licenseRegion = "China"
	}

	if strings.Contains(r.Region, "germany") {
		licenseRegion = "Germany"
	}

	return &schema.CostComponent{
		Name:           "SQL license",
		Unit:           "vCore-hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(*r.Cores)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr(vendorName),
			Region:        strPtr(licenseRegion),
			Service:       strPtr(sqlServiceName),
			ProductFamily: strPtr(sqlProductFamily),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", ValueRegex: strPtr(fmt.Sprintf("/%s - %s/", r.Tier, "SQL License"))},
			},
		},
		PriceFilter: priceFilterConsumption,
	}
}

func (r *SQLDatabase) mssqlStorageComponent() *schema.CostComponent {
	storageGB := decimalPtr(decimal.NewFromInt(5))
	if r.MaxSizeGB != nil {
		storageGB = decimalPtr(decimal.NewFromFloat(*r.MaxSizeGB))
	}

	storageTier := r.Tier
	if strings.ToLower(storageTier) == "general purpose - serverless" {
		storageTier = "General Purpose"
	}

	skuName := storageTier
	if r.ZoneRedundant {
		skuName += " Zone Redundancy"
	}

	productNameRegex := fmt.Sprintf("/%s - Storage/", storageTier)

	return &schema.CostComponent{
		Name:            "Storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: storageGB,
		ProductFilter: r.productFilter([]*schema.AttributeFilter{
			{Key: "productName", ValueRegex: strPtr(productNameRegex)},
			{Key: "skuName", Value: strPtr(skuName)},
			{Key: "meterName", ValueRegex: regexPtr("Data Stored$")},
		}),
	}
}

func (r *SQLDatabase) longTermRetentionMSSQLCostComponent() *schema.CostComponent {
	var retention *decimal.Decimal
	if r.LongTermRetentionStorageGB != nil {
		retention = decimalPtr(decimal.NewFromInt(*r.LongTermRetentionStorageGB))
	}

	return &schema.CostComponent{
		Name:            "Long-term retention",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: retention,
		ProductFilter: r.productFilter([]*schema.AttributeFilter{
			{Key: "productName", Value: strPtr("SQL Database - LTR Backup Storage")},
			{Key: "skuName", Value: strPtr("Backup RA-GRS")},
			{Key: "meterName", ValueRegex: strPtr("/RA-GRS Data Stored/i")},
		}),
		PriceFilter: priceFilterConsumption,
	}
}

func (r *SQLDatabase) mssqlSkuName(cores int64) string {
	sku := fmt.Sprintf("%d vCore", cores)

	if r.ZoneRedundant {
		sku += " Zone Redundancy"
	}
	return sku
}

func (r *SQLDatabase) productFilter(filters []*schema.AttributeFilter) *schema.ProductFilter {
	return &schema.ProductFilter{
		VendorName:       strPtr(vendorName),
		Region:           strPtr(r.Region),
		Service:          strPtr(sqlServiceName),
		ProductFamily:    strPtr(sqlProductFamily),
		AttributeFilters: filters,
	}
}
