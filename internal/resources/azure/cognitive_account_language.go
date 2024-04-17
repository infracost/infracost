package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
	"github.com/shopspring/decimal"
)

var validLanguageCommitmentTierTextAnalyticsRecords = []int64{1_000_000, 3_000_000, 10_000_000}
var validLanguageCommitmentTierSummarizationRecords = []int64{3_000_000, 10_000_000}

// CognitiveAccountLanguage struct represents the Azure AI Language Service.
// This supports the pay-as-you pricing and the standard and connected container commitment tiers.
// This doesn't currently support the disconnected container commitment tier.
//
// The commitment tiers are implemented using a usage-based cost component for the commitment amount.
// Since multiple commitment tiers can be used at the same time, we use separate usage-based cost
// components for each commitment tier and overage, instead of using the same cost component as the
// pay-as-you-go pricing.
//
// Resource information: https://learn.microsoft.com/en-us/azure/ai-services/language-service/overview
// Pricing information: https://azure.microsoft.com/en-gb/pricing/details/cognitive-services/language-service/
type CognitiveAccountLanguage struct {
	Address string
	Region  string

	Sku string

	// Usage attributes

	MonthlyLanguageTextAnalyticsRecords *int64 `infracost_usage:"monthly_language_text_analytics_records"`
	MonthlyLanguageSummarizationRecords *int64 `infracost_usage:"monthly_language_summarization_records"`

	MonthlyLanguageConversationalLanguageUnderstandingRecords               *int64   `infracost_usage:"monthly_language_conversational_language_understanding_records"`
	MonthlyLanguageConversationalLanguageUnderstandingAdvancedTrainingHours *float64 `infracost_usage:"monthly_language_conversational_language_understanding_advanced_training_hours"`
	MonthlyLanguageCustomizedTextClassificationRecords                      *int64   `infracost_usage:"monthly_language_customized_text_classification_records"`
	MonthlyLanguageCustomizedSummarizationRecords                           *int64   `infracost_usage:"monthly_language_customized_summarization_records"`
	MonthlyLanguageCustomizedQuestionAnsweringRecords                       *int64   `infracost_usage:"monthly_language_customized_question_answering_records"`
	MonthlyLanguageCustomizedTrainingHours                                  *float64 `infracost_usage:"monthly_language_customized_training_hours"`
	MonthlyLanguageTextAnalyticsForHealthRecords                            *int64   `infracost_usage:"monthly_language_text_analytics_for_health_records"`

	// Commitment tiers
	MonthlyCommitmentLanguageTextAnalyticsRecords                          *int64 `infracost_usage:"monthly_commitment_language_text_analytics_records"`
	MonthlyCommitmentLanguageTextAnalyticsOverageRecords                   *int64 `infracost_usage:"monthly_commitment_language_text_analytics_overage_records"`
	MonthlyCommitmentLanguageSummarizationRecords                          *int64 `infracost_usage:"monthly_commitment_language_summarization_records"`
	MonthlyCommitmentLanguageSummarizationOverageRecords                   *int64 `infracost_usage:"monthly_commitment_language_summarization_overage_records"`
	MonthlyConnectedContainerCommitmentLanguageTextAnalyticsRecords        *int64 `infracost_usage:"monthly_connected_container_commitment_language_text_analytics_records"`
	MonthlyConnectedContainerCommitmentLanguageTextAnalyticsOverageRecords *int64 `infracost_usage:"monthly_connected_container_commitment_language_text_analytics_overage_records"`
	MonthlyConnectedContainerCommitmentLanguageSummarizationRecords        *int64 `infracost_usage:"monthly_connected_container_commitment_language_summarization_records"`
	MonthlyConnectedContainerCommitmentLanguageSummarizationOverageRecords *int64 `infracost_usage:"monthly_connected_container_commitment_language_summarization_overage_records"`
}

// // CoreType returns the name of this resource type
func (r *CognitiveAccountLanguage) CoreType() string {
	return "CognitiveAccountLanguage"
}

// UsageSchema defines a list which represents the usage schema of CognitiveAccountLanguage.
func (r *CognitiveAccountLanguage) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_language_text_analytics_records", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_language_summarization_records", DefaultValue: 0, ValueType: schema.Int64},

		{Key: "monthly_language_conversational_language_understanding_records", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_language_conversational_language_understanding_advanced_training_hours", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "monthly_language_customized_text_classification_records", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_language_customized_summarization_records", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_language_customized_question_answering_records", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_language_customized_training_hours", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "monthly_language_text_analytics_for_health_records", DefaultValue: 0, ValueType: schema.Int64},

		// Commitment tiers
		{Key: "monthly_commitment_language_text_analytics_records", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_commitment_language_text_analytics_overage_records", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_commitment_language_summarization_records", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_commitment_language_summarization_overage_records", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_connected_container_commitment_language_text_analytics_records", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_connected_container_commitment_language_text_analytics_overage_records", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_connected_container_commitment_language_summarization_records", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_connected_container_commitment_language_summarization_overage_records", DefaultValue: 0, ValueType: schema.Int64},
	}
}

// PopulateUsage parses the u schema.UsageData into the CognitiveAccountLanguage.
// It uses the `infracost_usage` struct tags to populate data into the CognitiveAccountLanguage.
func (r *CognitiveAccountLanguage) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid CognitiveAccountLanguage struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *CognitiveAccountLanguage) BuildResource() *schema.Resource {
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

	if r.MonthlyCommitmentLanguageTextAnalyticsRecords != nil {
		costComponents = append(costComponents, r.commitmentTextAnalyticsCostComponents(standardCommitmentTier, *r.MonthlyCommitmentLanguageTextAnalyticsRecords, intPtrToDecimalPtr(r.MonthlyCommitmentLanguageTextAnalyticsOverageRecords))...)
	}
	if r.MonthlyConnectedContainerCommitmentLanguageTextAnalyticsRecords != nil {
		costComponents = append(costComponents, r.commitmentTextAnalyticsCostComponents(connectedContainerCommitmentTier, *r.MonthlyConnectedContainerCommitmentLanguageTextAnalyticsRecords, intPtrToDecimalPtr(r.MonthlyConnectedContainerCommitmentLanguageTextAnalyticsOverageRecords))...)
	}
	if (r.MonthlyCommitmentLanguageTextAnalyticsRecords == nil && r.MonthlyConnectedContainerCommitmentLanguageTextAnalyticsRecords == nil) || r.MonthlyLanguageTextAnalyticsRecords != nil {
		costComponents = append(costComponents, r.textAnalyticsCostComponents()...)
	}

	if r.MonthlyCommitmentLanguageSummarizationRecords != nil {
		costComponents = append(costComponents, r.commitmentSummarizationCostComponents(standardCommitmentTier, *r.MonthlyCommitmentLanguageSummarizationRecords, intPtrToDecimalPtr(r.MonthlyCommitmentLanguageSummarizationOverageRecords))...)
	}
	if r.MonthlyConnectedContainerCommitmentLanguageSummarizationRecords != nil {
		costComponents = append(costComponents, r.commitmentSummarizationCostComponents(connectedContainerCommitmentTier, *r.MonthlyConnectedContainerCommitmentLanguageSummarizationRecords, intPtrToDecimalPtr(r.MonthlyConnectedContainerCommitmentLanguageSummarizationOverageRecords))...)
	}
	if (r.MonthlyCommitmentLanguageSummarizationRecords == nil && r.MonthlyConnectedContainerCommitmentLanguageSummarizationRecords == nil) || r.MonthlyLanguageSummarizationRecords != nil {
		costComponents = append(costComponents, r.summarizationCostComponent())
	}

	costComponents = append(costComponents, r.conversationalLanguageUnderstandingCostComponent())
	costComponents = append(costComponents, r.conversationalLanguageUnderstandingAdvancedTrainingCostComponent())
	costComponents = append(costComponents, r.customizedTextClassificationCostComponent())
	costComponents = append(costComponents, r.customizedSummarizationCostComponent())
	costComponents = append(costComponents, r.customizedQuestionAnsweringCostComponents()...)
	costComponents = append(costComponents, r.customizedTrainingCostComponent())
	costComponents = append(costComponents, r.textAnalyticsForHealthCostComponents()...)

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *CognitiveAccountLanguage) textAnalyticsCostComponents() []*schema.CostComponent {
	var costComponents []*schema.CostComponent

	tierLimits := []int{500, 2_000, 7_500}
	tierData := []struct {
		suffix     string
		startUsage string
	}{
		{suffix: " (first 500K)", startUsage: "0"},
		{suffix: " (500K-2.5M)", startUsage: "500"},
		{suffix: " (2.5M-10M)", startUsage: "2500"},
		{suffix: " (over 10M)", startUsage: "10000"},
	}

	var qty *decimal.Decimal
	if r.MonthlyLanguageTextAnalyticsRecords != nil {
		qty = decimalPtr(decimal.NewFromInt(*r.MonthlyLanguageTextAnalyticsRecords).Div(decimal.NewFromInt(1_000)))
		tierQtys := usage.CalculateTierBuckets(*qty, tierLimits)

		for i, d := range tierData {
			if len(tierQtys) <= i {
				break
			}

			if tierQtys[i].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, r.textAnalyticsCostComponent(d.suffix, d.startUsage, decimalPtr(tierQtys[i])))
			}
		}
	} else {
		costComponents = append(costComponents, r.textAnalyticsCostComponent(tierData[1].suffix, tierData[1].startUsage, nil))
	}

	return costComponents
}

func (r *CognitiveAccountLanguage) textAnalyticsCostComponent(suffix string, startUsage string, qty *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Text analytics%s", suffix),
		Unit:            "1K records",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr(vendorName),
			Region:        strPtr(r.Region),
			Service:       strPtr("Cognitive Services"),
			ProductFamily: strPtr("AI + Machine Learning"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Language")},
				{Key: "skuName", Value: strPtr("Standard")},
				{Key: "meterName", Value: strPtr("Standard Text Records")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(startUsage),
		},
	}
}

func (r *CognitiveAccountLanguage) commitmentTextAnalyticsCostComponents(commitmentTierType int64, amount int64, overage *decimal.Decimal) []*schema.CostComponent {
	if !containsInt64(validLanguageCommitmentTierTextAnalyticsRecords, amount) {
		logging.Logger.Warn().Msgf("Invalid commitment tier %d for %s", amount, r.Address)
		return nil
	}

	desc := amountToDescription(amount)

	qty := decimal.NewFromInt(amount).Div(decimal.NewFromInt(1_000))

	skuName := "Commitment Tier Azure " + desc
	commitmentLabel := "commitment"
	if commitmentTierType == connectedContainerCommitmentTier {
		skuName = "Commitment Tier Connected " + desc
		commitmentLabel = "connected container commitment"
	}

	costComponents := []*schema.CostComponent{
		{
			Name: fmt.Sprintf("Text analytics (%s)", commitmentLabel),
			Unit: "1K records",
			// Use a monthly quantity of 1 and a unit multiplier so we show the
			// correct number of transactions in the qty field, and divide the price
			// by the number of transactions in the commitment tier
			UnitMultiplier:  decimal.NewFromInt(1).Div(qty),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			UnitRounding:    int32Ptr(0),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Language")},
					{Key: "skuName", Value: strPtr(skuName)},
					{Key: "meterName", ValueRegex: regexPtr("Unit$")},
				},
			},
		},
	}

	if overage != nil {
		overageQty := decimalPtr(overage.Div(decimal.NewFromInt(1_000)))

		costComponents = append(costComponents, &schema.CostComponent{
			Name:            fmt.Sprintf("Text requests (%s overage)", commitmentLabel),
			Unit:            "1K records",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: overageQty,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Language")},
					{Key: "skuName", Value: strPtr(skuName)},
					{Key: "meterName", ValueRegex: regexPtr("Overage")},
				},
			},
		})
	}

	return costComponents
}

func (r *CognitiveAccountLanguage) summarizationCostComponent() *schema.CostComponent {
	var qty *decimal.Decimal
	if r.MonthlyLanguageSummarizationRecords != nil {
		qty = decimalPtr(decimal.NewFromInt(*r.MonthlyLanguageSummarizationRecords).Div(decimal.NewFromInt(1_000)))
	}

	return &schema.CostComponent{
		Name:            "Summarization",
		Unit:            "1K records",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr(vendorName),
			Region:        strPtr(r.Region),
			Service:       strPtr("Cognitive Services"),
			ProductFamily: strPtr("AI + Machine Learning"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Language")},
				{Key: "skuName", Value: strPtr("Standard")},
				{Key: "meterName", Value: strPtr("Standard Summarization Text Records")},
			},
		},
	}
}

func (r *CognitiveAccountLanguage) commitmentSummarizationCostComponents(commitmentTierType int64, amount int64, overage *decimal.Decimal) []*schema.CostComponent {
	if !containsInt64(validLanguageCommitmentTierSummarizationRecords, amount) {
		logging.Logger.Warn().Msgf("Invalid commitment tier %d for %s", amount, r.Address)
		return nil
	}

	desc := amountToDescription(amount)

	qty := decimal.NewFromInt(amount).Div(decimal.NewFromInt(1_000))

	skuName := "Commitment Tier Summarization Azure " + desc
	commitmentLabel := "commitment"
	if commitmentTierType == connectedContainerCommitmentTier {
		skuName = "Commitment Tier Summarization Connected " + desc
		commitmentLabel = "connected container commitment"
	}

	costComponents := []*schema.CostComponent{
		{
			Name: fmt.Sprintf("Summarization (%s)", commitmentLabel),
			Unit: "1K records",
			// Use a monthly quantity of 1 and a unit multiplier so we show the
			// correct number of transactions in the qty field, and divide the price
			// by the number of transactions in the commitment tier
			UnitMultiplier:  decimal.NewFromInt(1).Div(qty),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			UnitRounding:    int32Ptr(0),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Language")},
					{Key: "skuName", Value: strPtr(skuName)},
					{Key: "meterName", ValueRegex: regexPtr("Unit$")},
				},
			},
		},
	}

	if overage != nil {
		overageQty := decimalPtr(overage.Div(decimal.NewFromInt(1_000)))

		costComponents = append(costComponents, &schema.CostComponent{
			Name:            fmt.Sprintf("Summarization (%s overage)", commitmentLabel),
			Unit:            "1K records",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: overageQty,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Language")},
					{Key: "skuName", Value: strPtr(skuName)},
					{Key: "meterName", ValueRegex: regexPtr("Overage")},
				},
			},
		})
	}

	return costComponents
}

func (r *CognitiveAccountLanguage) conversationalLanguageUnderstandingCostComponent() *schema.CostComponent {
	var qty *decimal.Decimal
	if r.MonthlyLanguageConversationalLanguageUnderstandingRecords != nil {
		qty = decimalPtr(decimal.NewFromInt(*r.MonthlyLanguageConversationalLanguageUnderstandingRecords).Div(decimal.NewFromInt(1_000)))
	}

	return &schema.CostComponent{
		Name:            "Conversational language understanding",
		Unit:            "1K records",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr(vendorName),
			Region:        strPtr(r.Region),
			Service:       strPtr("Cognitive Services"),
			ProductFamily: strPtr("AI + Machine Learning"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Language")},
				{Key: "skuName", Value: strPtr("Standard")},
				{Key: "meterName", Value: strPtr("Standard CLU Text Records")},
			},
		},
	}
}

func (r *CognitiveAccountLanguage) conversationalLanguageUnderstandingAdvancedTrainingCostComponent() *schema.CostComponent {
	var qty *decimal.Decimal
	if r.MonthlyLanguageConversationalLanguageUnderstandingAdvancedTrainingHours != nil {
		qty = decimalPtr(decimal.NewFromFloat(*r.MonthlyLanguageConversationalLanguageUnderstandingAdvancedTrainingHours))
	}

	return &schema.CostComponent{
		Name:            "Conversational language understanding advanced training",
		Unit:            "hour",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr(vendorName),
			Region:        strPtr(r.Region),
			Service:       strPtr("Cognitive Services"),
			ProductFamily: strPtr("AI + Machine Learning"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Language")},
				{Key: "skuName", Value: strPtr("Standard")},
				{Key: "meterName", Value: strPtr("Standard CLU Advanced Training Unit")},
			},
		},
	}
}

func (r *CognitiveAccountLanguage) customizedTextClassificationCostComponent() *schema.CostComponent {
	var qty *decimal.Decimal
	if r.MonthlyLanguageCustomizedTextClassificationRecords != nil {
		qty = decimalPtr(decimal.NewFromInt(*r.MonthlyLanguageCustomizedTextClassificationRecords).Div(decimal.NewFromInt(1_000)))
	}

	return &schema.CostComponent{
		Name:            "Customized text classification",
		Unit:            "1K records",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr(vendorName),
			Region:        strPtr(r.Region),
			Service:       strPtr("Cognitive Services"),
			ProductFamily: strPtr("AI + Machine Learning"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Language")},
				{Key: "skuName", Value: strPtr("Standard")},
				{Key: "meterName", Value: strPtr("Standard Custom Text Records")},
			},
		},
	}
}

func (r *CognitiveAccountLanguage) customizedSummarizationCostComponent() *schema.CostComponent {
	var qty *decimal.Decimal
	if r.MonthlyLanguageCustomizedSummarizationRecords != nil {
		qty = decimalPtr(decimal.NewFromInt(*r.MonthlyLanguageCustomizedSummarizationRecords).Div(decimal.NewFromInt(1_000)))
	}

	return &schema.CostComponent{
		Name:            "Customized summarization",
		Unit:            "1K records",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr(vendorName),
			Region:        strPtr(r.Region),
			Service:       strPtr("Cognitive Services"),
			ProductFamily: strPtr("AI + Machine Learning"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Language")},
				{Key: "skuName", Value: strPtr("Standard")},
				{Key: "meterName", Value: strPtr("Standard Custom Summarization Text Records")},
			},
		},
	}
}

func (r *CognitiveAccountLanguage) customizedQuestionAnsweringCostComponents() []*schema.CostComponent {
	var costComponents []*schema.CostComponent

	tierLimits := []int{2_500}
	tierData := []struct {
		suffix     string
		startUsage string
	}{
		{suffix: " (first 2.5M)", startUsage: "0"},
		{suffix: " (over 2.5M)", startUsage: "2500"},
	}

	var qty *decimal.Decimal
	if r.MonthlyLanguageCustomizedQuestionAnsweringRecords != nil {
		qty = decimalPtr(decimal.NewFromInt(*r.MonthlyLanguageCustomizedQuestionAnsweringRecords).Div(decimal.NewFromInt(1_000)))
		tierQtys := usage.CalculateTierBuckets(*qty, tierLimits)

		for i, d := range tierData {
			if len(tierQtys) <= i {
				break
			}

			if tierQtys[i].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, r.customizedQuestionAnsweringCostComponent(d.suffix, d.startUsage, decimalPtr(tierQtys[i])))
			}
		}
	} else {
		costComponents = append(costComponents, r.customizedQuestionAnsweringCostComponent(tierData[1].suffix, tierData[1].startUsage, nil))
	}

	return costComponents
}

func (r *CognitiveAccountLanguage) customizedQuestionAnsweringCostComponent(suffix string, startUsage string, qty *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Customized question answering%s", suffix),
		Unit:            "1K records",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr(vendorName),
			Region:        strPtr(r.Region),
			Service:       strPtr("Cognitive Services"),
			ProductFamily: strPtr("AI + Machine Learning"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Language")},
				{Key: "skuName", Value: strPtr("Standard")},
				{Key: "meterName", Value: strPtr("Standard QA Text Records")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(startUsage),
		},
	}
}

func (r *CognitiveAccountLanguage) customizedTrainingCostComponent() *schema.CostComponent {
	var qty *decimal.Decimal
	if r.MonthlyLanguageCustomizedTrainingHours != nil {
		qty = decimalPtr(decimal.NewFromFloat(*r.MonthlyLanguageCustomizedTrainingHours))
	}

	return &schema.CostComponent{
		Name:            "Customized training",
		Unit:            "hour",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr(vendorName),
			Region:        strPtr(r.Region),
			Service:       strPtr("Cognitive Services"),
			ProductFamily: strPtr("AI + Machine Learning"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Language")},
				{Key: "skuName", Value: strPtr("Standard")},
				{Key: "meterName", Value: strPtr("Standard Custom Training Unit")},
			},
		},
	}
}

func (r *CognitiveAccountLanguage) textAnalyticsForHealthCostComponents() []*schema.CostComponent {
	var costComponents []*schema.CostComponent

	tierLimits := []int{5, 495, 2_000, 7_500}
	tierData := []struct {
		suffix     string
		startUsage string
	}{
		{suffix: " (first 5K)", startUsage: "0"},
		{suffix: " (5K-500K)", startUsage: "5"},
		{suffix: " (500K-2.5M)", startUsage: "500"},
		{suffix: " (2.5M-10M)", startUsage: "2500"},
		{suffix: " (over 10M)", startUsage: "10000"},
	}

	var qty *decimal.Decimal
	if r.MonthlyLanguageTextAnalyticsForHealthRecords != nil {
		qty = decimalPtr(decimal.NewFromInt(*r.MonthlyLanguageTextAnalyticsForHealthRecords).Div(decimal.NewFromInt(1_000)))
		tierQtys := usage.CalculateTierBuckets(*qty, tierLimits)

		for i, d := range tierData {
			// Skip the first tier since it's free
			if i == 0 {
				continue
			}

			if len(tierQtys) <= i {
				break
			}

			if tierQtys[i].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, r.textAnalyticsForHealthCostComponent(d.suffix, d.startUsage, decimalPtr(tierQtys[i])))
			}
		}
	} else {
		costComponents = append(costComponents, r.textAnalyticsForHealthCostComponent(tierData[1].suffix, tierData[1].startUsage, nil))
	}

	return costComponents
}

func (r *CognitiveAccountLanguage) textAnalyticsForHealthCostComponent(suffix string, startUsage string, qty *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Text analytics for health%s", suffix),
		Unit:            "1K records",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr(vendorName),
			Region:        strPtr(r.Region),
			Service:       strPtr("Cognitive Services"),
			ProductFamily: strPtr("AI + Machine Learning"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Language")},
				{Key: "skuName", Value: strPtr("Standard")},
				{Key: "meterName", Value: strPtr("Standard Health Text Records")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(startUsage),
		},
	}
}
