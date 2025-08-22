package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"
)

// ServicePlan struct represents a user commitment to an App Service Plan. A service plan has a dedicated
// amount of compute and storage and can be used to run any number of apps/containers.
//
// Resource information: https://learn.microsoft.com/en-us/azure/app-service/overview-hosting-plans
// Pricing information: https://azure.microsoft.com/en-gb/pricing/details/app-service/windows/
type ServicePlan struct {
	Address     string
	SKUName     string
	WorkerCount int64
	OSType      string
	Region      string
	IsDevTest   bool
}

func (r *ServicePlan) CoreType() string {
	return "ServicePlan"
}

func (r *ServicePlan) UsageSchema() []*schema.UsageItem {
	return nil
}

// PopulateUsage parses the u schema.UsageData into the ServicePlan struct
// It uses the `infracost_usage` struct tags to populate data into the ServicePlan
func (r *ServicePlan) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid ServicePlan struct.
//
// ServicePlan only has one cost component associated with the compute cost of the plan.
func (r *ServicePlan) BuildResource() *schema.Resource {
	productName := "Standard Plan"
	sku := r.SKUName

	if len(r.SKUName) < 2 || strings.ToLower(r.SKUName[:2]) == "ep" || strings.ToLower(r.SKUName[:2]) == "ws" || strings.ToLower(r.SKUName[:2]) == "y1" {
		return &schema.Resource{
			Name:        r.Address,
			IsSkipped:   true,
			NoPrice:     true,
			UsageSchema: r.UsageSchema(),
		}
	}

	firstLetter := strings.ToLower(r.SKUName[:1])
	os := strings.ToLower(r.OSType)
	var additionalAttributeFilters []*schema.AttributeFilter

	switch firstLetter {
	case "s":
		sku = "S" + r.SKUName[1:]
	case "b":
		sku = "B" + r.SKUName[1:]
		productName = "Basic Plan"
	case "f":
		productName = "Free Plan"
	case "d":
		sku = "Shared"
		productName = "Shared Plan"
	case "p", "i":
		sku, productName, additionalAttributeFilters = getVersionedAppServicePlanSKU(sku, os)
	}

	if strings.ToLower(r.SKUName) == "shared" {
		sku = "Shared"
		productName = "Shared Plan"
	}

	if os == "linux" && productName != "Isolated Plan" && productName != "Premium Plan" && productName != "Shared Plan" {
		productName += " - Linux"
	}

	purchaseOption := "Consumption"
	name := fmt.Sprintf("Instance usage (%s)", r.SKUName)
	if r.IsDevTest && strings.Contains(os, "windows") {
		purchaseOption = "DevTestConsumption"
		name = fmt.Sprintf("Instance usage (dev/test, %s)", r.SKUName)
	}

	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			servicePlanCostComponent(
				r.Region,
				name,
				productName,
				sku,
				r.WorkerCount,
				purchaseOption,
				additionalAttributeFilters...,
			),
		},
	}
}
