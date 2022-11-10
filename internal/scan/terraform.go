package scan

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/imdario/mergo"
	jsoniter "github.com/json-iterator/go"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
)

// GetPricesFunc fetches a price for the given resource r using client c.
// This interface is extracted to avoid circular deps and ease of testing.
type GetPricesFunc func(ctx *config.RunContext, c *apiclient.PricingAPIClient, r *schema.Resource) error

// TerraformPlanScanner scans a plan for Infracost Cloud cost optimizations. These optimizations are provided by the
// recommendations API and the scanner links any suggestions to raw resources. It attempts to find cost estimates for any
// recommendations that are found.
type TerraformPlanScanner struct {
	pricingAPIClient        *apiclient.PricingAPIClient
	recommendationAPIClient apiclient.RecommendationClient
	logger                  *log.Entry
	ctx                     *config.RunContext
	getPrices               GetPricesFunc
}

// NewTerraformPlanScanner returns an initialised TerraformPlanScanner.
func NewTerraformPlanScanner(ctx *config.RunContext, logger *log.Entry, getPrices GetPricesFunc) *TerraformPlanScanner {
	return &TerraformPlanScanner{
		pricingAPIClient:        apiclient.NewPricingAPIClient(ctx),
		recommendationAPIClient: apiclient.NewRecommendationClient(ctx.Config, logger),
		logger:                  logger,
		ctx:                     ctx,
		getPrices:               getPrices,
	}
}

// ScanPlan scans the provided projectPlan for the project, if any Recommendations are found for the plan
// the Scanner will attempt to fetch costs for the suggestion and given resource. These suggestions will only
// be provided for resources that are marked as a schema.CoreResource.
func (s *TerraformPlanScanner) ScanPlan(project *schema.Project, projectPlan []byte) error {
	apiRecommendations, err := s.recommendationAPIClient.GetRecommendations(projectPlan)
	if err != nil {
		return fmt.Errorf("failed to get suggestions %w", err)
	}

	var recMap = make(map[string]schema.Recommendations)
	for _, apiRecommendation := range apiRecommendations {
		recommendation := schema.Recommendation{
			ID:                 apiRecommendation.ID,
			Title:              apiRecommendation.Title,
			Description:        apiRecommendation.Description,
			ResourceType:       apiRecommendation.ResourceType,
			ResourceAttributes: apiRecommendation.ResourceAttributes,
			Address:            apiRecommendation.Address,
			Suggested:          apiRecommendation.Suggested,
			NoCost:             apiRecommendation.NoCost,
		}

		if v, ok := recMap[apiRecommendation.ResourceType]; ok {
			recMap[apiRecommendation.ResourceType] = append(v, recommendation)
			continue
		}

		recMap[apiRecommendation.ResourceType] = schema.Recommendations{recommendation}
	}

	var costedRecommendations schema.Recommendations
	for _, resource := range project.PartialResources {
		coreResource := resource.CoreResource
		if coreResource == nil {
			continue
		}

		if _, ok := recMap[coreResource.CoreType()]; !ok {
			continue
		}

		// TODO: fetch usage and populate resource
		coreResource.PopulateUsage(nil)
		baselineSchema, err := jsoniter.Marshal(coreResource)
		if err != nil {
			s.logger.WithError(err).Debug("could not marshal initial schema for recommendation resource")
			continue
		}

		baselineResource, err := s.buildResource(coreResource)
		if err != nil {
			s.logger.WithError(err).Debug("could not fetch prices for initial resource")
			continue
		}

		for _, recommendation := range recMap[coreResource.CoreType()] {
			if recommendation.Address != baselineResource.Name {
				continue
			}

			costedRecommendation, err := s.costSuggestion(coreResource, baselineSchema, baselineResource, recommendation)
			if err != nil {
				s.logger.WithError(err).Debugf("failed to cost recommendation for resource %s", baselineResource.Name)
				continue
			}

			costedRecommendations = append(costedRecommendations, costedRecommendation)
		}
	}

	sort.Sort(costedRecommendations)
	if project.Metadata == nil {
		project.Metadata = &schema.ProjectMetadata{}
	}

	project.Metadata.Recommendations = costedRecommendations

	return nil
}

func (s *TerraformPlanScanner) costSuggestion(coreResource schema.CoreResource, baselineSchema []byte, baselineResource *schema.Resource, recommendation schema.Recommendation) (schema.Recommendation, error) {
	if recommendation.NoCost {
		return recommendation, nil
	}

	suggestedAttributes := recommendation.ResourceAttributes
	err := mergeSuggestionWithResource(baselineSchema, suggestedAttributes, coreResource)
	if err != nil {
		return recommendation, fmt.Errorf("could not merge costed resource from scan with baseline resource type: %s", coreResource.CoreType())
	}

	schemaResource, err := s.buildResource(coreResource)
	if err != nil {
		return recommendation, fmt.Errorf("could not fetch prices for costed resource type: %s", coreResource.CoreType())
	}

	diff := decimal.Zero
	if schemaResource.MonthlyCost != nil {
		diff = baselineResource.MonthlyCost.Sub(*schemaResource.MonthlyCost)
	}

	return schema.Recommendation{
		ID:                 recommendation.ID,
		Title:              recommendation.Title,
		Description:        recommendation.Description,
		ResourceType:       recommendation.ResourceType,
		ResourceAttributes: recommendation.ResourceAttributes,
		Address:            recommendation.Address,
		Suggested:          recommendation.Suggested,
		NoCost:             recommendation.NoCost,
		Cost:               &diff,
	}, nil
}

func (s *TerraformPlanScanner) buildResource(coreResource schema.CoreResource) (*schema.Resource, error) {
	r := coreResource.BuildResource()
	err := s.getPrices(s.ctx, s.pricingAPIClient, r)
	if err != nil {
		return nil, fmt.Errorf("could not fetch prices for core resource %s %w", coreResource.CoreType(), err)
	}

	r.CalculateCosts()

	return r, nil
}

func mergeSuggestionWithResource(schema []byte, suggestedSchema []byte, resource schema.CoreResource) error {
	var initialAttributes map[string]interface{}
	err := jsoniter.Unmarshal(schema, &initialAttributes)
	if err != nil {
		return fmt.Errorf("failed to unmarshal initial resource attributes %w", err)
	}

	var suggestedAttributes map[string]interface{}
	err = jsoniter.Unmarshal(suggestedSchema, &suggestedAttributes)
	if err != nil {
		return fmt.Errorf("failed to unmarshal suggested resource attributes %w", err)
	}

	err = mergo.Merge(&initialAttributes, suggestedAttributes, mergo.WithOverride, mergo.WithSliceDeepCopy)
	if err != nil {
		return fmt.Errorf("failed to merge initial attribute with suggested %w", err)
	}

	nb, err := jsoniter.Marshal(initialAttributes)
	if err != nil {
		return fmt.Errorf("failed to marshal attributes with suggestions")
	}

	err = json.Unmarshal(nb, &resource)
	if err != nil {
		return fmt.Errorf("failed to unmarshall attributes with suggestions")
	}

	return nil
}
