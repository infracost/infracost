package azure

import (
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// ContainerApp struct represents an Azure Container Apps resource.
//
// Resource information: https://azure.microsoft.com/en-us/services/container-apps/
// Pricing information: https://azure.microsoft.com/en-us/pricing/details/container-apps/
type ContainerApp struct {
	Address             string
	Region              string
	WorkloadProfileName string
	TotalvCPU           float64
	TotalMemory         float64
	MinReplicas         int64

	Requests           *int64   `infracost_usage:"requests"`
	ConcurrentRequests *int64   `infracost_usage:"concurrent_requests"`
	ExecutionTimeMS    *float64 `infracost_usage:"execution_time_ms"`
	UsageMinReplicas   *int64   `infracost_usage:"min_replicas"`
}

// CoreType returns the name of this resource type
func (r *ContainerApp) CoreType() string {
	return "ContainerApp"
}

// UsageSchema defines a list which represents the usage schema of ContainerApp.
func (r *ContainerApp) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "requests", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "concurrent_requests", DefaultValue: 1, ValueType: schema.Int64},
		{Key: "execution_time_ms", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "min_replicas", DefaultValue: 0, ValueType: schema.Int64},
	}
}

// PopulateUsage parses the u schema.UsageData into the ContainerApp.
// It uses the `infracost_usage` struct tags to populate data into the ContainerApp.
func (r *ContainerApp) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid ContainerApp struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *ContainerApp) BuildResource() *schema.Resource {
	var activeVCPU, idleVCPU, activeMem, idleMem *decimal.Decimal
	var freeVCPUUsed, freeMemUsed, freeRequestsQty *decimal.Decimal
	var billableRequests *decimal.Decimal

	requests := int64(0)
	if r.Requests != nil {
		requests = *r.Requests
	}

	concurrentRequests := int64(1)
	if r.ConcurrentRequests != nil && *r.ConcurrentRequests > 0 {
		concurrentRequests = *r.ConcurrentRequests
	}

	executionTimeMS := 0.0
	if r.ExecutionTimeMS != nil {
		executionTimeMS = *r.ExecutionTimeMS
	}

	minReplicas := r.MinReplicas
	if r.UsageMinReplicas != nil {
		minReplicas = *r.UsageMinReplicas
	}

	// Calculate Active and Idle Seconds based on Azure Calculator logic
	// Active Seconds = (Requests * (Execution Time (ms) / 1000)) / Concurrent Requests
	activeSeconds := decimal.NewFromInt(requests).Mul(decimal.NewFromFloat(executionTimeMS).Div(decimal.NewFromInt(1000))).Div(decimal.NewFromInt(concurrentRequests))

	// Idle Seconds = Min Replicas * 730 hours * 3600 seconds
	secondsInMonth := decimal.NewFromInt(730 * 3600)
	idleSeconds := decimal.NewFromInt(minReplicas).Mul(secondsInMonth)

	if !activeSeconds.IsZero() || !idleSeconds.IsZero() {
		totalVCPU := decimal.NewFromFloat(r.TotalvCPU)
		totalMem := decimal.NewFromFloat(r.TotalMemory)

		// Active
		activeVCPUUsage := totalVCPU.Mul(activeSeconds)
		activeMemUsage := totalMem.Mul(activeSeconds)

		// Idle
		idleVCPUUsage := totalVCPU.Mul(idleSeconds)
		idleMemUsage := totalMem.Mul(idleSeconds)

		// Free tiers totals
		freeVCPUInit := decimal.NewFromInt(180000)
		freeMemInit := decimal.NewFromInt(360000)

		remainingFreeVCPU := freeVCPUInit
		remainingFreeMem := freeMemInit

		// Apply free tier to Active first
		billableActiveVCPU := activeVCPUUsage.Sub(remainingFreeVCPU)
		if billableActiveVCPU.IsNegative() {
			remainingFreeVCPU = billableActiveVCPU.Abs()
			billableActiveVCPU = decimal.Zero
		} else {
			remainingFreeVCPU = decimal.Zero
		}

		billableActiveMem := activeMemUsage.Sub(remainingFreeMem)
		if billableActiveMem.IsNegative() {
			remainingFreeMem = billableActiveMem.Abs()
			billableActiveMem = decimal.Zero
		} else {
			remainingFreeMem = decimal.Zero
		}

		// Apply remaining free tier to Idle
		billableIdleVCPU := idleVCPUUsage.Sub(remainingFreeVCPU)
		if billableIdleVCPU.IsNegative() {
			remainingFreeVCPU = billableIdleVCPU.Abs()
			billableIdleVCPU = decimal.Zero
		} else {
			remainingFreeVCPU = decimal.Zero
		}

		billableIdleMem := idleMemUsage.Sub(remainingFreeMem)
		if billableIdleMem.IsNegative() {
			remainingFreeMem = billableIdleMem.Abs()
			billableIdleMem = decimal.Zero
		} else {
			remainingFreeMem = decimal.Zero
		}

		// Calculate used free
		usedVCPU := freeVCPUInit.Sub(remainingFreeVCPU)
		usedMem := freeMemInit.Sub(remainingFreeMem)

		freeVCPUUsed = &usedVCPU
		freeMemUsed = &usedMem

		activeVCPU = &billableActiveVCPU
		idleVCPU = &billableIdleVCPU
		activeMem = &billableActiveMem
		idleMem = &billableIdleMem
	}

	if requests > 0 {
		freeRequestsLimit := int64(2000000)

		var freeReq int64
		var billReq int64

		if requests <= freeRequestsLimit {
			freeReq = requests
			billReq = 0
		} else {
			freeReq = freeRequestsLimit
			billReq = requests - freeRequestsLimit
		}

		f := decimal.NewFromInt(freeReq).Div(decimal.NewFromInt(1000000))
		b := decimal.NewFromInt(billReq).Div(decimal.NewFromInt(1000000))

		freeRequestsQty = &f
		billableRequests = &b
	}

	costComponents := make([]*schema.CostComponent, 0)
	components := []*schema.CostComponent{
		r.vCPUCostComponent(r.WorkloadProfileName, "active", activeVCPU),
		r.memoryCostComponent(r.WorkloadProfileName, "active", activeMem),
		r.vCPUCostComponent(r.WorkloadProfileName, "idle", idleVCPU),
		r.memoryCostComponent(r.WorkloadProfileName, "idle", idleMem),
		r.requestsCostComponent(r.WorkloadProfileName, billableRequests),
	}
	for _, c := range components {
		if c != nil {
			costComponents = append(costComponents, c)
		}
	}

	if freeVCPUUsed != nil && freeVCPUUsed.GreaterThan(decimal.Zero) {
		c := &schema.CostComponent{
			Name:            "vCPU (free)",
			Unit:            "vCPU-seconds",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: freeVCPUUsed,
		}
		c.SetCustomPrice(&decimal.Zero)
		costComponents = append(costComponents, c)
	}

	if freeMemUsed != nil && freeMemUsed.GreaterThan(decimal.Zero) {
		c := &schema.CostComponent{
			Name:            "Memory (free)",
			Unit:            "GiB-seconds",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: freeMemUsed,
		}
		c.SetCustomPrice(&decimal.Zero)
		costComponents = append(costComponents, c)
	}

	if freeRequestsQty != nil && freeRequestsQty.GreaterThan(decimal.Zero) {
		c := &schema.CostComponent{
			Name:            "Requests (free)",
			Unit:            "1M requests",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: freeRequestsQty,
		}
		c.SetCustomPrice(&decimal.Zero)
		costComponents = append(costComponents, c)
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *ContainerApp) vCPUCostComponent(profile string, usageType string, quantity *decimal.Decimal) *schema.CostComponent {
	if profile != "Consumption" && profile != "" {
		return nil // Workload profiles not supported yet in MVP
	}

	if quantity != nil && quantity.LessThanOrEqual(decimal.Zero) {
		return nil
	}

	name := "vCPU (active)"
	meterName := "vCPU Active Usage"
	if usageType == "idle" {
		name = "vCPU (idle)"
		meterName = "vCPU Idle Usage"
	}

	return &schema.CostComponent{
		Name:            name,
		Unit:            "vCPU-seconds",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(r.Region),
			Service:    strPtr("Azure Container Apps"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", ValueRegex: strPtr(fmt.Sprintf("/%s/i", meterName))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
		UsageBased: usageType == "active",
	}
}

func (r *ContainerApp) memoryCostComponent(profile string, usageType string, quantity *decimal.Decimal) *schema.CostComponent {
	if profile != "Consumption" && profile != "" {
		return nil
	}

	if quantity != nil && quantity.LessThanOrEqual(decimal.Zero) {
		return nil
	}

	name := "Memory (active)"
	meterName := "Memory Active Usage"
	if usageType == "idle" {
		name = "Memory (idle)"
		meterName = "Memory Idle Usage"
	}

	return &schema.CostComponent{
		Name:            name,
		Unit:            "GiB-seconds",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(r.Region),
			Service:    strPtr("Azure Container Apps"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", ValueRegex: strPtr(fmt.Sprintf("/%s/i", meterName))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
		UsageBased: usageType == "active",
	}
}

func (r *ContainerApp) requestsCostComponent(profile string, quantity *decimal.Decimal) *schema.CostComponent {
	if profile != "Consumption" && profile != "" {
		return nil
	}

	if quantity != nil && quantity.LessThanOrEqual(decimal.Zero) {
		return nil
	}

	return &schema.CostComponent{
		Name:            "Requests",
		Unit:            "1M requests",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(r.Region),
			Service:    strPtr("Azure Container Apps"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", ValueRegex: strPtr("/Requests/i")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
		UsageBased: true,
	}
}
