package azure

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

var validSpeechCommitmentTierHrs = []int64{2_000, 10_000, 50_000}
var validSpeechCommitmentTierChars = []int64{80_000_000, 400_000_000, 2_000_000_000}

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

	Sku string

	// Usage attributes

	// Speech to text
	MonthlySpeechToTextStandardHrs                                   *float64 `infracost_usage:"monthly_speech_to_text_standard_hrs"`
	MonthlySpeechToTextBatchHrs                                      *float64 `infracost_usage:"monthly_speech_to_text_standard_batch_hrs"`
	MonthlySpeechToTextCustomModelHrs                                *float64 `infracost_usage:"monthly_speech_to_text_custom_hrs"`
	MonthlySpeechToTextCustomModelBatchHrs                           *float64 `infracost_usage:"monthly_speech_to_text_custom_batch_hrs"`
	MonthlySpeechToTextCustomEndpointHrs                             *float64 `infracost_usage:"monthly_speech_to_text_custom_endpoint_hrs"`
	MonthlySpeechToTextConversationTranscriptionMultiChannelAudioHrs *float64 `infracost_usage:"monthly_speech_to_text_conversation_transcription_multi_channel_audio_hrs"`
	MonthlySpeechToTextCustomTrainingHrs                             *float64 `infracost_usage:"monthly_speech_to_text_custom_training_hrs"`
	MonthlySpeechToTextEnhancedAddOnsHrs                             *float64 `infracost_usage:"monthly_speech_to_text_enhanced_add_ons_hrs"`

	// Text to speech
	MonthlyTextToSpeechNeuralChars             *int64   `infracost_usage:"monthly_text_to_speech_neural_chars"`
	MonthlyTextToSpeechCustomNeuralTrainingHrs *float64 `infracost_usage:"monthly_text_to_speech_custom_neural_training_hrs"`
	MonthlyTextToSpeechCustomNeuralChars       *int64   `infracost_usage:"monthly_text_to_speech_custom_neural_chars"`
	MonthlyTextToSpeechCustomNeuralEndpointHrs *float64 `infracost_usage:"monthly_text_to_speech_custom_neural_endpoint_hrs"`
	MonthlyTextToSpeechLongAudioChars          *int64   `infracost_usage:"monthly_text_to_speech_long_audio_chars"`
	MonthlyTextToSpeechPersonalVoiceProfiles   *int64   `infracost_usage:"monthly_text_to_speech_personal_voice_profiles"`
	MonthlyTextToSpeechPersonalVoiceChars      *int64   `infracost_usage:"monthly_text_to_speech_personal_voice_chars"`

	// Speech translation
	MonthlySpeechTranslationHrs *float64 `infracost_usage:"monthly_speech_translation_hrs"`

	// Speaker recognition
	MonthlySpeakerVerificationTransactions   *int64 `infracost_usage:"monthly_speaker_verification_transactions"`
	MonthlySpeakerIdentificationTransactions *int64 `infracost_usage:"monthly_speaker_identification_transactions"`

	// Voice storage
	MonthlyVoiceProfiles *int64 `infracost_usage:"monthly_voice_profiles"`

	// Standard commitment tier
	MonthlyCommitmentSpeechToTextHrs                      *int64   `infracost_usage:"monthly_commitment_speech_to_text_standard_hrs"`
	MonthlyCommitmentSpeechToTextOverageHrs               *float64 `infracost_usage:"monthly_commitment_speech_to_text_standard_overage_hrs"`
	MonthlyCommitmentSpeechToTextCustomModelHrs           *int64   `infracost_usage:"monthly_commitment_speech_to_text_custom_hrs"`
	MonthlyCommitmentSpeechToTextCustomModelOverageHrs    *float64 `infracost_usage:"monthly_commitment_speech_to_text_custom_overage_hrs"`
	MonthlyCommitmentSpeechToTextEnhancedAddOnsHrs        *int64   `infracost_usage:"monthly_commitment_speech_to_text_enhanced_add_ons_hrs"`
	MonthlyCommitmentSpeechToTextEnhancedAddOnsOverageHrs *float64 `infracost_usage:"monthly_commitment_speech_to_text_enhanced_add_ons_overage_hrs"`
	MonthlyCommitmentTextToSpeechNeuralCommitmentChars    *int64   `infracost_usage:"monthly_commitment_text_to_speech_neural_commitment_chars"`
	MonthlyCommitmentTextToSpeechNeuralOverageChars       *int64   `infracost_usage:"monthly_commitment_text_to_speech_neural_overage_chars"`

	// Connected container
	MonthlyConnectedContainerCommitmentSpeechToTextHrs                      *int64   `infracost_usage:"monthly_connected_container_commitment_speech_to_text_standard_hrs"`
	MonthlyConnectedContainerCommitmentSpeechToTextOverageHrs               *float64 `infracost_usage:"monthly_connected_container_commitment_speech_to_text_standard_overage_hrs"`
	MonthlyConnectedContainerCommitmentSpeechToTextCustomModelHrs           *int64   `infracost_usage:"monthly_connected_container_commitment_speech_to_text_custom_hrs"`
	MonthlyConnectedContainerCommitmentSpeechToTextCustomModelOverageHrs    *float64 `infracost_usage:"monthly_connected_container_commitment_speech_to_text_custom_overage_hrs"`
	MonthlyConnectedContainerCommitmentSpeechToTextEnhancedAddOnsHrs        *int64   `infracost_usage:"monthly_connected_container_commitment_speech_to_text_enhanced_add_ons_hrs"`
	MonthlyConnectedContainerCommitmentSpeechToTextEnhancedAddOnsOverageHrs *float64 `infracost_usage:"monthly_connected_container_commitment_speech_to_text_enhanced_add_ons_overage_hrs"`
	MonthlyConnectedContainerCommitmentTextToSpeechNeuralCommitmentChars    *int64   `infracost_usage:"monthly_connected_container_commitment_text_to_speech_neural_commitment_chars"`
	MonthlyConnectedContainerCommitmentTextToSpeechNeuralOverageChars       *int64   `infracost_usage:"monthly_connected_container_commitment_text_to_speech_neural_overage_chars"`
}

// // CoreType returns the name of this resource type
func (r *CognitiveAccountSpeech) CoreType() string {
	return "CognitiveAccountSpeech"
}

// UsageSchema defines a list which represents the usage schema of CognitiveAccountSpeech.
func (r *CognitiveAccountSpeech) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		// Speech to text
		{Key: "monthly_speech_to_text_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_speech_to_text_custom_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_speech_to_text_batch_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_speech_to_text_custom_batch_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_speech_to_text_custom_endpoint_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_speech_to_text_conversation_transcription_multi_channel_audio_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_speech_to_text_custom_training_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_speech_to_text_enhanced_add_ons_hrs", ValueType: schema.Float64, DefaultValue: 0},
		// Text to speech
		{Key: "monthly_text_to_speech_neural_chars", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_text_to_speech_custom_neural_training_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_text_to_speech_custom_neural_chars", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_text_to_speech_custom_neural_endpoint_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_text_to_speech_long_audio_chars", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_text_to_speech_personal_voice_profiles", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_text_to_speech_personal_voice_chars", ValueType: schema.Int64, DefaultValue: 0},
		// Speech translation
		{Key: "monthly_speech_translation_hrs", ValueType: schema.Float64, DefaultValue: 0},
		// Speaker recognition
		{Key: "monthly_speaker_verification_transactions", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_speaker_identification_transactions", ValueType: schema.Int64, DefaultValue: 0},
		// Voice storage
		{Key: "monthly_voice_profiles", ValueType: schema.Int64, DefaultValue: 0},

		// Standard commitment tier
		{Key: "monthly_commitment_speech_to_text_standard_hrs", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_commitment_speech_to_text_standard_overage_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_commitment_speech_to_text_custom_hrs", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_commitment_speech_to_text_custom_overage_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_commitment_speech_to_text_enhanced_add_ons_hrs", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_commitment_speech_to_text_enhanced_add_ons_overage_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_commitment_text_to_speech_neural_commitment_chars", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_commitment_text_to_speech_neural_overage_chars", ValueType: schema.Int64, DefaultValue: 0},

		// Connected container
		{Key: "monthly_connected_container_commitment_speech_to_text_standard_hrs", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_connected_container_commitment_speech_to_text_standard_overage_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_connected_container_commitment_speech_to_text_custom_hrs", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_connected_container_commitment_speech_to_text_custom_overage_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_connected_container_commitment_speech_to_text_enhanced_add_ons_hrs", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_connected_container_commitment_speech_to_text_enhanced_add_ons_overage_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_connected_container_commitment_text_to_speech_neural_commitment_chars", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_connected_container_commitment_text_to_speech_neural_overage_chars", ValueType: schema.Int64, DefaultValue: 0},
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
	if !strings.EqualFold(r.Sku, "s0") && !strings.EqualFold(r.Sku, "s1") && !strings.EqualFold(r.Sku, "s") {
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
	if r.MonthlyCommitmentSpeechToTextHrs != nil {
		costComponents = append(costComponents, r.commitmentTierHourlyCostComponents("Speech to text", standardCommitmentTier, "Commitment Tier Speech to Text Azure", *r.MonthlyCommitmentSpeechToTextHrs, floatPtrToDecimalPtr(r.MonthlyCommitmentSpeechToTextOverageHrs))...)
	}
	if r.MonthlyConnectedContainerCommitmentSpeechToTextHrs != nil {
		costComponents = append(costComponents, r.commitmentTierHourlyCostComponents("Speech to text", connectedContainerCommitmentTier, "Commitment Tier Speech to Text Connected", *r.MonthlyConnectedContainerCommitmentSpeechToTextHrs, floatPtrToDecimalPtr(r.MonthlyConnectedContainerCommitmentSpeechToTextOverageHrs))...)
	}
	if (r.MonthlyCommitmentSpeechToTextHrs == nil && r.MonthlyConnectedContainerCommitmentSpeechToTextHrs == nil) || r.MonthlySpeechToTextStandardHrs != nil {
		costComponents = append(costComponents, r.s0HourlyCostComponent("Speech to text", "S1 Speech To Text", r.MonthlySpeechToTextStandardHrs))
	}

	costComponents = append(costComponents,
		r.s0HourlyCostComponent("Speech to text batch", "S1 Speech to Text Batch", r.MonthlySpeechToTextBatchHrs),
	)

	if r.MonthlyCommitmentSpeechToTextCustomModelHrs != nil {
		costComponents = append(costComponents, r.commitmentTierHourlyCostComponents("Speech to text custom model", standardCommitmentTier, "Commitment Tier Custom Speech to Text Azure", *r.MonthlyCommitmentSpeechToTextCustomModelHrs, floatPtrToDecimalPtr(r.MonthlyCommitmentSpeechToTextCustomModelOverageHrs))...)
	}
	if r.MonthlyConnectedContainerCommitmentSpeechToTextCustomModelHrs != nil {
		costComponents = append(costComponents, r.commitmentTierHourlyCostComponents("Speech to text custom model", connectedContainerCommitmentTier, "Commitment Tier Custom Speech to Text Connected", *r.MonthlyConnectedContainerCommitmentSpeechToTextCustomModelHrs, floatPtrToDecimalPtr(r.MonthlyConnectedContainerCommitmentSpeechToTextCustomModelOverageHrs))...)
	}
	if (r.MonthlyCommitmentSpeechToTextCustomModelHrs == nil && r.MonthlyConnectedContainerCommitmentSpeechToTextCustomModelHrs == nil) || r.MonthlySpeechToTextCustomModelHrs != nil {
		costComponents = append(costComponents, r.s0HourlyCostComponent("Speech to text custom model", "S1 Custom Speech To Text", r.MonthlySpeechToTextCustomModelHrs))
	}

	costComponents = append(costComponents, []*schema.CostComponent{
		r.s0HourlyCostComponent("Speech to text custom model batch", "S1 Custom Speech to Text Batch", r.MonthlySpeechToTextCustomModelBatchHrs),
		r.s0CostComponent("Speech to text custom endpoint hosting", "S1 Custom Speech Model Hosting Unit", floatPtrToDecimalPtr(r.MonthlySpeechToTextCustomEndpointHrs), "hours", 1, &schema.PriceFilter{Unit: strPtr("1/Hour")}), // This uses a different unit, so we don't want to filter the price by '1 Hour'
		r.s0HourlyCostComponent("Speech to text custom training", "S1 Custom Speech Training", r.MonthlySpeechToTextCustomTrainingHrs),
	}...)

	if r.MonthlyCommitmentSpeechToTextEnhancedAddOnsHrs != nil {
		costComponents = append(costComponents, r.commitmentTierHourlyCostComponents("Speech to text enhanced add-ons", standardCommitmentTier, "Commitment Tier STT AddOn Azure", *r.MonthlyCommitmentSpeechToTextEnhancedAddOnsHrs, floatPtrToDecimalPtr(r.MonthlyCommitmentSpeechToTextEnhancedAddOnsOverageHrs))...)
	}
	if r.MonthlyConnectedContainerCommitmentSpeechToTextEnhancedAddOnsHrs != nil {
		costComponents = append(costComponents, r.commitmentTierHourlyCostComponents("Speech to text enhanced add-ons", connectedContainerCommitmentTier, "Commitment Tier STT AddOn Connected", *r.MonthlyConnectedContainerCommitmentSpeechToTextEnhancedAddOnsHrs, floatPtrToDecimalPtr(r.MonthlyConnectedContainerCommitmentSpeechToTextEnhancedAddOnsOverageHrs))...)
	}
	if (r.MonthlyCommitmentSpeechToTextEnhancedAddOnsHrs == nil && r.MonthlyConnectedContainerCommitmentSpeechToTextEnhancedAddOnsHrs == nil) || r.MonthlySpeechToTextEnhancedAddOnsHrs != nil {
		costComponents = append(costComponents, r.s0HourlyCostComponent("Speech to text enhanced add-ons", "S1 Speech to Text Enhanced Feature Audio", r.MonthlySpeechToTextEnhancedAddOnsHrs))
	}

	costComponents = append(costComponents,
		r.s0HourlyCostComponent("Speech to text conversation transcription multi-channel audio", "S1 Conversation Transcription Multichannel Audio", r.MonthlySpeechToTextConversationTranscriptionMultiChannelAudioHrs),
	)

	// Text to speech
	if r.MonthlyCommitmentTextToSpeechNeuralCommitmentChars != nil {
		if !containsInt64(validSpeechCommitmentTierChars, *r.MonthlyCommitmentTextToSpeechNeuralCommitmentChars) {
			logging.Logger.Warn().Msgf("Invalid commitment tier amount %d for %s", *r.MonthlyCommitmentTextToSpeechNeuralCommitmentChars, r.Address)
		} else {
			costComponents = append(costComponents, r.commitmentTierCostComponents("Text to speech neural", standardCommitmentTier, "Commitment Tier Neural Text to Speech Azure", *r.MonthlyCommitmentTextToSpeechNeuralCommitmentChars, intPtrToDecimalPtr(r.MonthlyCommitmentTextToSpeechNeuralOverageChars), "1M chars", 1_000_000)...)
		}
	}
	if r.MonthlyConnectedContainerCommitmentTextToSpeechNeuralCommitmentChars != nil {
		if !containsInt64(validSpeechCommitmentTierChars, *r.MonthlyConnectedContainerCommitmentTextToSpeechNeuralCommitmentChars) {
			logging.Logger.Warn().Msgf("Invalid commitment tier amount %d for %s", *r.MonthlyConnectedContainerCommitmentTextToSpeechNeuralCommitmentChars, r.Address)
		} else {
			costComponents = append(costComponents, r.commitmentTierCostComponents("Text to speech neural", connectedContainerCommitmentTier, "Commit Tier Neural TTS Connected", *r.MonthlyConnectedContainerCommitmentTextToSpeechNeuralCommitmentChars, intPtrToDecimalPtr(r.MonthlyConnectedContainerCommitmentTextToSpeechNeuralOverageChars), "1M chars", 1_000_000)...)
		}
	}
	if (r.MonthlyCommitmentTextToSpeechNeuralCommitmentChars == nil && r.MonthlyConnectedContainerCommitmentTextToSpeechNeuralCommitmentChars == nil) || r.MonthlyTextToSpeechNeuralChars != nil {
		costComponents = append(costComponents, r.s0CostComponent("Text to speech neural", "S1 Neural Text To Speech Characters", intPtrToDecimalPtr(r.MonthlyTextToSpeechNeuralChars), "1M chars", 1_000_000, nil))
	}

	costComponents = append(costComponents, []*schema.CostComponent{
		r.s0HourlyCostComponent("Text to speech custom neural training", "S1 Custom Neural Training", r.MonthlyTextToSpeechCustomNeuralTrainingHrs),
		r.s0CostComponent("Text to speech custom neural", "S1 Custom Neural Realtime Characters", intPtrToDecimalPtr(r.MonthlyTextToSpeechCustomNeuralChars), "1M chars", 1_000_000, nil),
		r.s0CostComponent("Text to speech custom neural endpoint hosting", "S1 Custom Neural Voice Model Hosting Unit", floatPtrToDecimalPtr(r.MonthlyTextToSpeechCustomNeuralEndpointHrs), "hours", 1, &schema.PriceFilter{Unit: strPtr("1/Hour")}), // This uses a different unit, so we don't want to filter the price by '1 Hour'
		r.s0CostComponent("Text to speech long audio", "S1 Neural Long Audio Characters", intPtrToDecimalPtr(r.MonthlyTextToSpeechLongAudioChars), "1M chars", 1_000_000, nil),
	}...)

	costComponents = append(costComponents, r.personalVoiceCostComponents()...)

	// Speech translation

	costComponents = append(costComponents,
		r.s0HourlyCostComponent("Speech translation", "S1 Speech Translation", r.MonthlySpeechTranslationHrs),
	)

	// Speaker recognition
	costComponents = append(costComponents, []*schema.CostComponent{
		r.s0CostComponent("Speaker verification", "S1 Speaker Verification Transactions", intPtrToDecimalPtr(r.MonthlySpeakerVerificationTransactions), "1K transactions", 1_000, nil),
		r.s0CostComponent("Speaker identification", "S1 Speaker Identification Transactions", intPtrToDecimalPtr(r.MonthlySpeakerIdentificationTransactions), "1K transactions", 1_000, nil),
	}...)

	// Voice profiles
	costComponents = append(costComponents,
		r.s0CostComponent("Voice profiles", "S1 Voice Storage", intPtrToDecimalPtr(r.MonthlyVoiceProfiles), "1K profiles", 1_000, nil),
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
	if r.MonthlyTextToSpeechPersonalVoiceProfiles != nil {
		profilesQty = decimalPtr(decimal.NewFromInt(*r.MonthlyTextToSpeechPersonalVoiceProfiles).Div(decimal.NewFromInt(1_000)))
	}

	var charsQty *decimal.Decimal
	if r.MonthlyTextToSpeechPersonalVoiceChars != nil {
		charsQty = decimalPtr(decimal.NewFromInt(*r.MonthlyTextToSpeechPersonalVoiceChars).Div(decimal.NewFromInt(1_000_000)))
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
					{Key: "meterName", Value: strPtr("Text to Speech - Personal Voice Voice Storage")},
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
			// correct number of hours/chars in the qty field, and divide the price
			// by the number of hours/chars in the commitment tier
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
	if !containsInt64(validSpeechCommitmentTierHrs, amount) {
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
