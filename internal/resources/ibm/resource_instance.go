package ibm

import (
	"fmt"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

// ResourceInstance struct represents a resource instance
//
// This terraform resource is opaque and can handle a number of services, provided with the right parameters
type ResourceInstance struct {
	Name       string
	Address    string
	Service    string
	Plan       string
	Location   string
	Parameters gjson.Result

	// KMS
	// Catalog Link: https://cloud.ibm.com/catalog/services/key-protect
	// Pricing Link: https://cloud.ibm.com/docs/key-protect?topic=key-protect-pricing-plan&interface=ui
	KMS_KeyVersions *int64 `infracost_usage:"kms_key_versions"`
	// Secrets Manager
	// Catalog link: https://cloud.ibm.com/catalog/services/secrets-manager
	SecretsManager_Instance      *int64 `infracost_usage:"secretsmanager_instance"`
	SecretsManager_ActiveSecrets *int64 `infracost_usage:"secretsmanager_active_secrets"`
	// App ID
	// Catalog https://cloud.ibm.com/catalog/services/app-id
	// Pricing https://cloud.ibm.com/docs/appid?topic=appid-pricing
	AppID_Authentications         *int64 `infracost_usage:"appid_authentications"`
	AppID_Users                   *int64 `infracost_usage:"appid_users"`
	AppID_AdvancedAuthentications *int64 `infracost_usage:"appid_advanced_authentications"`
	// App Connect
	// Catalog https://cloud.ibm.com/catalog/services/app-connect
	// Pricing https://www.ibm.com/products/app-connect/pricing
	AppConnect_GigabyteTransmittedOutbounds *float64 `infracost_usage:"appconnect_gigabyte_transmitted_outbounds"`
	AppConnect_ThousandRuns                 *float64 `infracost_usage:"appconnect_thousand_runs"`
	AppConnect_VCPUHours                    *float64 `infracost_usage:"appconnect_vcpu_hours"`
	// LogDNA
	// Catalog https://cloud.ibm.com/catalog/services/logdna
	LogDNA_GigabyteMonths *float64 `infracost_usage:"logdna_gigabyte_months"`
	// Activity Tracker
	// Catalog https://cloud.ibm.com/catalog/services/logdnaat
	ActivityTracker_GigabyteMonths *float64 `infracost_usage:"activitytracker_gigabyte_months"`
	// Monitoring (Sysdig)
	// Catalog https://cloud.ibm.com/catalog/services/ibm-cloud-monitoring
	// Pricing https://cloud.ibm.com/docs/monitoring?topic=monitoring-pricing_plans
	Monitoring_NodeHour       *float64 `infracost_usage:"monitoring_node_hour"`
	Monitoring_ContainerHour  *float64 `infracost_usage:"monitoring_container_hour"`
	Monitoring_APICall        *float64 `infracost_usage:"monitoring_api_call"`
	Monitoring_TimeSeriesHour *float64 `infracost_usage:"monitoring_timeseries_hour"`
	// Continuous Delivery
	// Catalog https://cloud.ibm.com/catalog/services/continuous-delivery
	// Pricing https://cloud.ibm.com/docs/ContinuousDelivery?topic=ContinuousDelivery-limitations_usage&interface=ui
	ContinuousDelivery_AuthorizedUsers *int64 `infracost_usage:"continuousdelivery_authorized_users"`
	// Watson Machine Learning
	// https://dataplatform.cloud.ibm.com/docs/content/wsj/getting-started/wml-plans.html?context=cpdaas
	WML_CUH      *float64 `infracost_usage:"wml_capacity_unit_hour"`
	WML_Instance *float64 `infracost_usage:"wml_instance"`
	WML_Class1RU *float64 `infracost_usage:"wml_class1_ru"`
	WML_Class2RU *float64 `infracost_usage:"wml_class2_ru"`
	WML_Class3RU *float64 `infracost_usage:"wml_class3_ru"`
}

type ResourceCostComponentsFunc func(*ResourceInstance) []*schema.CostComponent

// PopulateUsage parses the u schema.UsageData into the ResourceInstance.
// It uses the `infracost_usage` struct tags to populate data into the ResourceInstance.
func (r *ResourceInstance) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// ResourceInstanceUsageSchema defines a list which represents the usage schema of ResourceInstance.
var ResourceInstanceUsageSchema = []*schema.UsageItem{
	{Key: "kms_key_versions", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "secretsmanager_instance", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "secretsmanager_active_secrets", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "appid_authentications", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "appid_users", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "appid_advanced_authentications", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "appconnect_gigabyte_transmitted_outbounds", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "appconnect_thousand_runs", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "appconnect_vcpu_hours", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "logdna_gigabyte_months", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "activitytracker_gigabyte_months", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "monitoring_node_hour", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "monitoring_node_hour_lite", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "monitoring_container_hour", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "monitoring_api_call", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "monitoring_timeseries_hour", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "continuousdelivery_authorized_users", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "wml_capacity_unit_hour", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "wml_instance", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "wml_class1_ru", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "wml_class2_ru", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "wml_class3_ru", DefaultValue: 0, ValueType: schema.Float64},
}

var ResourceInstanceCostMap map[string]ResourceCostComponentsFunc = map[string]ResourceCostComponentsFunc{
	"kms":                 GetKMSCostComponents,
	"secrets-manager":     GetSecretsManagerCostComponents,
	"appid":               GetAppIDCostComponents,
	"appconnect":          GetAppConnectCostComponents,
	"power-iaas":          GetPowerCostComponents,
	"logdna":              GetLogDNACostComponents,
	"logdnaat":            GetActivityTrackerCostComponents,
	"sysdig-monitor":      GetSysdigCostComponenets,
	"continuous-delivery": GetContinuousDeliveryCostComponenets,
	"pm-20":               GetWMLCostComponents,
}

func KMSKeyVersionsFreeCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	if r.KMS_KeyVersions != nil {
		q = decimalPtr(decimal.NewFromInt(*r.KMS_KeyVersions))
		if q.GreaterThan(decimal.NewFromInt(5)) {
			q = decimalPtr(decimal.NewFromInt(5))
		}
	}
	costComponent := schema.CostComponent{
		Name:            "Key versions free allowance (first 5 Key Versions)",
		Unit:            "Key Versions",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    strPtr("kms"),
		},
	}
	costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
	return &costComponent
}

func KMSKeyVersionCostComponents(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	if r.KMS_KeyVersions != nil {
		q = decimalPtr(decimal.NewFromInt(*r.KMS_KeyVersions))
		if q.LessThanOrEqual(decimal.NewFromInt(5)) {
			q = decimalPtr(decimal.NewFromInt(0))
		} else {
			q = decimalPtr(q.Sub(decimal.NewFromInt(5)))
		}
	}
	costComponent := schema.CostComponent{
		Name:            "Key versions",
		Unit:            "Key Versions",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
	}
	return &costComponent
}

func GetKMSCostComponents(r *ResourceInstance) []*schema.CostComponent {
	return []*schema.CostComponent{
		KMSKeyVersionsFreeCostComponent(r),
		KMSKeyVersionCostComponents(r),
	}
}

func SecretsManagerInstanceCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	if r.SecretsManager_Instance != nil {
		q = decimalPtr(decimal.NewFromInt(*r.SecretsManager_Instance))
	}
	costComponent := schema.CostComponent{
		Name:            "Instance",
		Unit:            "Instance",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("INSTANCES"),
		},
	}
	return &costComponent
}

func SecretsManagerActiveSecretsCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	if r.SecretsManager_ActiveSecrets != nil {
		q = decimalPtr(decimal.NewFromInt(*r.SecretsManager_ActiveSecrets))
	}
	costComponent := schema.CostComponent{
		Name:            "Active Secrets",
		Unit:            "Secrets",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("ACTIVE_SECRETS"),
		},
	}
	return &costComponent
}

func GetSecretsManagerCostComponents(r *ResourceInstance) []*schema.CostComponent {
	if r.Plan == "standard" {
		return []*schema.CostComponent{
			SecretsManagerInstanceCostComponent(r),
			SecretsManagerActiveSecretsCostComponent(r),
		}
	} else {
		costComponent := schema.CostComponent{
			Name:            fmt.Sprintf("Plan: %s", r.Plan),
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		}
		costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
		return []*schema.CostComponent{
			&costComponent,
		}
	}
}

func GetPowerCostComponents(r *ResourceInstance) []*schema.CostComponent {
	q := decimalPtr(decimal.NewFromInt(1))

	costComponent := schema.CostComponent{
		Name:            r.Name,
		Unit:            "Instance",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
	}
	costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
	return []*schema.CostComponent{
		&costComponent,
	}
}

func AppIDUserCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	if r.AppID_Users != nil {
		q = decimalPtr(decimal.NewFromInt(*r.AppID_Users))
	}
	costComponent := schema.CostComponent{
		Name:            "Users",
		Unit:            "Users",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("USERS_PER_MONTH"),
		},
	}
	return &costComponent
}

func AppIDAuthenticationCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	if r.AppID_Authentications != nil {
		q = decimalPtr(decimal.NewFromInt(*r.AppID_Authentications))
	}
	costComponent := schema.CostComponent{
		Name:            "Authentications",
		Unit:            "Authentications",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("AUTHENTICATIONS_PER_MONTH"),
		},
	}
	return &costComponent
}

func AppIDAdvancedAuthenticationCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	if r.AppID_AdvancedAuthentications != nil {
		q = decimalPtr(decimal.NewFromInt(*r.AppID_AdvancedAuthentications))
	}
	costComponent := schema.CostComponent{
		Name:            "Advanced Authentications",
		Unit:            "Authentications",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("ADVANCED_AUTHENTICATIONS_PER_MONTH"),
		},
	}
	return &costComponent
}

func GetAppIDCostComponents(r *ResourceInstance) []*schema.CostComponent {
	if r.Plan == "graduated-tier" {
		return []*schema.CostComponent{
			AppIDUserCostComponent(r),
			AppIDAuthenticationCostComponent(r),
			AppIDAdvancedAuthenticationCostComponent(r),
		}
	} else {
		costComponent := schema.CostComponent{
			Name:            fmt.Sprintf("Plan: %s", r.Plan),
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		}
		costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
		return []*schema.CostComponent{
			&costComponent,
		}
	}
}

func AppConnectFlowsRunsCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	if r.AppConnect_ThousandRuns != nil {
		q = decimalPtr(decimal.NewFromFloat(*r.AppConnect_ThousandRuns))
	}
	return &schema.CostComponent{
		Name:            "Flow runs",
		Unit:            "1k runs",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("ITEMS_PER_MONTH"),
		},
	}
}

func AppConnectEgressCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	if r.AppConnect_GigabyteTransmittedOutbounds != nil {
		q = decimalPtr(decimal.NewFromFloat(*r.AppConnect_GigabyteTransmittedOutbounds))
	}
	return &schema.CostComponent{
		Name:            "Gigabytes transmitted outbounds",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("GIGABYTE_TRANSMITTED_OUTBOUND"),
		},
	}
}

func AppConnectCpuCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	if r.AppConnect_VCPUHours != nil {
		q = decimalPtr(decimal.NewFromFloat(*r.AppConnect_VCPUHours))
	}
	return &schema.CostComponent{
		Name:            "VCPU",
		Unit:            "VCPU hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("VIRTUAL_PROCESSOR_CORE_HOURS"),
		},
	}
}

func GetAppConnectCostComponents(r *ResourceInstance) []*schema.CostComponent {
	if r.Plan == "appconnectplanprofessional" {
		return []*schema.CostComponent{
			AppConnectFlowsRunsCostComponent(r),
			AppConnectEgressCostComponent(r),
		}
	} else if r.Plan == "appconnectplanenterprise" {
		return []*schema.CostComponent{
			AppConnectFlowsRunsCostComponent(r),
			AppConnectEgressCostComponent(r),
			AppConnectCpuCostComponent(r),
		}
	} else if r.Plan == "lite" {
		costComponent := schema.CostComponent{
			Name:            "Lite plan",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		}
		costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
		return []*schema.CostComponent{
			&costComponent,
		}
	} else {
		costComponent := schema.CostComponent{
			Name:            fmt.Sprintf("Plan %s with customized pricing", r.Plan),
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		}
		costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
		return []*schema.CostComponent{
			&costComponent,
		}
	}
}

func GetLogDNACostComponents(r *ResourceInstance) []*schema.CostComponent {
	var q *decimal.Decimal
	if r.LogDNA_GigabyteMonths != nil {
		q = decimalPtr(decimal.NewFromFloat(*r.LogDNA_GigabyteMonths))
	}
	if r.Plan == "lite" {
		costComponent := schema.CostComponent{
			Name:            "Lite plan",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		}
		costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
		return []*schema.CostComponent{
			&costComponent,
		}
	} else {
		return []*schema.CostComponent{{
			Name:            "Gigabyte Months",
			Unit:            "Gigabyte Months",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: q,
			ProductFilter: &schema.ProductFilter{
				VendorName: strPtr("ibm"),
				Region:     strPtr(r.Location),
				Service:    &r.Service,
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "planName", Value: &r.Plan},
				},
			},
			PriceFilter: &schema.PriceFilter{
				Unit: strPtr("GIGABYTE_MONTHS"),
			},
		}}
	}
}

func GetSysdigTimeseriesCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	if r.Monitoring_TimeSeriesHour != nil {
		q = decimalPtr(decimal.NewFromFloat(*r.Monitoring_TimeSeriesHour))
	}
	return &schema.CostComponent{
		Name:            "Additional Time series",
		Unit:            "Time series hour",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("TIME_SERIES_HOURS"),
		},
	}
}

func GetSysdigContainerCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	if r.Monitoring_ContainerHour != nil {
		q = decimalPtr(decimal.NewFromFloat(*r.Monitoring_ContainerHour))
	}
	return &schema.CostComponent{
		Name:            "Additional Containers",
		Unit:            "Container Hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("CONTAINER_HOURS"),
		},
	}
}

func GetSysdigApiCallCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	if r.Monitoring_APICall != nil {
		q = decimalPtr(decimal.NewFromFloat(*r.Monitoring_APICall))
	}
	return &schema.CostComponent{
		Name:            "Additional API Calls",
		Unit:            "API Calls",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("API_CALL_HOURS"),
		},
	}
}

func GetSysdigNodeHourCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	if r.Monitoring_NodeHour != nil {
		q = decimalPtr(decimal.NewFromFloat(*r.Monitoring_NodeHour))
	}
	return &schema.CostComponent{
		Name:            "Base Node Hour",
		Unit:            "Node Hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("NODE_HOURS"),
		},
	}
}

func GetSysdigCostComponenets(r *ResourceInstance) []*schema.CostComponent {

	if r.Plan == "lite" {
		costComponent := &schema.CostComponent{
			Name:            "Lite plan",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		}
		costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
		return []*schema.CostComponent{costComponent}
	} else {
		return []*schema.CostComponent{
			GetSysdigTimeseriesCostComponent(r),
			GetSysdigContainerCostComponent(r),
			GetSysdigApiCallCostComponent(r),
			GetSysdigNodeHourCostComponent(r),
		}
	}
}

func GetActivityTrackerCostComponents(r *ResourceInstance) []*schema.CostComponent {
	var q *decimal.Decimal
	if r.ActivityTracker_GigabyteMonths != nil {
		q = decimalPtr(decimal.NewFromFloat(*r.ActivityTracker_GigabyteMonths))
	}
	if r.Plan == "lite" {
		costComponent := &schema.CostComponent{
			Name:            "Lite plan",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		}
		costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
		return []*schema.CostComponent{costComponent}
	} else {
		return []*schema.CostComponent{{
			Name:            "Gigabyte Months",
			Unit:            "Gigabyte Months",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: q,
			ProductFilter: &schema.ProductFilter{
				VendorName: strPtr("ibm"),
				Region:     strPtr(r.Location),
				Service:    &r.Service,
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "planName", Value: &r.Plan},
				},
			},
			PriceFilter: &schema.PriceFilter{
				Unit: strPtr("GIGABYTE_MONTHS"),
			},
		}}
	}
}

func GetContinuousDeliveryCostComponenets(r *ResourceInstance) []*schema.CostComponent {
	var q *decimal.Decimal
	if r.ContinuousDelivery_AuthorizedUsers != nil {
		q = decimalPtr(decimal.NewFromInt(*r.ContinuousDelivery_AuthorizedUsers))
	}
	if r.Plan == "lite" {
		costComponent := &schema.CostComponent{
			Name:            "Lite plan",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		}
		costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
		return []*schema.CostComponent{costComponent}
	} else {
		return []*schema.CostComponent{{
			Name:            "Authorized Users",
			Unit:            "Authorized Users",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: q,
			ProductFilter: &schema.ProductFilter{
				VendorName: strPtr("ibm"),
				Region:     strPtr(r.Location),
				Service:    &r.Service,
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "planName", Value: &r.Plan},
				},
			},
			PriceFilter: &schema.PriceFilter{
				Unit: strPtr("AUTHORIZED_USERS_PER_MONTH"),
			},
		}}
	}
}

// BuildResource builds a schema.Resource from a valid ResourceInstance struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *ResourceInstance) BuildResource() *schema.Resource {
	costComponentsFunc, ok := ResourceInstanceCostMap[r.Service]

	if !ok {
		return &schema.Resource{
			Name:        r.Address,
			UsageSchema: ResourceInstanceUsageSchema,
		}
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    ResourceInstanceUsageSchema,
		CostComponents: costComponentsFunc(r),
	}
}
