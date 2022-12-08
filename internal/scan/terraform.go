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
// policy API and the scanner links any suggestions to raw resources. It attempts to find cost estimates for any
// policies that are found.
type TerraformPlanScanner struct {
	pricingAPIClient *apiclient.PricingAPIClient
	policyAPIClient  apiclient.PolicyClient
	logger           *log.Entry
	ctx              *config.RunContext
	getPrices        GetPricesFunc
}

// NewTerraformPlanScanner returns an initialised TerraformPlanScanner.
func NewTerraformPlanScanner(ctx *config.RunContext, logger *log.Entry, getPrices GetPricesFunc) *TerraformPlanScanner {
	return &TerraformPlanScanner{
		pricingAPIClient: apiclient.NewPricingAPIClient(ctx),
		policyAPIClient:  apiclient.NewPolicyClient(ctx.Config, logger),
		logger:           logger,
		ctx:              ctx,
		getPrices:        getPrices,
	}
}

// ScanPlan scans the provided projectPlan for the project, if any Policies are found for the plan
// the Scanner will attempt to fetch costs for the suggestion and given resource. These suggestions will only
// be provided for resources that are marked as a schema.CoreResource.
func (s *TerraformPlanScanner) ScanPlan(project *schema.Project, projectPlan []byte) error {
	apiPolicies, err := s.policyAPIClient.GetPolicies(projectPlan)
	if err != nil {
		return fmt.Errorf("failed to get suggestions %w", err)
	}

	var recMap = make(map[string]schema.Policies)
	for _, apiPolicy := range apiPolicies {
		policy := schema.Policy{
			ID:                 apiPolicy.ID,
			Title:              apiPolicy.Title,
			Description:        apiPolicy.Description,
			ResourceType:       apiPolicy.ResourceType,
			ResourceAttributes: apiPolicy.ResourceAttributes,
			Address:            apiPolicy.Address,
			Suggested:          apiPolicy.Suggested,
			NoCost:             apiPolicy.NoCost,
		}

		if v, ok := recMap[apiPolicy.ResourceType]; ok {
			recMap[apiPolicy.ResourceType] = append(v, policy)
			continue
		}

		recMap[apiPolicy.ResourceType] = schema.Policies{policy}
	}

	var costedPolicies schema.Policies
	for _, resource := range project.PartialResources {
		coreResource := resource.CoreResource
		if coreResource == nil {
			continue
		}

		if _, ok := recMap[coreResource.CoreType()]; !ok {
			continue
		}

		baselineSchema, err := jsoniter.Marshal(coreResource)
		if err != nil {
			s.logger.WithError(err).Debug("could not marshal initial schema for policy resource")
			continue
		}

		baselineResource, err := s.buildResource(coreResource, resource.ResourceData.UsageData)
		if err != nil {
			s.logger.WithError(err).Debug("could not fetch prices for initial resource")
			continue
		}

		for _, policy := range recMap[coreResource.CoreType()] {
			if policy.Address != baselineResource.Name {
				continue
			}

			costedPolicy, err := s.costSuggestion(coreResource, baselineSchema, baselineResource, resource.ResourceData.UsageData, policy)
			if err != nil {
				s.logger.WithError(err).Debugf("failed to cost policy for resource %s", baselineResource.Name)
				continue
			}

			costedPolicies = append(costedPolicies, costedPolicy)
		}
	}

	sort.Sort(costedPolicies)
	if project.Metadata == nil {
		project.Metadata = &schema.ProjectMetadata{}
	}

	project.Metadata.Policies = costedPolicies

	return nil
}

func (s *TerraformPlanScanner) costSuggestion(coreResource schema.CoreResource, baselineSchema []byte, baselineResource *schema.Resource, usage *schema.UsageData, policy schema.Policy) (schema.Policy, error) {
	if policy.NoCost {
		return policy, nil
	}

	suggestedAttributes := policy.ResourceAttributes
	err := mergeSuggestionWithResource(baselineSchema, suggestedAttributes, coreResource)
	if err != nil {
		return policy, fmt.Errorf("could not merge costed resource from scan with baseline resource type: %s", coreResource.CoreType())
	}

	schemaResource, err := s.buildResource(coreResource, usage)
	if err != nil {
		return policy, fmt.Errorf("could not fetch prices for costed resource type: %s", coreResource.CoreType())
	}

	diff := decimal.Zero
	if schemaResource.MonthlyCost != nil {
		diff = baselineResource.MonthlyCost.Sub(*schemaResource.MonthlyCost)
	}

	return schema.Policy{
		ID:                 policy.ID,
		Title:              policy.Title,
		Description:        policy.Description,
		ResourceType:       policy.ResourceType,
		ResourceAttributes: policy.ResourceAttributes,
		Address:            policy.Address,
		Suggested:          policy.Suggested,
		NoCost:             policy.NoCost,
		Cost:               &diff,
	}, nil
}

func (s *TerraformPlanScanner) buildResource(coreResource schema.CoreResource, usage *schema.UsageData) (*schema.Resource, error) {
	coreResource.PopulateUsage(usage)
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
		return fmt.Errorf("failed to unmarshal attributes with suggestions")
	}

	return nil
}
