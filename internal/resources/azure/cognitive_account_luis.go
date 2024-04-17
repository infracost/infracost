package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

var validLUISCommitmentTierRequests = []int64{1_000_000, 5_000_000, 25_000_000}

// CognitiveAccountLUIS struct represents the Azure LUIS AI resource.
// This supports the pay-as-you pricing and the standard and connected container commitment tiers.
// This doesn't currently support the disconnected container commitment tier.
//
// The commitment tiers are implemented using a usage-based cost component for the commitment amount.
// Since multiple commitment tiers can be used at the same time, we use separate usage-based cost
// components for each commitment tier and overage, instead of using the same cost component as the
// pay-as-you-go pricing.
//
// Resource information: https://learn.microsoft.com/en-us/azure/ai-services/luis/what-is-luis
// Pricing information: https://azure.microsoft.com/en-gb/pricing/details/cognitive-services/language-understanding-intelligent-services/
type CognitiveAccountLUIS struct {
	Address string
	Region  string

	Sku string

	// Usage attributes

	MonthlyLUISTextRequests                                    *int64 `infracost_usage:"monthly_luis_text_requests"`
	MonthlyLUISSpeechRequests                                  *int64 `infracost_usage:"monthly_luis_speech_requests"`
	MonthlyCommitmentLUISTextRequests                          *int64 `infracost_usage:"monthly_commitment_luis_text_requests"`
	MonthlyCommitmentLUISTextOverageRequests                   *int64 `infracost_usage:"monthly_commitment_luis_text_overage_requests"`
	MonthlyConnectedContainerCommitmentLUISTextRequests        *int64 `infracost_usage:"monthly_connected_container_commitment_luis_text_requests"`
	MonthlyConnectedContainerCommitmentLUISTextOverageRequests *int64 `infracost_usage:"monthly_connected_container_commitment_luis_text_overage_requests"`
}

// // CoreType returns the name of this resource type
func (r *CognitiveAccountLUIS) CoreType() string {
	return "CognitiveAccountLUIS"
}

// UsageSchema defines a list which represents the usage schema of CognitiveAccountLUIS.
func (r *CognitiveAccountLUIS) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_luis_text_requests", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_luis_speech_requests", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_commitment_luis_text_requests", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_commitment_luis_text_overage_requests", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_connected_container_commitment_luis_text_requests", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_connected_container_commitment_luis_text_overage_requests", ValueType: schema.Int64, DefaultValue: 0},
	}
}

// PopulateUsage parses the u schema.UsageData into the CognitiveAccountLUIS.
// It uses the `infracost_usage` struct tags to populate data into the CognitiveAccountLUIS.
func (r *CognitiveAccountLUIS) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid CognitiveAccountLUIS struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *CognitiveAccountLUIS) BuildResource() *schema.Resource {
	// F0 is the free tier
	if strings.EqualFold(r.Sku, "f0") {
		return &schema.Resource{
			Name:      r.Address,
			NoPrice:   true,
			IsSkipped: true,
		}
	}

	// For some reason the SKU can either be S0 or S1 but they map to the same thing
	if !strings.EqualFold(r.Sku, "s0") && !strings.EqualFold(r.Sku, "s1") && !strings.EqualFold(r.Sku, "s") {
		logging.Logger.Warn().Msgf("Unsupported SKU %s for %s", r.Sku, r.Address)
		return nil
	}

	costComponents := make([]*schema.CostComponent, 0)

	if r.MonthlyCommitmentLUISTextRequests != nil {
		costComponents = append(costComponents, r.commitmentTextRequestsCostComponents()...)
	}

	if r.MonthlyConnectedContainerCommitmentLUISTextRequests != nil {
		costComponents = append(costComponents, r.connectedContainerCommitmentTextRequestsCostComponents()...)
	}

	if (r.MonthlyCommitmentLUISTextRequests == nil && r.MonthlyConnectedContainerCommitmentLUISTextRequests == nil) || r.MonthlyLUISTextRequests != nil {
		costComponents = append(costComponents, r.textRequestsCostComponent())
	}

	costComponents = append(costComponents, r.speechRequestsCostComponent())

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *CognitiveAccountLUIS) textRequestsCostComponent() *schema.CostComponent {
	var qty *decimal.Decimal
	if r.MonthlyLUISTextRequests != nil {
		qty = decimalPtr(decimal.NewFromInt(*r.MonthlyLUISTextRequests).Div(decimal.NewFromInt(1_000)))
	}

	return &schema.CostComponent{
		Name:            "Text requests",
		Unit:            "1K transactions",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr(vendorName),
			Region:        strPtr(r.Region),
			Service:       strPtr("Cognitive Services"),
			ProductFamily: strPtr("AI + Machine Learning"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Language Understanding")},
				{Key: "skuName", Value: strPtr("S1")},
				{Key: "meterName", Value: strPtr("S1 Transactions")},
			},
		},
	}
}

func (r *CognitiveAccountLUIS) commitmentTextRequestsCostComponents() []*schema.CostComponent {
	amount := *r.MonthlyCommitmentLUISTextRequests
	if !containsInt64(validLUISCommitmentTierRequests, amount) {
		logging.Logger.Warn().Msgf("Invalid commitment tier %d for %s", amount, r.Address)
		return nil
	}

	desc := amountToDescription(amount)
	skuName := "Commitment Tier Azure" + " " + desc

	costComponents := []*schema.CostComponent{
		{
			Name: "Text requests (commitment)",
			Unit: "1M transactions",
			// Use a monthly quantity of 1 and a unit multiplier so we show the
			// correct number of transactions in the qty field, and divide the price
			// by the number of transactions in the commitment tier
			UnitMultiplier:  decimal.NewFromInt(1).Div(decimal.NewFromInt(amount)),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Language Understanding")},
					{Key: "skuName", Value: strPtr(skuName)},
					{Key: "meterName", ValueRegex: regexPtr("Unit$")},
				},
			},
		},
	}

	if r.MonthlyCommitmentLUISTextOverageRequests != nil {
		overageQty := decimalPtr(decimal.NewFromInt(*r.MonthlyCommitmentLUISTextOverageRequests).Div(decimal.NewFromInt(1_000)))

		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Text requests (commitment overage)",
			Unit:            "1K transactions",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: overageQty,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Language Understanding")},
					{Key: "skuName", Value: strPtr(skuName)},
					{Key: "meterName", ValueRegex: regexPtr("Overage")},
				},
			},
		})
	}

	return costComponents
}

func (r *CognitiveAccountLUIS) connectedContainerCommitmentTextRequestsCostComponents() []*schema.CostComponent {
	amount := *r.MonthlyConnectedContainerCommitmentLUISTextRequests
	if !containsInt64(validLUISCommitmentTierRequests, amount) {
		logging.Logger.Warn().Msgf("Invalid commitment tier %d for %s", amount, r.Address)
		return nil
	}

	desc := amountToDescription(amount)
	skuName := "Commitment Tier Azure" + " " + desc

	qty := decimal.NewFromInt(amount).Div(decimal.NewFromInt(1_000_000))

	costComponents := []*schema.CostComponent{
		{
			Name: "Text requests (connected container commitment)",
			Unit: "1M transactions",
			// Use a monthly quantity of 1 and a unit multiplier so we show the
			// correct number of transactions in the qty field, and divide the price
			// by the number of transactions in the commitment tier
			UnitMultiplier:  decimal.NewFromInt(1).Div(qty),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Language Understanding")},
					{Key: "skuName", Value: strPtr(skuName)},
					{Key: "meterName", ValueRegex: strPtr("Unit$")},
				},
			},
		},
	}

	if r.MonthlyConnectedContainerCommitmentLUISTextOverageRequests != nil {
		overageQty := decimalPtr(decimal.NewFromInt(*r.MonthlyConnectedContainerCommitmentLUISTextOverageRequests).Div(decimal.NewFromInt(1_000)))

		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Text requests (connected container commitment overage)",
			Unit:            "1K transactions",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: overageQty,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Language Understanding")},
					{Key: "skuName", Value: strPtr(skuName)},
					{Key: "meterName", ValueRegex: regexPtr("Overage")},
				},
			},
		})
	}

	return costComponents
}

func (r *CognitiveAccountLUIS) speechRequestsCostComponent() *schema.CostComponent {
	var qty *decimal.Decimal
	if r.MonthlyLUISSpeechRequests != nil {
		qty = decimalPtr(decimal.NewFromInt(*r.MonthlyLUISSpeechRequests).Div(decimal.NewFromInt(1_000)))
	}

	c := &schema.CostComponent{
		Name:            "Speech requests",
		Unit:            "1K transactions",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
	}

	// The Azure Retail API doesn't have prices for this
	if strings.HasPrefix(r.Region, "usgov") {
		c.SetCustomPrice(decimalPtr(decimal.NewFromFloat(6.875)))
	} else {
		c.SetCustomPrice(decimalPtr(decimal.NewFromFloat(5.5)))
	}

	return c
}
