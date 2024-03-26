package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

var validCommitmentTierHrs = []int64{2_000, 10_000, 50_000}
var validCommitmentTierChars = []int64{80_000_000, 400_000_000, 2_000_000_000}

const (
	standardCommitmentTier           = iota
	connectedContainerCommitmentTier = iota
)

// CognitiveAccountSpeech struct represents the Azure Speech AI resource.
// This supports the pay-as-you pricing and the standard and connected container commitment tiers.
// This doesn't currently support the disconnected container commitment tier.
//
// The commitment tiers are implemented using a usage-based cost component for the commitment amount.
// Since multiple commitment tiers can be used at the same time, we use separate usage-based cost
// components for each commitment tier and overage, instead of using the same cost component as the
// pay-as-you-go pricing.
//
// Resource information:https://azure.microsoft.com/en-us/products/ai-services/ai-speech
// Pricing information: https://azure.microsoft.com/en-gb/pricing/details/cognitive-services/speech-services/
type CognitiveAccountSpeech struct {
	Address string
	Region  string

	Kind string
	Sku  string

	// Usage attributes

	// Speech to text
	SpeechToTextStandardMonthlyHrs                                   *float64 `infracost_usage:"speech_to_text_standard_monthly_hrs"`
	SpeechToTextBatchMonthlyHrs                                      *float64 `infracost_usage:"speech_to_text_standard_batch_monthly_hrs"`
	SpeechToTextCustomModelMonthlyHrs                                *float64 `infracost_usage:"speech_to_text_custom_monthly_hrs"`
	SpeechToTextCustomModelBatchMonthlyHrs                           *float64 `infracost_usage:"speech_to_text_custom_batch_monthly_hrs"`
	SpeechToTextCustomEndpointMonthlyHrs                             *float64 `infracost_usage:"speech_to_text_custom_endpoint_monthly_hrs"`
	SpeechToTextConversationTranscriptionMultiChannelAudioMonthlyHrs *float64 `infracost_usage:"speech_to_text_conversation_transcription_multi_channel_audio_monthly_hrs"`
	SpeechToTextCustomTrainingMonthlyHrs                             *float64 `infracost_usage:"speech_to_text_custom_training_monthly_hrs"`
	SpeechToTextEnhancedAddOnsMonthlyHrs                             *float64 `infracost_usage:"speech_to_text_enhanced_add_ons_monthly_hrs"`

	// Text to speech
	TextToSpeechNeuralChars                    *int64   `infracost_usage:"text_to_speech_neural_chars"`
	TextToSpeechCustomNeuralTrainingMonthlyHrs *float64 `infracost_usage:"text_to_speech_custom_neural_training_monthly_hrs"`
	TextToSpeechCustomNeuralChars              *int64   `infracost_usage:"text_to_speech_custom_neural_chars"`
	TextToSpeechCustomNeuralEndpointMonthlyHrs *float64 `infracost_usage:"text_to_speech_custom_neural_endpoint_monthly_hrs"`
	TextToSpeechLongAudioChars                 *int64   `infracost_usage:"text_to_speech_long_audio_chars"`
	TextToSpeechPersonalVoiceProfiles          *int64   `infracost_usage:"text_to_speech_personal_voice_profiles"`
	TextToSpeechPersonalVoiceChars             *int64   `infracost_usage:"text_to_speech_personal_voice_chars"`

	// Speech translation
	SpeechTranslationMonthlyHrs *float64 `infracost_usage:"speech_translation_monthly_hrs"`

	// Speaker recognition
	SpeakerVerificationTransactions   *int64 `infracost_usage:"speaker_verification_transactions"`
	SpeakerIdentificationTransactions *int64 `infracost_usage:"speaker_identification_transactions"`

	// Voice storage
	VoiceProfiles *int64 `infracost_usage:"voice_profiles"`

	// Standard commitment tier
	SpeechToTextCommitmentHrs               *int64   `infracost_usage:"speech_to_text_standard_commitment_hrs"`
	SpeechToTextOverageHrs                  *float64 `infracost_usage:"speech_to_text_standard_overage_hrs"`
	SpeechToTextCustomModelCommitmentHrs    *int64   `infracost_usage:"speech_to_text_custom_commitment_hrs"`
	SpeechToTextCustomModelOverageHrs       *float64 `infracost_usage:"speech_to_text_custom_overage_hrs"`
	SpeechToTextEnhancedAddOnsCommitmentHrs *int64   `infracost_usage:"speech_to_text_enhanced_add_ons_commitment_hrs"`
	SpeechToTextEnhancedAddOnsOverageHrs    *float64 `infracost_usage:"speech_to_text_enhanced_add_ons_overage_hrs"`
	TextToSpeechNeuralCommitmentChars       *int64   `infracost_usage:"text_to_speech_neural_commitment_chars"`
	TextToSpeechNeuralOverageChars          *int64   `infracost_usage:"text_to_speech_neural_overage_chars"`

	// Connected container
	ConnectedContainerSpeechToTextCommitmentHrs               *int64   `infracost_usage:"connected_container_speech_to_text_standard_commitment_hrs"`
	ConnectedContainerSpeechToTextOverageHrs                  *float64 `infracost_usage:"connected_container_speech_to_text_standard_overage_hrs"`
	ConnectedContainerSpeechToTextCustomModelCommitmentHrs    *int64   `infracost_usage:"connected_container_speech_to_text_custom_commitment_hrs"`
	ConnectedContainerSpeechToTextCustomModelOverageHrs       *float64 `infracost_usage:"connected_container_speech_to_text_custom_overage_hrs"`
	ConnectedContainerSpeechToTextEnhancedAddOnsCommitmentHrs *int64   `infracost_usage:"connected_container_speech_to_text_enhanced_add_ons_commitment_hrs"`
	ConnectedContainerSpeechToTextEnhancedAddOnsOverageHrs    *float64 `infracost_usage:"connected_container_speech_to_text_enhanced_add_ons_overage_hrs"`
	ConnectedContainerTextToSpeechNeuralCommitmentChars       *int64   `infracost_usage:"connected_container_text_to_speech_neural_commitment_chars"`
	ConnectedContainerTextToSpeechNeuralOverageChars          *int64   `infracost_usage:"connected_container_text_to_speech_neural_overage_chars"`
}

// // CoreType returns the name of this resource type
func (r *CognitiveAccountSpeech) CoreType() string {
	return "CognitiveAccountSpeech"
}

// UsageSchema defines a list which represents the usage schema of CognitiveAccountSpeech.
func (r *CognitiveAccountSpeech) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		// Speech to text
		{Key: "speech_to_text_monthly_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "speech_to_text_custom_monthly_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "speech_to_text_batch_monthly_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "speech_to_text_custom_batch_monthly_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "speech_to_text_custom_endpoint_monthly_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "speech_to_text_conversation_transcription_multi_channel_audio_monthly_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "speech_to_text_custom_training_monthly_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "speech_to_text_enhanced_add_ons_monthly_hrs", ValueType: schema.Float64, DefaultValue: 0},
		// Text to speech
		{Key: "text_to_speech_neural_chars", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "text_to_speech_custom_neural_training_monthly_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "text_to_speech_custom_neural_chars", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "text_to_speech_custom_neural_endpoint_monthly_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "text_to_speech_long_audio_chars", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "text_to_speech_personal_voice_profiles", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "text_to_speech_personal_voice_chars", ValueType: schema.Int64, DefaultValue: 0},
		// Speech translation
		{Key: "speech_translation_monthly_hrs", ValueType: schema.Float64, DefaultValue: 0},
		// Speaker recognition
		{Key: "speaker_verification_transactions", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "speaker_identification_transactions", ValueType: schema.Int64, DefaultValue: 0},
		// Voice storage
		{Key: "voice_profiles", ValueType: schema.Int64, DefaultValue: 0},

		// Standard commitment tier
		{Key: "speech_to_text_standard_commitment_hrs", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "speech_to_text_standard_overage_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "speech_to_text_custom_commitment_hrs", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "speech_to_text_custom_overage_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "speech_to_text_enhanced_add_ons_commitment_hrs", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "speech_to_text_enhanced_add_ons_overage_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "text_to_speech_neural_commitment_chars", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "text_to_speech_neural_overage_chars", ValueType: schema.Int64, DefaultValue: 0},

		// Connected container
		{Key: "connected_container_speech_to_text_standard_commitment_hrs", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "connected_container_speech_to_text_standard_overage_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "connected_container_speech_to_text_custom_commitment_hrs", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "connected_container_speech_to_text_custom_overage_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "connected_container_speech_to_text_enhanced_add_ons_commitment_hrs", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "connected_container_speech_to_text_enhanced_add_ons_overage_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "connected_container_text_to_speech_neural_commitment_chars", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "connected_container_text_to_speech_neural_overage_chars", ValueType: schema.Int64, DefaultValue: 0},
	}
}

// PopulateUsage parses the u schema.UsageData into the CognitiveAccountSpeech.
// It uses the `infracost_usage` struct tags to populate data into the CognitiveAccountSpeech.
func (r *CognitiveAccountSpeech) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid CognitiveAccountSpeech struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *CognitiveAccountSpeech) BuildResource() *schema.Resource {
	// F0 is the free tier
	if strings.EqualFold(r.Sku, "f0") {
		return &schema.Resource{
			Name:      r.Address,
			NoPrice:   true,
			IsSkipped: true,
		}
	}

	// For some reason the SKU can either be S0 or S1 but they map to the same thing
	if !strings.EqualFold(r.Sku, "s0") && !strings.EqualFold(r.Sku, "s1") {
		logging.Logger.Warn().Msgf("Unsupported SKU %s for %s", r.Sku, r.Address)
		return nil
	}

	costComponents := r.costComponents()

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *CognitiveAccountSpeech) costComponents() []*schema.CostComponent {
	costComponents := make([]*schema.CostComponent, 0)

	// Speech to text
	if r.SpeechToTextCommitmentHrs != nil {
		costComponents = append(costComponents, r.commitmentTierHourlyCostComponents("Speech to text", standardCommitmentTier, "Commitment Tier Speech to Text Azure", *r.SpeechToTextCommitmentHrs, floatPtrToDecimalPtr(r.SpeechToTextOverageHrs))...)
	}
	if r.ConnectedContainerSpeechToTextCommitmentHrs != nil {
		costComponents = append(costComponents, r.commitmentTierHourlyCostComponents("Speech to text", connectedContainerCommitmentTier, "Commitment Tier Speech to Text Connected", *r.ConnectedContainerSpeechToTextCommitmentHrs, floatPtrToDecimalPtr(r.ConnectedContainerSpeechToTextOverageHrs))...)
	}
	if (r.SpeechToTextCommitmentHrs == nil && r.ConnectedContainerSpeechToTextCommitmentHrs == nil) || r.SpeechToTextStandardMonthlyHrs != nil {
		costComponents = append(costComponents, r.s0HourlyCostComponent("Speech to text", "Speech To Text", r.SpeechToTextStandardMonthlyHrs))
	}

	costComponents = append(costComponents,
		r.s0HourlyCostComponent("Speech to text batch", "Speech to Text Batch", r.SpeechToTextBatchMonthlyHrs),
	)

	if r.SpeechToTextCustomModelCommitmentHrs != nil {
		costComponents = append(costComponents, r.commitmentTierHourlyCostComponents("Speech to text custom model", standardCommitmentTier, "Commitment Tier Custom Speech to Text Azure", *r.SpeechToTextCustomModelCommitmentHrs, floatPtrToDecimalPtr(r.SpeechToTextCustomModelOverageHrs))...)
	}
	if r.ConnectedContainerSpeechToTextCustomModelCommitmentHrs != nil {
		costComponents = append(costComponents, r.commitmentTierHourlyCostComponents("Speech to text custom model", connectedContainerCommitmentTier, "Commitment Tier Custom Speech to Text Connected", *r.ConnectedContainerSpeechToTextCustomModelCommitmentHrs, floatPtrToDecimalPtr(r.ConnectedContainerSpeechToTextCustomModelOverageHrs))...)
	}
	if (r.SpeechToTextCustomModelCommitmentHrs == nil && r.ConnectedContainerSpeechToTextCustomModelCommitmentHrs == nil) || r.SpeechToTextCustomModelMonthlyHrs != nil {
		costComponents = append(costComponents, r.s0HourlyCostComponent("Speech to text custom model", "Custom Speech To Text", r.SpeechToTextCustomModelMonthlyHrs))
	}

	costComponents = append(costComponents, []*schema.CostComponent{
		r.s0HourlyCostComponent("Speech to text custom model batch", "Custom Speech to Text Batch", r.SpeechToTextCustomModelBatchMonthlyHrs),
		r.s0CostComponent("Speech to text custom endpoint hosting", "Custom Speech Model Hosting Unit", floatPtrToDecimalPtr(r.SpeechToTextCustomEndpointMonthlyHrs), "hours", 1, &schema.PriceFilter{Unit: strPtr("1/Hour")}), // This uses a different unit, so we don't want to filter the price by '1 Hour'
		r.s0HourlyCostComponent("Speech to text custom training", "Custom Speech Training", r.SpeechToTextCustomTrainingMonthlyHrs),
	}...)

	if r.SpeechToTextEnhancedAddOnsCommitmentHrs != nil {
		costComponents = append(costComponents, r.commitmentTierHourlyCostComponents("Speech to text enhanced add-ons", standardCommitmentTier, "Commitment Tier STT AddOn Azure", *r.SpeechToTextEnhancedAddOnsCommitmentHrs, floatPtrToDecimalPtr(r.SpeechToTextEnhancedAddOnsOverageHrs))...)
	}
	if r.ConnectedContainerSpeechToTextEnhancedAddOnsCommitmentHrs != nil {
		costComponents = append(costComponents, r.commitmentTierHourlyCostComponents("Speech to text enhanced add-ons", connectedContainerCommitmentTier, "Commitment Tier STT AddOn Connected", *r.ConnectedContainerSpeechToTextEnhancedAddOnsCommitmentHrs, floatPtrToDecimalPtr(r.ConnectedContainerSpeechToTextEnhancedAddOnsOverageHrs))...)
	}
	if (r.SpeechToTextEnhancedAddOnsCommitmentHrs == nil && r.ConnectedContainerSpeechToTextEnhancedAddOnsCommitmentHrs == nil) || r.SpeechToTextEnhancedAddOnsMonthlyHrs != nil {
		costComponents = append(costComponents, r.s0HourlyCostComponent("Speech to text enhanced add-ons", "S1 Speech to Text Enhanced Feature Audio", r.SpeechToTextEnhancedAddOnsMonthlyHrs))
	}

	costComponents = append(costComponents,
		r.s0HourlyCostComponent("Speech to text conversation transcription multi-channel audio", "Conversation Transcription Multichannel Audio", r.SpeechToTextConversationTranscriptionMultiChannelAudioMonthlyHrs),
	)

	// Text to speech
	if r.TextToSpeechNeuralCommitmentChars != nil {
		if !containsInt64(validCommitmentTierChars, *r.TextToSpeechNeuralCommitmentChars) {
			logging.Logger.Warn().Msgf("Invalid commitment tier amount %d for %s", *r.TextToSpeechNeuralCommitmentChars, r.Address)
		} else {
			costComponents = append(costComponents, r.commitmentTierCostComponents("Text to speech neural", standardCommitmentTier, "Commitment Tier Neural Text to Speech Azure", *r.TextToSpeechNeuralCommitmentChars, intPtrToDecimalPtr(r.TextToSpeechNeuralOverageChars), "1M chars", 1_000_000)...)
		}
	}
	if r.ConnectedContainerTextToSpeechNeuralCommitmentChars != nil {
		if !containsInt64(validCommitmentTierChars, *r.ConnectedContainerTextToSpeechNeuralCommitmentChars) {
			logging.Logger.Warn().Msgf("Invalid commitment tier amount %d for %s", *r.ConnectedContainerTextToSpeechNeuralCommitmentChars, r.Address)
		} else {
			costComponents = append(costComponents, r.commitmentTierCostComponents("Text to speech neural", connectedContainerCommitmentTier, "Commitment Tier Neural Text to Speech Connected", *r.ConnectedContainerTextToSpeechNeuralCommitmentChars, intPtrToDecimalPtr(r.ConnectedContainerTextToSpeechNeuralOverageChars), "1M chars", 1_000_000)...)
		}
	}
	if (r.TextToSpeechNeuralCommitmentChars == nil && r.ConnectedContainerTextToSpeechNeuralCommitmentChars == nil) || r.TextToSpeechNeuralChars != nil {
		costComponents = append(costComponents, r.s0CostComponent("Text to speech neural", "Neural Text To Speech Characters", intPtrToDecimalPtr(r.TextToSpeechNeuralChars), "1M chars", 1_000_000, nil))
	}

	costComponents = append(costComponents, []*schema.CostComponent{
		r.s0HourlyCostComponent("Text to speech custom neural training", "Custom Neural Training", r.TextToSpeechCustomNeuralTrainingMonthlyHrs),
		r.s0CostComponent("Text to speech custom neural", "Custom Neural Realtime Characters", intPtrToDecimalPtr(r.TextToSpeechCustomNeuralChars), "1M chars", 1_000_000, nil),
		r.s0CostComponent("Text to speech custom neural endpoint hosting", "Custom Neural Voice Model Hosting Unit", floatPtrToDecimalPtr(r.TextToSpeechCustomNeuralEndpointMonthlyHrs), "hours", 1, &schema.PriceFilter{Unit: strPtr("1/Hour")}), // This uses a different unit, so we don't want to filter the price by '1 Hour'
		r.s0CostComponent("Text to speech long audio", "Neural Long Audio Characters", intPtrToDecimalPtr(r.TextToSpeechLongAudioChars), "1M chars", 1_000_000, nil),
	}...)

	costComponents = append(costComponents, r.personalVoiceCostComponents()...)

	// Speech translation

	costComponents = append(costComponents,
		r.s0HourlyCostComponent("Speech translation", "Speech Translation", r.SpeechTranslationMonthlyHrs),
	)

	// Speaker recognition
	costComponents = append(costComponents, []*schema.CostComponent{
		r.s0CostComponent("Speaker verification", "Speaker Verification Transactions", intPtrToDecimalPtr(r.SpeakerVerificationTransactions), "1K transactions", 1_000, nil),
		r.s0CostComponent("Speaker identification", "Speaker Identification Transactions", intPtrToDecimalPtr(r.SpeakerIdentificationTransactions), "1K transactions", 1_000, nil),
	}...)

	// Voice profiles
	costComponents = append(costComponents,
		r.s0CostComponent("Voice profiles", "Voice Storage", intPtrToDecimalPtr(r.VoiceProfiles), "1K profiles", 1_000, nil),
	)

	return costComponents
}

func (r *CognitiveAccountSpeech) s0CostComponent(name string, meterName string, qty *decimal.Decimal, unit string, qtyDiv int64, priceFilter *schema.PriceFilter) *schema.CostComponent {
	if qty != nil {
		qty = decimalPtr(qty.Div(decimal.NewFromInt(qtyDiv)))
	}

	return &schema.CostComponent{
		Name:            name,
		Unit:            unit,
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr(vendorName),
			Region:        strPtr(r.Region),
			Service:       strPtr("Cognitive Services"),
			ProductFamily: strPtr("AI + Machine Learning"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Speech")},
				{Key: "skuName", Value: strPtr("S1")},
				{Key: "meterName", Value: strPtr(meterName)},
			},
		},
		PriceFilter: priceFilter,
	}
}

func (r *CognitiveAccountSpeech) s0HourlyCostComponent(name string, meterName string, hours *float64) *schema.CostComponent {
	var qty *decimal.Decimal
	if hours != nil {
		qty = decimalPtr(decimal.NewFromFloat(*hours))
	}

	return r.s0CostComponent(name, meterName, qty, "hours", 1, &schema.PriceFilter{Unit: strPtr("1 Hour")})
}

func (r *CognitiveAccountSpeech) personalVoiceCostComponents() []*schema.CostComponent {
	var profilesQty *decimal.Decimal
	if r.TextToSpeechPersonalVoiceProfiles != nil {
		profilesQty = decimalPtr(decimal.NewFromInt(*r.TextToSpeechPersonalVoiceProfiles).Div(decimal.NewFromInt(1_000)))
	}

	var charsQty *decimal.Decimal
	if r.TextToSpeechPersonalVoiceChars != nil {
		charsQty = decimalPtr(decimal.NewFromInt(*r.TextToSpeechPersonalVoiceChars).Div(decimal.NewFromInt(1_000_000)))
	}

	return []*schema.CostComponent{
		{
			Name:            "Text to speech personal voice profiles",
			Unit:            "1K profiles",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: profilesQty,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Speech")},
					{Key: "skuName", Value: strPtr("Text to Speech - Personal Voice")},
					{Key: "meterName", Value: strPtr("Voice Storage")},
				},
			},
		},
		{
			Name:            "Text to speech personal voice characters",
			Unit:            "1M chars",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: charsQty,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Speech")},
					{Key: "skuName", Value: strPtr("Text to Speech - Personal Voice")},
					{Key: "meterName", Value: strPtr("Text to Speech - Personal Voice Characters")},
				},
			},
		},
	}
}

func (r *CognitiveAccountSpeech) commitmentTierCostComponents(namePrefix string, commitmentTierType int, skuNamePrefix string, commitedAmount int64, overage *decimal.Decimal, unit string, qtyDiv int64) []*schema.CostComponent {
	desc := amountToDescription(commitedAmount)

	skuName := skuNamePrefix + " " + desc

	// Convert the amount to the correct unit
	// If the amount is 2_000, the qty divider is 1_000, so the qty should be 2
	// so we show the correct qty based on the units
	qty := decimal.NewFromInt(commitedAmount).Div(decimal.NewFromInt(qtyDiv))

	suffix := " (commitment)"
	if commitmentTierType == connectedContainerCommitmentTier {
		suffix = " (connected container commitment)"
	}

	costComponents := []*schema.CostComponent{
		{
			Name: fmt.Sprintf("%s%s", namePrefix, suffix),
			Unit: unit,
			// Use a monthly quantity of 1 and a unit multiplier so we show the
			// correct number of hours in the qty field, and divide the price
			// by the number of hours in the commitment tier
			UnitMultiplier:  decimal.NewFromInt(1).Div(qty),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Speech")},
					{Key: "skuName", Value: strPtr(skuName)},
					{Key: "meterName", ValueRegex: regexPtr("Unit$")},
				},
			},
		},
	}

	if overage != nil && overage.GreaterThan(decimal.Zero) {
		overageQty := decimalPtr(overage.Div(decimal.NewFromInt(qtyDiv)))

		suffix := " (overage)"
		if commitmentTierType == connectedContainerCommitmentTier {
			suffix = " (connected container overage)"
		}

		costComponents = append(costComponents, &schema.CostComponent{
			Name:            fmt.Sprintf("%s%s", namePrefix, suffix),
			Unit:            unit,
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: overageQty,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Speech")},
					{Key: "skuName", Value: strPtr(skuName)},
					{Key: "meterName", ValueRegex: regexPtr("Overage")},
				},
			},
		})
	}

	return costComponents
}

func (r *CognitiveAccountSpeech) commitmentTierHourlyCostComponents(namePrefix string, commitmentTierType int, skuNamePrefix string, amount int64, overage *decimal.Decimal) []*schema.CostComponent {
	if !containsInt64(validCommitmentTierHrs, amount) {
		logging.Logger.Warn().Msgf("Invalid commitment tier amount %d for %s", amount, r.Address)
		return []*schema.CostComponent{}
	}

	return r.commitmentTierCostComponents(namePrefix, commitmentTierType, skuNamePrefix, amount, overage, "hours", 1)
}

// amountToDescription converts an amount to a human readable description.
// Such as 20K or 2000M.
func amountToDescription(amount int64) string {
	if amount < 1000 {
		return fmt.Sprintf("%d", amount)
	}

	if amount < 1000000 {
		return fmt.Sprintf("%dK", amount/1_000)
	}

	return fmt.Sprintf("%dM", amount/1_000_000)
}
