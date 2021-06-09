package google

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetComputeTargetGrpcProxyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_compute_target_grpc_proxy",
		RFunc: NewComputeTargetProxy,
		Notes: []string{"Price for additional forwarding rule is used"},
	}
}
func GetComputeTargetHttpProxyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_compute_target_http_proxy",
		RFunc: NewComputeTargetProxy,
		Notes: []string{"Price for additional forwarding rule is used"},
	}
}
func GetComputeTargetHttpsProxyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_compute_target_https_proxy",
		RFunc: NewComputeTargetProxy,
		Notes: []string{"Price for additional forwarding rule is used"},
	}
}
func GetComputeTargetSslProxyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_compute_target_ssl_proxy",
		RFunc: NewComputeTargetProxy,
		Notes: []string{"Price for additional forwarding rule is used"},
	}
}
func GetComputeTargetTcpProxyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_compute_target_tcp_proxy",
		RFunc: NewComputeTargetProxy,
		Notes: []string{"Price for additional forwarding rule is used"},
	}
}
func GetComputeRegionTargetHttpProxyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_compute_region_target_http_proxy",
		RFunc: NewComputeTargetProxy,
		Notes: []string{"Price for additional forwarding rule is used"},
	}
}
func GetComputeRegionTargetHttpsProxyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_compute_region_target_https_proxy",
		RFunc: NewComputeTargetProxy,
		Notes: []string{"Price for additional forwarding rule is used"},
	}
}

func NewComputeTargetProxy(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var monthlyProxyInstances, monthlyDataProcessedGb *decimal.Decimal
	region := d.Get("region").String()
	costComponents := make([]*schema.CostComponent, 0)

	if u != nil && u.Get("monthly_proxy_instances").Type != gjson.Null {
		monthlyProxyInstances = decimalPtr(decimal.NewFromFloat(u.Get("monthly_proxy_instances").Float()))
	}

	costComponents = append(costComponents, proxyInstanceCostComponent(region, monthlyProxyInstances))

	if u != nil && u.Get("monthly_data_processed_gb").Type != gjson.Null {
		monthlyDataProcessedGb = decimalPtr(decimal.NewFromInt(u.Get("monthly_data_processed_gb").Int()))
	}

	costComponents = append(costComponents, dataProcessedCostComponent(region, monthlyDataProcessedGb))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func proxyInstanceCostComponent(region string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Proxy instance",
		Unit:            "hours",
		UnitMultiplier:  1,
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(region),
			Service:       strPtr("Compute Engine"),
			ProductFamily: strPtr("Network"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: strPtr("/^Network Load Balancing: Forwarding Rule Minimum/")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("OnDemand"),
		},
	}
}

func dataProcessedCostComponent(region string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Data processed",
		Unit:            "GB",
		UnitMultiplier:  1,
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(region),
			Service:       strPtr("Compute Engine"),
			ProductFamily: strPtr("Network"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: strPtr("/^Network Internal Load Balancing: Data Processing/")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("OnDemand"),
		},
	}
}
