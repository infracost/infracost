package azure

import (
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// SecurityCenterSubscriptionPricing struct represents the pricing structure for Microsoft Defender for Cloud.
// Currently, pricing is supported through the usage file.
//
// Resource information: https://learn.microsoft.com/en-us/azure/defender-for-cloud/
// Pricing information: https://azure.microsoft.com/en-us/pricing/details/defender-for-cloud/
type SecurityCenterSubscriptionPricing struct {
	Address      string
	Region       string
	Tier         string
	ResourceType string

	MonthlyServersPlan1Nodes *float64 `infracost_usage:"monthly_servers_plan_1_nodes"`
	MonthlyServersPlan2Nodes *float64 `infracost_usage:"monthly_servers_plan_2_nodes"`

	MonthlyContainersVCores        *float64 `infracost_usage:"monthly_containers_vcores"`
	MonthlyContainerRegistryImages *float64 `infracost_usage:"monthly_container_registry_images"`

	MonthlySQLAzureConnectedInstances *float64 `infracost_usage:"monthly_sql_azure_connected_instances"`
	MonthlySQLOutsideAzureVCores      *float64 `infracost_usage:"monthly_sql_outside_azure_vcores"`
	MonthlyMySQLInstances             *float64 `infracost_usage:"monthly_mysql_instances"`
	MonthlyPostgreSQLInstances        *float64 `infracost_usage:"monthly_postgresql_instances"`
	MonthlyMariaDBInstances           *float64 `infracost_usage:"monthly_mariadb_instances"`
	CosmosDBRequestUnits              *float64 `infracost_usage:"cosmosdb_request_units"`

	MonthlyStorageAccounts *float64 `infracost_usage:"monthly_storage_accounts"`

	MonthlyAppServiceNodes  *float64 `infracost_usage:"monthly_app_service_nodes"`
	MonthlyKeyVaults        *int64   `infracost_usage:"monthly_key_vaults"`
	MonthlyARMSubscriptions *int64   `infracost_usage:"monthly_arm_subscriptions"`
	MonthlyDNSQueries       *int64   `infracost_usage:"monthly_dns_queries"`

	MonthlyKubernetesCores *float64 `infracost_usage:"monthly_kubernetes_cores"`
}

// CoreType returns the name of this resource type
func (r *SecurityCenterSubscriptionPricing) CoreType() string {
	return "SecurityCenterSubscriptionPricing"
}

// UsageSchema defines a list which represents the usage schema of SecurityCenterSubscriptionPricing.
func (r *SecurityCenterSubscriptionPricing) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_servers_plan_1_nodes", DefaultValue: 0.0, ValueType: schema.Float64},
		{Key: "monthly_servers_plan_2_nodes", DefaultValue: 0.0, ValueType: schema.Float64},
		{Key: "monthly_containers_vcores", DefaultValue: 0.0, ValueType: schema.Float64},
		{Key: "monthly_container_registry_images", DefaultValue: 0.0, ValueType: schema.Float64},
		{Key: "monthly_sql_azure_connected_instances", DefaultValue: 0.0, ValueType: schema.Float64},
		{Key: "monthly_sql_outside_azure_vcores", DefaultValue: 0.0, ValueType: schema.Float64},
		{Key: "monthly_mysql_instances", DefaultValue: 0.0, ValueType: schema.Float64},
		{Key: "monthly_postgresql_instances", DefaultValue: 0.0, ValueType: schema.Float64},
		{Key: "monthly_mariadb_instances", DefaultValue: 0.0, ValueType: schema.Float64},
		{Key: "cosmosdb_request_units", DefaultValue: 0.0, ValueType: schema.Float64},
		{Key: "monthly_storage_accounts", DefaultValue: 0.0, ValueType: schema.Float64},
		{Key: "monthly_app_service_nodes", DefaultValue: 0.0, ValueType: schema.Float64},
		{Key: "monthly_key_vaults", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_arm_subscriptions", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_dns_queries", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_kubernetes_cores", DefaultValue: 0.0, ValueType: schema.Float64},
	}
}

// PopulateUsage parses the u schema.UsageData into the SecurityCenterSubscriptionPricing.
// It uses the `infracost_usage` struct tags to populate data into the SecurityCenterSubscriptionPricing.
func (r *SecurityCenterSubscriptionPricing) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid SecurityCenterSubscriptionPricing struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *SecurityCenterSubscriptionPricing) BuildResource() *schema.Resource {
	if strings.ToLower(r.Tier) == "free" {
		return &schema.Resource{
			Name:      r.Address,
			IsSkipped: true,
			NoPrice:   true,
		}
	}

	var costComponents []*schema.CostComponent
	switch strings.ToLower(r.ResourceType) {
	case "appservices":
		costComponents = []*schema.CostComponent{r.addAppServiceCostComponent()}
	case "containerregistry":
		costComponents = []*schema.CostComponent{r.addContainerRegistryCostComponent()}
	case "keyvaults":
		costComponents = []*schema.CostComponent{r.addKeyVaultCostComponent()}
	case "kubernetesservice":
		costComponents = []*schema.CostComponent{r.addKubernetesCostComponent()}
	case "sqlservers":
		costComponents = []*schema.CostComponent{r.addSQLOutsideAzureCostComponent()}
	case "sqlservervirtualmachines":
		costComponents = []*schema.CostComponent{r.addSQLAzureConnectedCostComponent()}
	case "storageaccounts":
		costComponents = []*schema.CostComponent{r.addStorageCostComponent()}
	case "virtualmachines":
		costComponents = []*schema.CostComponent{
			r.addServersP1CostComponent(),
			r.addServersP2CostComponent(),
		}
	case "arm":
		costComponents = []*schema.CostComponent{r.addARMCostComponent()}
	case "dns":
		costComponents = []*schema.CostComponent{r.addDNSCostComponent()}
	case "opensourcerelationaldatabases":
		costComponents = []*schema.CostComponent{
			r.addMySQLCostComponent(),
			r.addPostgreSQLCostComponent(),
			r.addMariaDBCostComponent(),
		}
	case "containers":
		costComponents = []*schema.CostComponent{r.addContainersCostComponent()}
	case "cosmosdbs":
		costComponents = []*schema.CostComponent{r.addCosmosDBCostComponent()}
	default:
		logging.Logger.Warn().Msgf("Skipping resource %s. Unknown resource type  '%s'", r.Address, r.ResourceType)
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *SecurityCenterSubscriptionPricing) addServersP1CostComponent() *schema.CostComponent {
	var vmHours *decimal.Decimal
	if r.MonthlyServersPlan1Nodes != nil {
		vmHours = decimalPtr(decimal.NewFromFloat(*r.MonthlyServersPlan1Nodes).Mul(schema.HourToMonthUnitMultiplier))
	}

	return &schema.CostComponent{
		Name:            "Defender for servers, plan 1",
		Unit:            "server",
		UnitMultiplier:  schema.HourToMonthUnitMultiplier,
		MonthlyQuantity: vmHours,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        r.normalizedRegion(),
			ProductFamily: strPtr("Security"),
			Service:       strPtr("Microsoft Defender for Cloud"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Microsoft Defender for Servers")},
				{Key: "meterName", Value: strPtr("Standard P1 Node")},
			},
		},
		UsageBased: true,
	}
}

func (r *SecurityCenterSubscriptionPricing) addServersP2CostComponent() *schema.CostComponent {
	var vmHours *decimal.Decimal
	if r.MonthlyServersPlan2Nodes != nil {
		vmHours = decimalPtr(decimal.NewFromFloat(*r.MonthlyServersPlan2Nodes).Mul(schema.HourToMonthUnitMultiplier))
	}

	return &schema.CostComponent{
		Name:            "Defender for servers, plan 2",
		Unit:            "server",
		UnitMultiplier:  schema.HourToMonthUnitMultiplier,
		MonthlyQuantity: vmHours,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        r.normalizedRegion(),
			ProductFamily: strPtr("Security"),
			Service:       strPtr("Microsoft Defender for Cloud"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Microsoft Defender for Servers")},
				{Key: "meterName", Value: strPtr("Standard P2 Node")},
			},
		},
		UsageBased: true,
	}
}

func (r *SecurityCenterSubscriptionPricing) addContainersCostComponent() *schema.CostComponent {
	var vmHours *decimal.Decimal
	if r.MonthlyContainersVCores != nil {
		vmHours = decimalPtr(decimal.NewFromFloat(*r.MonthlyContainersVCores).Mul(schema.HourToMonthUnitMultiplier))
	}

	return &schema.CostComponent{
		Name:            "Defender for containers",
		Unit:            "vCore",
		UnitMultiplier:  schema.HourToMonthUnitMultiplier,
		MonthlyQuantity: vmHours,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        r.normalizedRegion(),
			ProductFamily: strPtr("Security"),
			Service:       strPtr("Microsoft Defender for Cloud"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Microsoft Defender for Containers")},
				{Key: "meterName", Value: strPtr("Standard vCore vCore Pack")},
			},
		},
		UsageBased: true,
	}
}

func (r *SecurityCenterSubscriptionPricing) addSQLAzureConnectedCostComponent() *schema.CostComponent {
	var instances *decimal.Decimal
	if r.MonthlySQLAzureConnectedInstances != nil {
		instances = decimalPtr(decimal.NewFromFloat(*r.MonthlySQLAzureConnectedInstances))
	}

	return &schema.CostComponent{
		Name:            "Defender for SQL, Azure-connected",
		Unit:            "instance",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: instances,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        r.normalizedRegion(),
			ProductFamily: strPtr("Security"),
			Service:       strPtr("Microsoft Defender for Cloud"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Microsoft Defender for SQL")},
				{Key: "meterName", Value: strPtr("Standard Node")},
			},
		},
		UsageBased: true,
	}
}

func (r *SecurityCenterSubscriptionPricing) addSQLOutsideAzureCostComponent() *schema.CostComponent {
	var vCoreHours *decimal.Decimal
	if r.MonthlySQLOutsideAzureVCores != nil {
		vCoreHours = decimalPtr(decimal.NewFromFloat(*r.MonthlySQLOutsideAzureVCores).Mul(schema.HourToMonthUnitMultiplier))
	}

	return &schema.CostComponent{
		Name:            "Defender for SQL, outside Azure",
		Unit:            "vCore",
		UnitMultiplier:  schema.HourToMonthUnitMultiplier,
		MonthlyQuantity: vCoreHours,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        r.normalizedRegion(),
			ProductFamily: strPtr("Security"),
			Service:       strPtr("Microsoft Defender for Cloud"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Microsoft Defender for SQL")},
				{Key: "meterName", Value: strPtr("Standard vCore")},
			},
		},
		UsageBased: true,
	}
}

func (r *SecurityCenterSubscriptionPricing) addMySQLCostComponent() *schema.CostComponent {
	var instances *decimal.Decimal
	if r.MonthlyMySQLInstances != nil {
		instances = decimalPtr(decimal.NewFromFloat(*r.MonthlyMySQLInstances))
	}

	return &schema.CostComponent{
		Name:            "Defender for MySQL",
		Unit:            "instance",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: instances,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        r.normalizedRegion(),
			ProductFamily: strPtr("Security"),
			Service:       strPtr("Microsoft Defender for Cloud"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Microsoft Defender for MySQL")},
				{Key: "meterName", Value: strPtr("Standard Node")},
			},
		},
		UsageBased: true,
	}
}

func (r *SecurityCenterSubscriptionPricing) addPostgreSQLCostComponent() *schema.CostComponent {
	var instances *decimal.Decimal
	if r.MonthlyPostgreSQLInstances != nil {
		instances = decimalPtr(decimal.NewFromFloat(*r.MonthlyPostgreSQLInstances))
	}

	return &schema.CostComponent{
		Name:            "Defender for PostgreSQL",
		Unit:            "instance",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: instances,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        r.normalizedRegion(),
			ProductFamily: strPtr("Security"),
			Service:       strPtr("Microsoft Defender for Cloud"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Microsoft Defender for PostgreSQL")},
				{Key: "meterName", Value: strPtr("Standard Node")},
			},
		},
		UsageBased: true,
	}
}

func (r *SecurityCenterSubscriptionPricing) addMariaDBCostComponent() *schema.CostComponent {
	var instances *decimal.Decimal
	if r.MonthlyMariaDBInstances != nil {
		instances = decimalPtr(decimal.NewFromFloat(*r.MonthlyMariaDBInstances))
	}

	region := r.normalizedRegion()
	if *region == "Global" {
		// force to west-us2 since price is not available in Global
		region = strPtr("westus2")
	}

	return &schema.CostComponent{
		Name:           "Defender for MariaDB",
		Unit:           "instance",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: instances,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        region,
			ProductFamily: strPtr("Security"),
			Service:       strPtr("Microsoft Defender for Cloud"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Microsoft Defender for MariaDB")},
				{Key: "meterName", Value: strPtr("Standard Instance")},
			},
		},
		UsageBased: true,
	}
}

func (r *SecurityCenterSubscriptionPricing) addCosmosDBCostComponent() *schema.CostComponent {
	var averageRUs *decimal.Decimal
	if r.CosmosDBRequestUnits != nil {
		averageRUs = decimalPtr(decimal.NewFromFloat(*r.CosmosDBRequestUnits).Div(decimal.NewFromInt(100)))
	}

	return &schema.CostComponent{
		Name:           "Defender for Cosmos DB",
		Unit:           "RU/s x 100",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: averageRUs,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        r.normalizedRegion(),
			ProductFamily: strPtr("Security"),
			Service:       strPtr("Microsoft Defender for Cloud"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Microsoft Defender for Azure Cosmos DB")},
				{Key: "meterName", Value: strPtr("Standard 100 RU/s")},
			},
		},
		UsageBased: true,
	}
}

func (r *SecurityCenterSubscriptionPricing) addStorageCostComponent() *schema.CostComponent {
	var storageAccounts *decimal.Decimal
	if r.MonthlyStorageAccounts != nil {
		storageAccounts = decimalPtr(decimal.NewFromFloat(*r.MonthlyStorageAccounts))
	}

	return &schema.CostComponent{
		Name:           "Defender for storage",
		Unit:           "storage account",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: storageAccounts,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        r.normalizedRegion(),
			ProductFamily: strPtr("Security"),
			Service:       strPtr("Microsoft Defender for Cloud"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Microsoft Defender for Storage")},
				{Key: "meterName", Value: strPtr("Standard Node")},
			},
		},
		UsageBased: true,
	}
}

func (r *SecurityCenterSubscriptionPricing) addAppServiceCostComponent() *schema.CostComponent {
	var nodes *decimal.Decimal
	if r.MonthlyAppServiceNodes != nil {
		nodes = decimalPtr(decimal.NewFromFloat(*r.MonthlyAppServiceNodes))
	}

	return &schema.CostComponent{
		Name:           "Defender for app service",
		Unit:           "node",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: nodes,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        r.normalizedRegion(),
			ProductFamily: strPtr("Security"),
			Service:       strPtr("Microsoft Defender for Cloud"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Microsoft Defender for App Service")},
				{Key: "meterName", Value: strPtr("Standard Node")},
			},
		},
		UsageBased: true,
	}
}

func (r *SecurityCenterSubscriptionPricing) addKeyVaultCostComponent() *schema.CostComponent {
	var keyVaults *decimal.Decimal
	if r.MonthlyKeyVaults != nil {
		keyVaults = decimalPtr(decimal.NewFromInt(*r.MonthlyKeyVaults))
	}

	region := r.normalizedRegion()
	if *region == "Global" {
		// force to west-us2 since price is not available in Global
		region = strPtr("westus2")
	}

	return &schema.CostComponent{
		Name:           "Defender for Key Vault",
		Unit:           "key vault",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: keyVaults,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        region,
			ProductFamily: strPtr("Security"),
			Service:       strPtr("Microsoft Defender for Cloud"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Microsoft Defender for Key Vault")},
				{Key: "meterName", Value: strPtr("Per node Std Node")},
			},
		},
		UsageBased: true,
	}
}

func (r *SecurityCenterSubscriptionPricing) addARMCostComponent() *schema.CostComponent {
	var subscriptions *decimal.Decimal
	if r.MonthlyARMSubscriptions != nil {
		subscriptions = decimalPtr(decimal.NewFromInt(*r.MonthlyARMSubscriptions))
	}

	region := r.normalizedRegion()
	if *region == "Global" {
		// force to west-us2 since price is not available in Global
		region = strPtr("westus2")
	}

	return &schema.CostComponent{
		Name:           "Defender for ARM",
		Unit:           "subscription",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: subscriptions,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        region,
			ProductFamily: strPtr("Security"),
			Service:       strPtr("Microsoft Defender for Cloud"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Microsoft Defender for Resource Manager")},
				{Key: "meterName", Value: strPtr("Per node Std Node")},
			},
		},
		UsageBased: true,
	}
}

func (r *SecurityCenterSubscriptionPricing) addDNSCostComponent() *schema.CostComponent {
	var apiCalls *decimal.Decimal
	if r.MonthlyDNSQueries != nil {
		apiCalls = decimalPtr(decimal.NewFromInt(*r.MonthlyDNSQueries).Div(decimal.NewFromInt(1000000)))
	}

	region := r.normalizedRegion()
	if *region == "Global" {
		// force to west-us2 since price is not available in Global
		region = strPtr("westus2")
	}

	return &schema.CostComponent{
		Name:            "Defender for DNS",
		Unit:            "1M queries",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: apiCalls,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        region,
			ProductFamily: strPtr("Security"),
			Service:       strPtr("Microsoft Defender for Cloud"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Microsoft Defender for DNS")},
				{Key: "meterName", Value: strPtr("Standard Queries")},
			},
		},
		UsageBased: true,
	}
}

func (r *SecurityCenterSubscriptionPricing) addKubernetesCostComponent() *schema.CostComponent {
	var nodes *decimal.Decimal
	if r.MonthlyKubernetesCores != nil {
		nodes = decimalPtr(decimal.NewFromFloat(*r.MonthlyKubernetesCores))
	}

	return &schema.CostComponent{
		Name:           "Defender for kubernetes",
		Unit:           "core",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: nodes,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        r.normalizedRegion(),
			ProductFamily: strPtr("Security"),
			Service:       strPtr("Advanced Threat Protection"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Microsoft Defender for Kubernetes")},
				{Key: "meterName", Value: strPtr("Standard Cores")},
			},
		},
		UsageBased: true,
	}
}

func (r *SecurityCenterSubscriptionPricing) addContainerRegistryCostComponent() *schema.CostComponent {
	var instances *decimal.Decimal
	if r.MonthlyContainerRegistryImages != nil {
		instances = decimalPtr(decimal.NewFromFloat(*r.MonthlyContainerRegistryImages))
	}
	return &schema.CostComponent{
		Name:            "Defender for container registries",
		Unit:            "image",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: instances,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        r.normalizedRegion(),
			ProductFamily: strPtr("Security"),
			Service:       strPtr("Advanced Threat Protection"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Microsoft Defender for Container Registries")},
				{Key: "meterName", Value: strPtr("Standard Images")},
			},
		},
		UsageBased: true,
	}
}

func (r *SecurityCenterSubscriptionPricing) normalizedRegion() *string {
	if r.Region == "global" {
		return strPtr("Global")
	}
	return strPtr(r.Region)
}
