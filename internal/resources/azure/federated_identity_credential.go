package azure

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// FederatedIdentityCredential struct represents an Azure Federated Identity
// Credential. are a new type of credential that enables workload identity
// federation for software workloads. Workload identity federation allows you to
// access Microsoft Entra protected resources without needing to manage secrets
// (for supported scenarios).
//
// Resource information: https://learn.microsoft.com/en-us/graph/api/resources/federatedidentitycredentials-overview?view=graph-rest-1.0
// Pricing information: https://azure.microsoft.com/en-us/pricing/details/active-directory-external-identities/
type FederatedIdentityCredential struct {
	Address string
	Region  string

	MonthlyActiveP1Users *int64 `infracost_usage:"monthly_active_p1_users"`
	MonthlyActiveP2Users *int64 `infracost_usage:"monthly_active_p2_users"`
}

// CoreType returns the name of this resource type
func (r *FederatedIdentityCredential) CoreType() string {
	return "FederatedIdentityCredential"
}

// UsageSchema defines a list which represents the usage schema of FederatedIdentityCredential.
func (r *FederatedIdentityCredential) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_active_p1_users", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_active_p2_users", ValueType: schema.Int64, DefaultValue: 0},
	}
}

// PopulateUsage parses the u schema.UsageData into the FederatedIdentityCredential.
// It uses the `infracost_usage` struct tags to populate data into the FederatedIdentityCredential.
func (r *FederatedIdentityCredential) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid
// FederatedIdentityCredential struct. This method is called after the resource
// is initialised by an IaC provider. See providers folder for more information.
//
// BuildResource returns cost components for the monthly active users for both P1
// and P2 licence types for Microsoft Entra. It is not possible to infer the
// licence type from the IaC code, so we rely on the user to provide the monthly
// active users for each licence type as a usage parameter. The resource can not
// have both P1 and P2 licence types at the same time, so we check which one is
// set and return the cost component for that licence type.
func (r *FederatedIdentityCredential) BuildResource() *schema.Resource {
	if r.MonthlyActiveP1Users != nil {
		return &schema.Resource{
			Name:        r.Address,
			UsageSchema: r.UsageSchema(),
			CostComponents: []*schema.CostComponent{
				r.activeUserComponent("P1", decimalPtr(decimal.NewFromInt(*r.MonthlyActiveP1Users))),
			},
		}
	}

	if r.MonthlyActiveP2Users != nil {
		return &schema.Resource{
			Name:        r.Address,
			UsageSchema: r.UsageSchema(),
			CostComponents: []*schema.CostComponent{
				r.activeUserComponent("P2", decimalPtr(decimal.NewFromInt(*r.MonthlyActiveP2Users))),
			},
		}
	}

	return &schema.Resource{
		Name:        r.Address,
		UsageSchema: r.UsageSchema(),
		CostComponents: []*schema.CostComponent{
			r.activeUserComponent("P1", nil),
			r.activeUserComponent("P2", nil),
		},
	}
}

func (r *FederatedIdentityCredential) activeUserComponent(licence string, quantity *decimal.Decimal) *schema.CostComponent {
	var minusFreeTier *decimal.Decimal
	if quantity != nil {
		val := quantity.Sub(decimal.NewFromInt(50000))
		if val.LessThan(decimal.NewFromInt(0)) {
			minusFreeTier = decimalPtr(decimal.NewFromInt(0))
		} else {
			minusFreeTier = decimalPtr(val)
		}
	}

	titled := strings.ToUpper(licence)

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Active users (%s)", titled),
		Unit:            "users",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: minusFreeTier,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Azure Active Directory for External Identities"),
			ProductFamily: strPtr("Security"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr(titled)},
				{Key: "meterName", Value: strPtr(titled + " Monthly Active Users")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("50000"),
		},
		UsageBased: true,
	}
}
