package scan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/imdario/mergo"
	jsoniter "github.com/json-iterator/go"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/prices"
	"github.com/infracost/infracost/internal/providers/terraform"
	"github.com/infracost/infracost/internal/schema"
)

type Scanner struct {
	client           *http.Client
	pricingAPIClient *apiclient.PricingAPIClient
	logger           *log.Entry
	ctx              *config.RunContext
}

func NewScanner(ctx *config.RunContext, logger *log.Entry) Scanner {
	return Scanner{
		client:           &http.Client{Timeout: time.Second * 5},
		pricingAPIClient: apiclient.NewPricingAPIClient(ctx),
		logger:           logger,
		ctx:              ctx,
	}
}

func (s Scanner) Scan() ([]ProjectSuggestion, error) {
	jsons, err := s.readProjectDirs()
	if err != nil {
		return nil, err
	}

	var suggestions []ProjectSuggestion
	for _, j := range jsons {
		p, err := s.scanProject(j)
		if err != nil {
			s.logger.WithError(err).Errorf("failed to scan project %s", j.HCL.Module.ModulePath)
			continue
		}

		suggestions = append(suggestions, p)
	}

	return suggestions, nil
}

func (s Scanner) scanProject(j projectJSON) (ProjectSuggestion, error) {
	ps := ProjectSuggestion{Path: j.HCL.Module.ModulePath}

	buf := bytes.NewBuffer(j.HCL.JSON)
	req, err := http.NewRequest(http.MethodPost, "http://localhost:8081/recommend", buf)
	if err != nil {
		return ps, fmt.Errorf("failed to build request to suggestions API %w", err)
	}

	res, err := s.client.Do(req)
	if err != nil {
		return ps, fmt.Errorf("failed to make request to suggestions API %w", err)
	}
	var result RecommendDecisionResponse
	json.NewDecoder(res.Body).Decode(&result)
	if len(result.Result) == 0 {
		return ps, nil
	}

	var recMap = make(map[string][]Suggestion)
	for _, suggestion := range result.Result {
		if v, ok := recMap[suggestion.ResourceType]; ok {
			recMap[suggestion.ResourceType] = append(v, suggestion)
			continue
		}

		recMap[suggestion.ResourceType] = []Suggestion{suggestion}
	}

	masterProject, err := j.JSONProvider.LoadResourcesFromSrc(map[string]*schema.UsageData{}, j.HCL.JSON, nil)
	if err != nil {
		return ps, fmt.Errorf("could not load resources for project scan %w", err)
	}

	ps.Name = masterProject.Name

	var costedSuggestions []Suggestion
	for _, resource := range masterProject.PartialResources {
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
		err = prices.GetPrices(s.ctx, s.pricingAPIClient, initialResource)
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
			err = prices.GetPrices(s.ctx, s.pricingAPIClient, schemaResource)
			if err != nil {
				s.logger.WithError(err).Errorf("could not fetch prices for costed resource type: %s", coreResource.CoreType())
				continue
			}
			schemaResource.CalculateCosts()

			diff := decimal.Zero
			if schemaResource.MonthlyCost != nil {
				diff = initialResource.MonthlyCost.Sub(*schemaResource.MonthlyCost)
			}

			costedSuggestions = append(costedSuggestions, Suggestion{
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

	ps.Suggestions = costedSuggestions
	return ps, nil
}

func (s Scanner) readProjectDirs() ([]projectJSON, error) {
	var jsons []projectJSON

	for _, project := range s.ctx.Config.Projects {
		projectCtx := config.NewProjectContext(s.ctx, project, log.Fields{})
		hclProvider, err := terraform.NewHCLProvider(projectCtx, &terraform.HCLProviderConfig{SuppressLogging: true})
		if err != nil {
			return nil, err
		}

		projectJsons, err := hclProvider.LoadPlanJSONs()
		if err != nil {
			return nil, err
		}

		planJsonProvider := terraform.NewPlanJSONProvider(projectCtx, false)
		for _, j := range projectJsons {
			jsons = append(jsons, projectJSON{HCL: j, JSONProvider: planJsonProvider})
		}
	}

	return jsons, nil
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

type projectJSON struct {
	HCL          terraform.HclProject
	JSONProvider *terraform.PlanJSONProvider
}

type RecommendDecisionResponse struct {
	Result []Suggestion `json:"result"`
}

type Suggestion struct {
	ID                 string           `json:"id"`
	Title              string           `json:"title"`
	Description        string           `json:"description"`
	ResourceType       string           `json:"resourceType"`
	ResourceAttributes json.RawMessage  `json:"resourceAttributes"`
	Address            string           `json:"address"`
	Suggested          string           `json:"suggested"`
	NoCost             bool             `json:"no_cost"`
	Cost               *decimal.Decimal `json:"cost"`
}

type ProjectSuggestion struct {
	Path        string       `json:"path"`
	Name        string       `json:"name"`
	Suggestions []Suggestion `json:"suggestions"`
}
