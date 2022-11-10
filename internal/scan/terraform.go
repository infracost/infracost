package scan

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/imdario/mergo"
	jsoniter "github.com/json-iterator/go"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
)

type GetPricesFunc func(ctx *config.RunContext, c *apiclient.PricingAPIClient, r *schema.Resource) error

type TerraformPlanScanner struct {
	client                  *http.Client
	pricingAPIClient        *apiclient.PricingAPIClient
	recommendationAPIClient apiclient.RecommendationClient
	logger                  *log.Entry
	ctx                     *config.RunContext
	getPrices               GetPricesFunc
}

func NewTerraformPlanScanner(ctx *config.RunContext, logger *log.Entry, getPrices GetPricesFunc) *TerraformPlanScanner {
	return &TerraformPlanScanner{
		client:                  &http.Client{Timeout: time.Second * 5},
		pricingAPIClient:        apiclient.NewPricingAPIClient(ctx),
		recommendationAPIClient: apiclient.NewRecommendationClient(ctx.Config, logger),
		logger:                  logger,
		ctx:                     ctx,
		getPrices:               getPrices,
	}
}

func (s *TerraformPlanScanner) ScanPlan(project *schema.Project, projectPlan []byte) error {
	apiSuggestions, err := s.recommendationAPIClient.GetSuggestions(projectPlan)
	if err != nil {
		return fmt.Errorf("failed to get suggestions %w", err)
	}

	var recMap = make(map[string]schema.Suggestions)
	for _, apiSuggestion := range apiSuggestions {
		suggestion := schema.Suggestion{
			ID:                 apiSuggestion.ID,
			Title:              apiSuggestion.Title,
			Description:        apiSuggestion.Description,
			ResourceType:       apiSuggestion.ResourceType,
			ResourceAttributes: apiSuggestion.ResourceAttributes,
			Address:            apiSuggestion.Address,
			Suggested:          apiSuggestion.Suggested,
			NoCost:             apiSuggestion.NoCost,
		}

		if v, ok := recMap[apiSuggestion.ResourceType]; ok {
			recMap[apiSuggestion.ResourceType] = append(v, suggestion)
			continue
		}

		recMap[apiSuggestion.ResourceType] = schema.Suggestions{suggestion}
	}

	var costedSuggestions schema.Suggestions
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
		initialSchema, err := jsoniter.Marshal(coreResource)
		if err != nil {
			s.logger.WithError(err).Error("could not marshal initial schema for suggestion resource")
			continue
		}

		initialResource := coreResource.BuildResource()
		err = s.getPrices(s.ctx, s.pricingAPIClient, initialResource)
		if err != nil {
			s.logger.WithError(err).Error("could not fetch prices for initial resource")
			continue
		}

		initialResource.CalculateCosts()

		for _, suggestion := range recMap[coreResource.CoreType()] {
			if suggestion.Address != initialResource.Name {
				continue
			}

			if suggestion.NoCost {
				costedSuggestions = append(costedSuggestions, suggestion)
				continue
			}

			suggestedAttributes := suggestion.ResourceAttributes
			err = mergeSuggestionWithResource(initialSchema, suggestedAttributes, coreResource)
			if err != nil {
				s.logger.WithError(err).Errorf("could not merge costed resource from scan with baseline resource type: %s", coreResource.CoreType())
				continue
			}

			schemaResource := coreResource.BuildResource()
			err = s.getPrices(s.ctx, s.pricingAPIClient, schemaResource)
			if err != nil {
				s.logger.WithError(err).Errorf("could not fetch prices for costed resource type: %s", coreResource.CoreType())
				continue
			}
			schemaResource.CalculateCosts()

			diff := decimal.Zero
			if schemaResource.MonthlyCost != nil {
				diff = initialResource.MonthlyCost.Sub(*schemaResource.MonthlyCost)
			}

			costedSuggestions = append(costedSuggestions, schema.Suggestion{
				ID:                 suggestion.ID,
				Title:              suggestion.Title,
				Description:        suggestion.Description,
				ResourceType:       suggestion.ResourceType,
				ResourceAttributes: suggestion.ResourceAttributes,
				Address:            suggestion.Address,
				Suggested:          suggestion.Suggested,
				NoCost:             suggestion.NoCost,
				Cost:               &diff,
			})
		}
	}

	sort.Sort(costedSuggestions)
	if project.Metadata == nil {
		project.Metadata = &schema.ProjectMetadata{}
	}

	project.Metadata.Suggestions = costedSuggestions

	return nil
}

func mergeSuggestionWithResource(schema []byte, suggestedSchema []byte, resource schema.CoreResource) error {
	var initialAttributes map[string]interface{}
	jsoniter.Unmarshal(schema, &initialAttributes)

	var suggestedAttributes map[string]interface{}
	jsoniter.Unmarshal(suggestedSchema, &suggestedAttributes)

	err := mergo.Merge(&initialAttributes, suggestedAttributes, mergo.WithOverride, mergo.WithSliceDeepCopy)
	if err != nil {
		return err
	}

	nb, err := jsoniter.Marshal(initialAttributes)
	if err != nil {
		return err
	}
	err = json.Unmarshal(nb, &resource)
	if err != nil {
		return err
	}

	return nil
}
