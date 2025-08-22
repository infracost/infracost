package azure

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

const (
	sqlServerlessTier     = "general purpose - serverless"
	sqlHyperscaleTier     = "hyperscale"
	sqlGeneralPurposeTier = "general purpose"
)

var (
	mssqlTierMapping = map[string]string{
		"b": "Basic",
		"p": "Premium",
		"s": "Standard",
	}

	mssqlPremiumDTUIncludedStorage = map[string]float64{
		"p1":  500,
		"p2":  500,
		"p4":  500,
		"p6":  500,
		"p11": 4096,
		"p15": 4096,
	}

	mssqlStorageRedundancyTypeMapping = map[string]string{
		"geo":   "RA-GRS",
		"local": "LRS",
		"zone":  "ZRS",
	}
)

// SQLDatabase represents an Azure SQL database instance.
//
// More resource information here: https://azure.microsoft.com/en-gb/products/azure-sql/database/
// Pricing information here: https://azure.microsoft.com/en-gb/pricing/details/azure-sql-database/single/
type SQLDatabase struct {
	Address           string
	Region            string
	SKU               string
	IsElasticPool     bool
	LicenseType       string
	Tier              string
	Family            string
	Cores             *int64
	MaxSizeGB         *float64
	ReadReplicaCount  *int64
	ZoneRedundant     bool
	BackupStorageType string
	IsDevTest         bool

	// ExtraDataStorageGB represents a usage cost of additional backup storage used by the sql database.
	ExtraDataStorageGB *float64 `infracost_usage:"extra_data_storage_gb"`
	// MonthlyVCoreHours represents a usage param that allows users to define how many hours of usage a serverless sql database instance uses.
	MonthlyVCoreHours *int64 `infracost_usage:"monthly_vcore_hours"`
	// LongTermRetentionStorageGB defines a usage param that allows users to define how many GB of cold storage the database uses.
	// This is storage that can be kept for up to 10 years.
	LongTermRetentionStorageGB *int64 `infracost_usage:"long_term_retention_storage_gb"`
	// BackupStorageGB defines a usage param that allows users to define how many GB Point-In-Time Restore (PITR) backup storage the database uses.
	BackupStorageGB *int64 `infracost_usage:"backup_storage_gb"`
}

// PopulateUsage parses the u schema.UsageData into the SQLDatabase.
func (r *SQLDatabase) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *SQLDatabase) CoreType() string {
	return "SQLDatabase"
}

func (r *SQLDatabase) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "extra_data_storage_gb", DefaultValue: 0.0, ValueType: schema.Float64},
		{Key: "monthly_vcore_hours", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "long_term_retention_storage_gb", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "backup_storage_gb", DefaultValue: 0, ValueType: schema.Int64},
	}
}

// BuildResource builds a schema.Resource from a valid SQLDatabase.
// It returns a SQLDatabase as a *schema.Resource with cost components initialized.
//
// SQLDatabase splits pricing into two different models. DTU & vCores.
//
//	Database Transaction Unit (DTU) is made a performance metric representing a mixture of performance metrics
//	in Azure SQL. Some include: CPU, I/O, Memory. DTU is used as Azure tries to simplify billing by using a single metric.
//
//	Virtual Core (vCore) pricing is designed to translate from on premise hardware metrics (cores) into the cloud
//	SQL instance. vCore is designed to allow users to better estimate their resource limits, e.g. RAM.
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
//  3. Additional charge for SQL Server licensing based on vCores amount
//  4. Charges for storage used
//  5. Charges for long term data backup - this is configured using SQLDatabase.LongTermRetentionStorageGB
//
// This method is called after the resource is initialized by an IaC provider. SQLDatabase is used by both mssql_database
// and sql_database Terraform resources.
func (r *SQLDatabase) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: r.costComponents(),
	}
}

func (r *SQLDatabase) costComponents() []*schema.CostComponent {
	if r.IsElasticPool {
		return r.elasticPoolCostComponents()
	}

	if r.Cores != nil {
		return r.vCoreCostComponents()
	}

	return r.dtuCostComponents()
}

func (r *SQLDatabase) dtuCostComponents() []*schema.CostComponent {
	skuName := strings.ToLower(r.SKU)
	if skuName == "basic" {
		skuName = "b"
	}

	// This is a bit of a hack, but the Azure pricing API returns the price per day
	// and the Azure pricing calculator uses 730 hours to show the cost
	// so we need to convert the price per day to price per hour.
	// Use precision 24 to avoid rounding errors later since the default decimal precision is 16.
	daysInMonth := schema.HourToMonthUnitMultiplier.DivRound(decimal.NewFromInt(24), 24)

	name := fmt.Sprintf("Compute (%s)", strings.ToTitle(r.SKU))
	purchaseOption := priceFilterConsumption
	if r.IsDevTest {
		name = fmt.Sprintf("Compute (dev/test, %s)", strings.ToTitle(r.SKU))
		purchaseOption = priceFilterDevTestConsumption
	}

	costComponents := []*schema.CostComponent{
		{
			Name:            name,
			Unit:            "hours",
			UnitMultiplier:  daysInMonth.DivRound(schema.HourToMonthUnitMultiplier, 24),
			MonthlyQuantity: decimalPtr(daysInMonth),
			ProductFilter: r.productFilter([]*schema.AttributeFilter{
				{Key: "productName", ValueRegex: regexPtr("^SQL Database Single")},
				{Key: "skuName", ValueRegex: regexPtr(fmt.Sprintf("^%s$", skuName))},
				{Key: "meterName", ValueRegex: regexPtr("DTU(s)?$")},
			}),
			PriceFilter: purchaseOption,
		},
	}

	var extraStorageGB float64

	if !strings.HasPrefix(skuName, "b") && r.ExtraDataStorageGB != nil {
		extraStorageGB = *r.ExtraDataStorageGB
	} else if strings.HasPrefix(skuName, "s") && r.MaxSizeGB != nil {
		includedStorageGB := 250.0
		extraStorageGB = *r.MaxSizeGB - includedStorageGB
	} else if strings.HasPrefix(skuName, "p") && r.MaxSizeGB != nil {
		includedStorageGB, ok := mssqlPremiumDTUIncludedStorage[skuName]
		if ok {
			extraStorageGB = *r.MaxSizeGB - includedStorageGB
		}
	}

	if extraStorageGB > 0 {
		c := r.extraDataStorageCostComponent(extraStorageGB)
		if c != nil {
			costComponents = append(costComponents, c)
		}
	}

	costComponents = append(costComponents, r.longTermRetentionCostComponent())
	costComponents = append(costComponents, r.pitrBackupCostComponent())

	return costComponents
}

func (r *SQLDatabase) vCoreCostComponents() []*schema.CostComponent {
	costComponents := r.computeHoursCostComponents()

	if strings.ToLower(r.Tier) == sqlHyperscaleTier {
		costComponents = append(costComponents, r.readReplicaCostComponent())
	}

	if strings.ToLower(r.Tier) != sqlServerlessTier && strings.ToLower(r.LicenseType) == "licenseincluded" {
		costComponents = append(costComponents, r.sqlLicenseCostComponent())
	}

	costComponents = append(costComponents, r.storageCostComponent())

	if strings.ToLower(r.Tier) != sqlHyperscaleTier {
		costComponents = append(costComponents, r.longTermRetentionCostComponent())
		costComponents = append(costComponents, r.pitrBackupCostComponent())
	}

	return costComponents
}

func (r *SQLDatabase) elasticPoolCostComponents() []*schema.CostComponent {
	return []*schema.CostComponent{
		r.longTermRetentionCostComponent(),
		r.pitrBackupCostComponent(),
	}
}

func (r *SQLDatabase) computeHoursCostComponents() []*schema.CostComponent {
	if strings.ToLower(r.Tier) == sqlServerlessTier {
		return r.serverlessComputeHoursCostComponents()
	}

	return r.provisionedComputeCostComponents()
}

func (r *SQLDatabase) serverlessComputeHoursCostComponents() []*schema.CostComponent {
	productNameRegex := fmt.Sprintf("/%s - %s/", r.Tier, r.Family)

	var vCoreHours *decimal.Decimal
	if r.MonthlyVCoreHours != nil {
		vCoreHours = decimalPtr(decimal.NewFromInt(*r.MonthlyVCoreHours))
	}

	name := fmt.Sprintf("Compute (serverless, %s)", r.SKU)
	purchaseOption := priceFilterConsumption
	if r.IsDevTest && strings.ToLower(r.LicenseType) != "licenseincluded" {
		name = fmt.Sprintf("Compute (dev/test, serverless, %s)", r.SKU)
		purchaseOption = priceFilterDevTestConsumption
	}

	costComponents := []*schema.CostComponent{
		{
			Name:            name,
			Unit:            "vCore-hours",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: vCoreHours,
			ProductFilter: r.productFilter([]*schema.AttributeFilter{
				{Key: "productName", ValueRegex: strPtr(productNameRegex)},
				{Key: "skuName", Value: strPtr("1 vCore")},
				{Key: "meterName", ValueRegex: regexPtr("^(?!.* - Free$).*$")},
			}),
			PriceFilter: purchaseOption,
			UsageBased:  true,
		},
	}

	// Zone redundancy is free for premium and business critical tiers
	if r.ZoneRedundant {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            fmt.Sprintf("Zone redundancy (serverless, %s)", r.SKU),
			Unit:            "vCore-hours",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: vCoreHours,
			ProductFilter: r.productFilter([]*schema.AttributeFilter{
				{Key: "productName", ValueRegex: strPtr(productNameRegex)},
				{Key: "skuName", Value: strPtr("1 vCore Zone Redundancy")},
				{Key: "meterName", ValueRegex: regexPtr("^(?!.* - Free$).*$")},
			}),
			PriceFilter: priceFilterConsumption,
		})
	}

	return costComponents
}

func (r *SQLDatabase) provisionedComputeCostComponents() []*schema.CostComponent {
	var cores int64
	if r.Cores != nil {
		cores = *r.Cores
	}

	productNameRegex := fmt.Sprintf("/%s - %s/", r.Tier, r.Family)
	name := fmt.Sprintf("Compute (provisioned, %s)", r.SKU)
	purchaseOption := priceFilterConsumption
	if r.IsDevTest && strings.ToLower(r.LicenseType) != "licenseincluded" {
		name = fmt.Sprintf("Compute (dev/test, provisioned, %s)", r.SKU)
		purchaseOption = priceFilterDevTestConsumption
	}

	costComponents := []*schema.CostComponent{
		{
			Name:           name,
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: r.productFilter([]*schema.AttributeFilter{
				{Key: "productName", ValueRegex: strPtr(productNameRegex)},
				{Key: "skuName", Value: strPtr(fmt.Sprintf("%d vCore", cores))},
			}),
			PriceFilter: purchaseOption,
		},
	}

	// Zone redundancy is free for premium and business critical tiers
	if strings.EqualFold(r.Tier, sqlGeneralPurposeTier) && r.ZoneRedundant {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:           fmt.Sprintf("Zone redundancy (provisioned, %s)", r.SKU),
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: r.productFilter([]*schema.AttributeFilter{
				{Key: "productName", ValueRegex: strPtr(productNameRegex)},
				{Key: "skuName", Value: strPtr(fmt.Sprintf("%d vCore Zone Redundancy", cores))},
			}),
			PriceFilter: priceFilterConsumption,
		})
	}

	return costComponents
}

func (r *SQLDatabase) readReplicaCostComponent() *schema.CostComponent {
	productNameRegex := fmt.Sprintf("/%s - %s/", r.Tier, r.Family)
	skuName := mssqlSkuName(*r.Cores, r.ZoneRedundant)

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

func (r *SQLDatabase) longTermRetentionCostComponent() *schema.CostComponent {
	var retention *decimal.Decimal
	if r.LongTermRetentionStorageGB != nil {
		retention = decimalPtr(decimal.NewFromInt(*r.LongTermRetentionStorageGB))
	}

	redundancyType, ok := mssqlStorageRedundancyTypeMapping[strings.ToLower(r.BackupStorageType)]
	if !ok {
		logging.Logger.Warn().Msgf("Unrecognized backup storage type '%s'", r.BackupStorageType)
		redundancyType = "RA-GRS"
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Long-term retention (%s)", redundancyType),
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: retention,
		ProductFilter: r.productFilter([]*schema.AttributeFilter{
			{Key: "productName", Value: strPtr("SQL Database - LTR Backup Storage")},
			{Key: "skuName", Value: strPtr(fmt.Sprintf("Backup %s", redundancyType))},
			{Key: "meterName", ValueRegex: regexPtr(fmt.Sprintf("%s Data Stored", redundancyType))},
		}),
		PriceFilter: priceFilterConsumption,
		UsageBased:  true,
	}
}

func (r *SQLDatabase) pitrBackupCostComponent() *schema.CostComponent {
	var pitrGB *decimal.Decimal
	if r.BackupStorageGB != nil {
		pitrGB = decimalPtr(decimal.NewFromInt(*r.BackupStorageGB))
	}

	redundancyType, ok := mssqlStorageRedundancyTypeMapping[strings.ToLower(r.BackupStorageType)]
	if !ok {
		logging.Logger.Warn().Msgf("Unrecognized backup storage type '%s'", r.BackupStorageType)
		redundancyType = "RA-GRS"
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("PITR backup storage (%s)", redundancyType),
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: pitrGB,
		ProductFilter: r.productFilter([]*schema.AttributeFilter{
			{Key: "productName", ValueRegex: regexPtr("PITR Backup Storage")},
			{Key: "skuName", Value: strPtr(fmt.Sprintf("Backup %s", redundancyType))},
			{Key: "meterName", ValueRegex: regexPtr(fmt.Sprintf("%s Data Stored", redundancyType))},
		}),
		PriceFilter: priceFilterConsumption,
		UsageBased:  true,
	}
}

func (r *SQLDatabase) extraDataStorageCostComponent(extraStorageGB float64) *schema.CostComponent {
	tier := r.Tier
	if tier == "" {
		var ok bool
		tier, ok = mssqlTierMapping[strings.ToLower(r.SKU)[:1]]

		if !ok {
			logging.Logger.Warn().Msgf("Unrecognized tier for SKU '%s' for resource %s", r.SKU, r.Address)
			return nil
		}
	}

	return mssqlExtraDataStorageCostComponent(r.Region, tier, extraStorageGB)
}

func (r *SQLDatabase) sqlLicenseCostComponent() *schema.CostComponent {
	return mssqlLicenseCostComponent(r.Region, r.Cores, r.Tier)
}

func (r *SQLDatabase) storageCostComponent() *schema.CostComponent {
	return mssqlStorageCostComponent(r.Region, r.Tier, r.ZoneRedundant, r.MaxSizeGB)
}

func (r *SQLDatabase) productFilter(filters []*schema.AttributeFilter) *schema.ProductFilter {
	return mssqlProductFilter(r.Region, filters)
}

func mssqlSkuName(cores int64, zoneRedundant bool) string {
	sku := fmt.Sprintf("%d vCore", cores)

	if zoneRedundant {
		sku += " Zone Redundancy"
	}
	return sku
}

func mssqlProductFilter(region string, filters []*schema.AttributeFilter) *schema.ProductFilter {
	return &schema.ProductFilter{
		VendorName:       strPtr(vendorName),
		Region:           strPtr(region),
		Service:          strPtr("SQL Database"),
		ProductFamily:    strPtr("Databases"),
		AttributeFilters: filters,
	}
}

func mssqlExtraDataStorageCostComponent(region string, tier string, extraStorageGB float64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Extra data storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromFloat(extraStorageGB)),
		ProductFilter: mssqlProductFilter(region, []*schema.AttributeFilter{
			{Key: "productName", ValueRegex: strPtr(fmt.Sprintf("/SQL Database %s - Storage/i", tier))},
			{Key: "skuName", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", tier))},
			{Key: "meterName", Value: strPtr("Data Stored")},
		}),
		PriceFilter: priceFilterConsumption,
		UsageBased:  true,
	}
}

func mssqlLicenseCostComponent(region string, cores *int64, tier string) *schema.CostComponent {
	licenseRegion := "Global"
	if strings.Contains(region, "usgov") {
		licenseRegion = "US Gov"
	}

	if strings.Contains(region, "china") {
		licenseRegion = "China"
	}

	if strings.Contains(region, "germany") {
		licenseRegion = "Germany"
	}

	coresVal := int64(1)
	if cores != nil {
		coresVal = *cores
	}

	return &schema.CostComponent{
		Name:           "SQL license",
		Unit:           "vCore-hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(coresVal)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr(vendorName),
			Region:        strPtr(licenseRegion),
			Service:       strPtr("SQL Database"),
			ProductFamily: strPtr("Databases"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", ValueRegex: strPtr(fmt.Sprintf("/%s - %s/", tier, "SQL License"))},
			},
		},
		PriceFilter: priceFilterConsumption,
	}
}

func mssqlStorageCostComponent(region string, tier string, zoneRedundant bool, maxSizeGB *float64) *schema.CostComponent {
	storageGB := decimalPtr(decimal.NewFromInt(5))
	if maxSizeGB != nil {
		storageGB = decimalPtr(decimal.NewFromFloat(*maxSizeGB))
	}

	storageTier := tier
	if strings.EqualFold(tier, sqlServerlessTier) {
		storageTier = "General Purpose"
	}

	skuName := storageTier
	if (strings.EqualFold(tier, sqlGeneralPurposeTier) || strings.EqualFold(tier, sqlServerlessTier)) && zoneRedundant {
		skuName += " Zone Redundancy"
	}

	productNameRegex := fmt.Sprintf("/%s - Storage/", storageTier)

	filters := []*schema.AttributeFilter{
		{Key: "productName", ValueRegex: strPtr(productNameRegex)},
		{Key: "skuName", Value: strPtr(skuName)},
		{Key: "meterName", ValueRegex: regexPtr("Data Stored$")},
	}

	if skuName == "Hyperscale" {
		filters = append(filters, &schema.AttributeFilter{Key: "armSkuName", Value: strPtr(skuName)})
	}

	return &schema.CostComponent{
		Name:            "Storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: storageGB,
		ProductFilter:   mssqlProductFilter(region, filters),
	}
}
