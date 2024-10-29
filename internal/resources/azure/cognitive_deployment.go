package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

var languageModelSKUs = map[string]string{
	"gpt-35-turbo":          "Az-GPT35-Turbo-16K-0125",
	"gpt-35-turbo-16k":      "Az-GPT35-Turbo-16K-0125",
	"gpt-35-turbo-instruct": "Az-GPT35-Turbo-Instruct",
	"gpt-4":                 "Az-GPT4-8K",
	"gpt-4-32k":             "Az-GPT4-32K",
}

var languageModelGPT4Versions = map[string]string{
	"1106-preview":   "Az-GPT4-Turbo-128K",
	"0125-preview":   "Az-GPT4-Turbo-128K",
	"vision-preview": "Az-GPT4-Turbo-Vision-128K",
}

// The same skus and pricing applies to all models that support the assistant API
var assistantModels = map[string]struct{}{
	"gpt-35-turbo": {},
	"gpt-4":        {},
}

var baseModelSKUs = map[string]string{
	"babbage-002": "Babbage",
	"davinci-002": "Davinci",
}

var fineTuningSKUs = map[string]string{
	"babbage-002":      "Az-Babbage-002",
	"davinci-002":      "Az-Davinci-002",
	"gpt-35-turbo":     "Az-GPT35-Turbo-4K",
	"gpt-35-turbo-16k": "Az-GPT35-Turbo-16K",
}

var imageSKUs = map[string]string{
	"dall-e-2": "Az-Image-DALL-E",
	"dall-e-3": "Az-Image-Dall-E-3",
}

var embeddingSKUs = map[string]string{
	"text-embedding-ada-002": "Az-Embeddings-Ada",
	"text-embedding-3-small": "Az-Text-Embedding-3-Small",
	"text-embedding-3-large": "Az-Text-Embedding-3-Large",
}

var speechSKUs = map[string]string{
	"whisper": "Az-Speech-Whisper",
	"tts":     "Az-Speech-Text to Speech",
	"tts-hd":  "Az-Speech-Text to Speech HD",
}

// CognitiveDeployment struct represents an Azure OpenAI Deployment.
//
// Since the availability of models is very different across different regions we ignore any cost components
// that we don't have a price for. This is done by setting the `IgnoreIfMissingPrice` field to true.
// See the following URL for more information on different model availability in different regions:
// https://learn.microsoft.com/en-us/azure/ai-services/openai/concepts/models#standard-deployment-model-availability
//
// This only supports Pay-As-You-Go pricing tier, currently since Azure doesn't provide pricing for their
// Provisioned Throughput Units.
//
// This also doesn't support some models that have been deprecated by Azure. See the below for information on those resources:
// https://learn.microsoft.com/en-us/azure/ai-services/openai/concepts/legacy-models
//
// Resource information: https://azure.microsoft.com/en-gb/products/ai-services/openai-service
// Pricing information: https://azure.microsoft.com/en-gb/pricing/details/cognitive-services/openai-service/
type CognitiveDeployment struct {
	Address string
	Region  string
	Model   string
	Version string
	Tier    string

	// Usage-based attributes
	MonthlyLanguageInputTokens     *int64   `infracost_usage:"monthly_language_input_tokens"`
	MonthlyLanguageOutputTokens    *int64   `infracost_usage:"monthly_language_output_tokens"`
	MonthlyCodeInterpreterSessions *int64   `infracost_usage:"monthly_code_interpreter_sessions"`
	MonthlyBaseModelTokens         *int64   `infracost_usage:"monthly_base_model_tokens"`
	MonthlyFineTuningTrainingHours *float64 `infracost_usage:"monthly_fine_tuning_training_hours"`
	MonthlyFineTuningHostingHours  *float64 `infracost_usage:"monthly_fine_tuning_hosting_hours"`
	MonthlyFineTuningInputTokens   *int64   `infracost_usage:"monthly_fine_tuning_input_tokens"`
	MonthlyFineTuningOutputTokens  *int64   `infracost_usage:"monthly_fine_tuning_output_tokens"`
	MonthlyStandard10241024Images  *int64   `infracost_usage:"monthly_standard_1024_1024_images"`
	MonthlyStandard10241792Images  *int64   `infracost_usage:"monthly_standard_1024_1792_images"`
	MonthlyHD10241024Images        *int64   `infracost_usage:"monthly_hd_1024_1024_images"`
	MonthlyHD10241792Images        *int64   `infracost_usage:"monthly_hd_1024_1792_images"`
	MonthlyTextEmbeddingTokens     *int64   `infracost_usage:"monthly_text_embedding_tokens"`
	MonthlyTextToSpeechCharacters  *int64   `infracost_usage:"monthly_text_to_speech_characters"`
	MonthlyTextToSpeechHours       *float64 `infracost_usage:"monthly_text_to_speech_hours"`
}

// CoreType returns the name of this resource type
func (r *CognitiveDeployment) CoreType() string {
	return "CognitiveDeployment"
}

// UsageSchema defines a list which represents the usage schema of CognitiveDeployment.
func (r *CognitiveDeployment) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_language_input_tokens", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_language_output_tokens", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_code_interpreter_sessions", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_base_model_tokens", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_fine_tuning_training_hours", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_fine_tuning_hosting_hours", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_fine_tuning_input_tokens", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_fine_tuning_output_tokens", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_standard_1024_1024_images", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_standard_1024_1792_images", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_hd_1024_1024_images", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_hd_1024_1792_images", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_text_embeddings", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_text_to_speech_characters", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_text_to_speech_hours", ValueType: schema.Float64, DefaultValue: 0},
	}
}

// PopulateUsage parses the u schema.UsageData into the CognitiveDeployment.
// It uses the `infracost_usage` struct tags to populate data into the CognitiveDeployment.
func (r *CognitiveDeployment) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid CognitiveDeployment struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *CognitiveDeployment) BuildResource() *schema.Resource {
	if strings.EqualFold(r.Tier, "free") {
		return &schema.Resource{
			Name:      r.Address,
			NoPrice:   true,
			IsSkipped: true,
		}
	}

	costComponents := make([]*schema.CostComponent, 0)

	if _, ok := languageModelSKUs[r.Model]; ok {
		costComponents = append(costComponents, r.languageCostComponents()...)
	}

	if _, ok := assistantModels[r.Model]; ok {
		costComponents = append(costComponents, r.assistantCostComponents()...)
	}

	if _, ok := baseModelSKUs[r.Model]; ok {
		costComponents = append(costComponents, r.baseModelCostComponents()...)
	}

	if _, ok := fineTuningSKUs[r.Model]; ok {
		costComponents = append(costComponents, r.fineTuningCostComponents()...)
	}

	if _, ok := imageSKUs[r.Model]; ok {
		costComponents = append(costComponents, r.imageCostComponents()...)
	}

	if _, ok := embeddingSKUs[r.Model]; ok {
		costComponents = append(costComponents, r.embeddingCostComponents()...)
	}

	if _, ok := speechSKUs[r.Model]; ok {
		costComponents = append(costComponents, r.speechCostComponents()...)
	}

	if len(costComponents) == 0 {
		logging.Logger.Warn().Msgf("Skipping resource %s. Model '%s' is not supported", r.Address, r.Model)
		return nil
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *CognitiveDeployment) languageCostComponents() []*schema.CostComponent {
	skuPrefix := languageModelSKUs[r.Model]
	modelName := r.Model
	if r.Model == "gpt-4" {
		if v, ok := languageModelGPT4Versions[r.Version]; ok {
			skuPrefix = v
			modelName = fmt.Sprintf("%s-%s", r.Model, r.Version)
		}
	}

	var inputSku, outputSku string
	if r.Model == "gpt-35-turbo" || r.Model == "gpt-35-turbo-16k" {
		inputSku = fmt.Sprintf("%s Input", skuPrefix)
		outputSku = fmt.Sprintf("%s Output", skuPrefix)
	} else if r.Model == "gpt-35-turbo-instruct" {
		inputSku = fmt.Sprintf("%s-Input", skuPrefix)
		outputSku = fmt.Sprintf("%s-Output", skuPrefix)
	} else {
		inputSku = fmt.Sprintf("%s-Prompt", skuPrefix)
		outputSku = fmt.Sprintf("%s-Completion", skuPrefix)
	}

	var inputQty, outputQty *decimal.Decimal
	if r.MonthlyLanguageInputTokens != nil {
		inputQty = decimalPtr(decimal.NewFromInt(*r.MonthlyLanguageInputTokens).Div(decimal.NewFromInt(1_000)))
	}
	if r.MonthlyLanguageOutputTokens != nil {
		outputQty = decimalPtr(decimal.NewFromInt(*r.MonthlyLanguageOutputTokens).Div(decimal.NewFromInt(1_000)))
	}

	return []*schema.CostComponent{
		{
			Name:                 fmt.Sprintf("Language input (%s)", modelName),
			Unit:                 "1K tokens",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      inputQty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr(inputSku)},
					{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Tokens", inputSku))},
				},
			},
		},
		{
			Name:                 fmt.Sprintf("Language output (%s)", modelName),
			Unit:                 "1K tokens",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      outputQty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr(outputSku)},
					{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Tokens", outputSku))},
				},
			},
		},
	}
}

func (r *CognitiveDeployment) assistantCostComponents() []*schema.CostComponent {
	var qty *decimal.Decimal
	if r.MonthlyCodeInterpreterSessions != nil {
		qty = decimalPtr(decimal.NewFromInt(*r.MonthlyCodeInterpreterSessions))
	}

	return []*schema.CostComponent{
		{
			Name:                 fmt.Sprintf("Code interpreter sessions (%s)", r.Model),
			Unit:                 "sessions",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      qty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr("Az-Assistants-Code-Interpreter")},
					{Key: "meterName", Value: strPtr("Az-Assistants-Code-Interpreter Session")},
				},
			},
		},
	}
}

func (r *CognitiveDeployment) baseModelCostComponents() []*schema.CostComponent {
	skuPrefix := baseModelSKUs[r.Model]

	var qty *decimal.Decimal
	if r.MonthlyBaseModelTokens != nil {
		qty = decimalPtr(decimal.NewFromInt(*r.MonthlyBaseModelTokens).Div(decimal.NewFromInt(1_000)))
	}

	return []*schema.CostComponent{
		{
			Name:                 fmt.Sprintf("Base model tokens (%s)", r.Model),
			Unit:                 "1K tokens",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      qty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr(fmt.Sprintf("%s - Base", skuPrefix))},
					{Key: "meterName", Value: strPtr(fmt.Sprintf("Text-%s Unit", skuPrefix))},
				},
			},
		},
	}
}

func (r *CognitiveDeployment) fineTuningCostComponents() []*schema.CostComponent {
	skuPrefix := fineTuningSKUs[r.Model]

	var trainingQty, hostingQty, inputQty, outputQty *decimal.Decimal
	if r.MonthlyFineTuningTrainingHours != nil {
		trainingQty = decimalPtr(decimal.NewFromFloat(*r.MonthlyFineTuningTrainingHours))
	}
	if r.MonthlyFineTuningHostingHours != nil {
		hostingQty = decimalPtr(decimal.NewFromFloat(*r.MonthlyFineTuningHostingHours))
	}
	if r.MonthlyFineTuningInputTokens != nil {
		inputQty = decimalPtr(decimal.NewFromInt(*r.MonthlyFineTuningInputTokens).Div(decimal.NewFromInt(1_000)))
	}
	if r.MonthlyFineTuningOutputTokens != nil {
		outputQty = decimalPtr(decimal.NewFromInt(*r.MonthlyFineTuningOutputTokens).Div(decimal.NewFromInt(1_000)))
	}

	return []*schema.CostComponent{
		{
			Name:                 fmt.Sprintf("Fine tuning training (%s)", r.Model),
			Unit:                 "hours",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      trainingQty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr(fmt.Sprintf("%s-FTuned", skuPrefix))},
					{Key: "meterName", Value: strPtr(fmt.Sprintf("%s-FTuned Training Unit", skuPrefix))},
				},
			},
		},
		{
			Name:                 fmt.Sprintf("Fine tuning hosting (%s)", r.Model),
			Unit:                 "hours",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      hostingQty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr(fmt.Sprintf("%s-FTuned", skuPrefix))},
					{Key: "meterName", Value: strPtr(fmt.Sprintf("%s-FTuned Deployment Hosting Unit", skuPrefix))},
				},
			},
		},
		{
			Name:                 fmt.Sprintf("Fine tuning input (%s)", r.Model),
			Unit:                 "1K tokens",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      inputQty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr(fmt.Sprintf("%s-Fine Tuned-Input", skuPrefix))},
					{Key: "meterName", Value: strPtr(fmt.Sprintf("%s-Fine Tuned-Input Tokens", skuPrefix))},
				},
			},
		},
		{
			Name:                 fmt.Sprintf("Fine tuning output (%s)", r.Model),
			Unit:                 "1K tokens",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      outputQty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr(fmt.Sprintf("%s-Fine Tuned-Output", skuPrefix))},
					{Key: "meterName", Value: strPtr(fmt.Sprintf("%s-Fine Tuned-Output Tokens", skuPrefix))},
				},
			},
		},
	}
}

func (r *CognitiveDeployment) imageCostComponents() []*schema.CostComponent {
	skuPrefix := imageSKUs[r.Model]

	var standard10241024Qty, standard10241792Qty, hd10241024Qty, hd10241792Qty *decimal.Decimal
	if r.MonthlyStandard10241024Images != nil {
		standard10241024Qty = decimalPtr(decimal.NewFromInt(*r.MonthlyStandard10241024Images).Div(decimal.NewFromInt(100)))
	}

	if r.Model == "dall-e-2" {
		return []*schema.CostComponent{
			{
				Name:                 fmt.Sprintf("Standard 1024x1024 images (%s)", r.Model),
				Unit:                 "100 images",
				UnitMultiplier:       decimal.NewFromInt(1),
				MonthlyQuantity:      standard10241024Qty,
				IgnoreIfMissingPrice: true,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr(vendorName),
					Region:        strPtr(r.Region),
					Service:       strPtr("Cognitive Services"),
					ProductFamily: strPtr("AI + Machine Learning"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "productName", Value: strPtr("Azure OpenAI")},
						{Key: "skuName", Value: strPtr(skuPrefix)},
						{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Images", skuPrefix))},
					},
				},
			},
		}
	}

	if r.MonthlyStandard10241792Images != nil {
		standard10241792Qty = decimalPtr(decimal.NewFromInt(*r.MonthlyStandard10241792Images).Div(decimal.NewFromInt(100)))
	}
	if r.MonthlyHD10241024Images != nil {
		hd10241024Qty = decimalPtr(decimal.NewFromInt(*r.MonthlyHD10241024Images).Div(decimal.NewFromInt(100)))
	}
	if r.MonthlyHD10241792Images != nil {
		hd10241792Qty = decimalPtr(decimal.NewFromInt(*r.MonthlyHD10241792Images).Div(decimal.NewFromInt(100)))
	}

	return []*schema.CostComponent{
		{
			Name:                 fmt.Sprintf("Standard 1024x1024 images (%s)", r.Model),
			Unit:                 "100 images",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      standard10241024Qty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr(fmt.Sprintf("%s Standard LowRes", skuPrefix))},
					{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Standard LowRes Images", skuPrefix))},
				},
			},
		},
		{
			Name:                 fmt.Sprintf("Standard 1024x1792 images (%s)", r.Model),
			Unit:                 "100 images",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      standard10241792Qty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr(fmt.Sprintf("%s Standard HighRes", skuPrefix))},
					{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Standard HighRes Images", skuPrefix))},
				},
			},
		},
		{
			Name:                 fmt.Sprintf("HD 1024x1024 images (%s)", r.Model),
			Unit:                 "100 images",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      hd10241024Qty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr(fmt.Sprintf("%s HD LowRes", skuPrefix))},
					{Key: "meterName", Value: strPtr(fmt.Sprintf("%s HD LowRes Images", skuPrefix))},
				},
			},
		},
		{
			Name:                 fmt.Sprintf("HD 1024x1792 images (%s)", r.Model),
			Unit:                 "100 images",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      hd10241792Qty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr(fmt.Sprintf("%s HD HighRes", skuPrefix))},
					{Key: "meterName", Value: strPtr(fmt.Sprintf("%s HD HighRes Images", skuPrefix))},
				},
			},
		},
	}
}

func (r *CognitiveDeployment) embeddingCostComponents() []*schema.CostComponent {
	sku := embeddingSKUs[r.Model]

	var qty *decimal.Decimal
	if r.MonthlyTextEmbeddingTokens != nil {
		qty = decimalPtr(decimal.NewFromInt(*r.MonthlyTextEmbeddingTokens).Div(decimal.NewFromInt(1_000)))
	}

	return []*schema.CostComponent{
		{
			Name:                 fmt.Sprintf("Text embeddings (%s)", r.Model),
			Unit:                 "1K tokens",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      qty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr(sku)},
					{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Tokens", sku))},
				},
			},
		},
	}
}

func (r *CognitiveDeployment) speechCostComponents() []*schema.CostComponent {
	sku := speechSKUs[r.Model]

	var qty *decimal.Decimal

	if r.Model == "whisper" {
		if r.MonthlyTextToSpeechHours != nil {
			qty = decimalPtr(decimal.NewFromFloat(*r.MonthlyTextToSpeechHours))
		}

		return []*schema.CostComponent{
			{
				Name:                 fmt.Sprintf("Text to speech (%s)", r.Model),
				Unit:                 "hours",
				UnitMultiplier:       decimal.NewFromInt(1),
				MonthlyQuantity:      qty,
				IgnoreIfMissingPrice: true,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr(vendorName),
					Region:        strPtr(r.Region),
					Service:       strPtr("Cognitive Services"),
					ProductFamily: strPtr("AI + Machine Learning"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "productName", Value: strPtr("Azure OpenAI")},
						{Key: "skuName", Value: strPtr(sku)},
						{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Speech to Text Batch", sku))},
					},
				},
			},
		}
	}

	if r.MonthlyTextToSpeechCharacters != nil {
		qty = decimalPtr(decimal.NewFromInt(*r.MonthlyTextToSpeechCharacters).Div(decimal.NewFromInt(1_000_000)))
	}

	return []*schema.CostComponent{
		{
			Name:                 fmt.Sprintf("Text to speech (%s)", r.Model),
			Unit:                 "1M characters",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      qty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr(sku)},
					{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Characters", sku))},
				},
			},
		},
	}
}
