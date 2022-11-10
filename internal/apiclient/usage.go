package apiclient

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"
)

type UsageAPIClient struct {
	APIClient
	Currency string
}

// ActualCostsResult contains the cost information of actual costs retrieved from
// the Infracost Cloud Usage API
type ActualCostsResult struct {
	Address        string
	ResourceID     string
	StartTimestamp time.Time
	EndTimestamp   time.Time
	CostComponents []ActualCostComponent
}

// ActualCostComponent represents an individual line item of actual costs for a resource
type ActualCostComponent struct {
	UsageType       string
	Description     string
	MonthlyCost     string
	MonthlyQuantity string
	Price           string
	Unit            string
	Currency        string
}

// NewUsageAPIClient returns a new Infracost Cloud Usage API Client configured from the RunContext
func NewUsageAPIClient(ctx *config.RunContext) *UsageAPIClient {
	currency := ctx.Config.Currency
	if currency == "" {
		currency = "USD"
	}

	tlsConfig := tls.Config{} // nolint: gosec

	if ctx.Config.TLSCACertFile != "" {
		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}

		caCerts, err := os.ReadFile(ctx.Config.TLSCACertFile)
		if err != nil {
			logging.Logger.WithError(err).Errorf("Error reading CA cert file %s", ctx.Config.TLSCACertFile)
		} else {
			ok := rootCAs.AppendCertsFromPEM(caCerts)

			if !ok {
				logging.Logger.Warningf("No CA certs appended, only using system certs")
			} else {
				logging.Logger.Debugf("Loaded CA certs from %s", ctx.Config.TLSCACertFile)
			}
		}

		tlsConfig.RootCAs = rootCAs
	}

	if ctx.Config.TLSInsecureSkipVerify != nil {
		tlsConfig.InsecureSkipVerify = *ctx.Config.TLSInsecureSkipVerify
	}

	return &UsageAPIClient{
		APIClient: APIClient{
			endpoint:  ctx.Config.UsageAPIEndpoint,
			apiKey:    ctx.Config.APIKey,
			tlsConfig: &tlsConfig,
			uuid:      ctx.UUID(),
		},
		Currency: currency,
	}
}

// ListActualCosts queries the Infracost Cloud Usage API to retrieve any cloud provider
// reported costs associated with the resource.
func (c *UsageAPIClient) ListActualCosts(vars ActualCostsQueryVariables) (*ActualCostsResult, error) {
	query := c.buildActualCostsQuery(vars)

	logging.Logger.Debugf("Getting actual costs from %s for %s", c.endpoint, vars.Address)

	results, err := c.doQueries([]GraphQLQuery{query})
	if err != nil {
		return nil, err
	} else if len(results) > 0 && results[0].Get("errors").Exists() {
		return nil, fmt.Errorf("graphql error: %s", results[0].Get("errors").String())
	}

	if len(results) == 0 {
		return nil, nil
	}

	result := results[0]
	acr := &ActualCostsResult{
		Address:        result.Get("data.actualCosts.address").String(),
		ResourceID:     result.Get("data.actualCosts.resourceId").String(),
		StartTimestamp: result.Get("data.actualCosts.startAt").Time(),
		EndTimestamp:   result.Get("data.actualCosts.endAt").Time(),
	}

	for _, cc := range result.Get("data.actualCosts.costComponents").Array() {
		acr.CostComponents = append(acr.CostComponents, ActualCostComponent{
			UsageType:       cc.Get("usageType").String(),
			Description:     cc.Get("description").String(),
			Unit:            cc.Get("unit").String(),
			Price:           cc.Get("price").String(),
			MonthlyCost:     cc.Get("monthlyCost").String(),
			MonthlyQuantity: cc.Get("monthlyQuantity").String(),
			Currency:        cc.Get("currency").String(),
		})
	}

	return acr, nil
}

type ActualCostsQueryVariables struct {
	RepoURL              string `json:"repoUrl"`
	ProjectWithWorkspace string `json:"project"`
	Address              string `json:"address"`
	Currency             string `json:"currency"`
}

func (c *UsageAPIClient) buildActualCostsQuery(vars ActualCostsQueryVariables) GraphQLQuery {
	v := interfaceToMap(vars)

	query := `
		query($repoUrl: String!, $project: String!, $address: String!, $currency: String!) {
			actualCosts(repoUrl: $repoUrl, project: $project, address: $address, currency: $currency) {
				address
				resourceId
				startAt
				endAt
				costComponents {
					usageType
					description
					currency
					monthlyCost
					monthlyQuantity
					price
					unit
				}
			}
		}
	`

	return GraphQLQuery{query, v}
}

// ListUsageQuantities queries the Infracost Cloud Usage API to retrieve usage estimates
// derived from cloud provider reported usage and costs.
func (c *UsageAPIClient) ListUsageQuantities(vars UsageQuantitiesQueryVariables) (map[string]gjson.Result, error) {
	query := c.buildUsageQuantitiesQuery(vars)

	logging.Logger.Debugf("Getting usage quantities from %s for %s %s %v", c.endpoint, vars.ResourceType, vars.Address, vars.UsageKeys)

	results, err := c.doQueries([]GraphQLQuery{query})
	if err != nil {
		return nil, err
	} else if len(results) > 0 && results[0].Get("errors").Exists() {
		return nil, fmt.Errorf("graphql error: %s", results[0].Get("errors").String())
	}

	attribs := make(map[string]interface{})
	for _, result := range results {
		for _, q := range result.Get("data.usageQuantities").Array() {
			usageKey := q.Get("usageKey").String()
			unflattenUsageKey(attribs, usageKey, q.Get("monthlyQuantity").String())
		}
	}

	// now that we have converted the attribs to account for any flattened keys, convert the
	// structure to json so we can return it as the gjson.Result required by for UsageData.Attributes
	attribsJson, err := json.Marshal(attribs)
	if err != nil {
		return nil, err
	}

	return gjson.ParseBytes(attribsJson).Map(), nil
}

// unflattenUsageKey converts a "." separated usage key returned from the Usage API to the
// nested structure used by the usage-file.
//
// Nested usage keys are returned from the usage API in a graphQL-friendly flattened format
// with "." as a key separator. For example s3 standard usage is retrieved as:
// [
//
//	{ "usageKey": "standard.storage_gb", "monthlyQuantity": "123" },
//	{ "usageKey": "standard.monthly_tier_1_requests", "monthlyQuantity": "456" },
//	...
//
// ]
//
// When converted to a nested format needed for for UsageData.Attributes, the keys would be:
//
//	{
//	   "standard": {
//	     "storage_gb: "123",
//	     "monthly_tier_1_requests": "456",
//	     ...
//	   },
//	   ...
//	}
func unflattenUsageKey(attribs map[string]interface{}, usageKey string, value string) {
	split := strings.SplitN(usageKey, ".", 2)
	if len(split) <= 1 {
		attribs[usageKey] = value
		return
	}

	var childAttribs map[string]interface{}
	if val, ok := attribs[split[0]]; ok {
		childAttribs = val.(map[string]interface{})
	} else {
		// sub attrib map doesn't already exist so add it to the parent
		childAttribs = make(map[string]interface{})
		attribs[split[0]] = childAttribs
	}

	// populate the value in the childMap (recursively, in case there are multiple ".")
	unflattenUsageKey(childAttribs, split[1], value)
}

type UsageQuantitiesQueryVariables struct {
	RepoURL              string              `json:"repoUrl"`
	ProjectWithWorkspace string              `json:"project"`
	ResourceType         string              `json:"resourceType"`
	Address              string              `json:"address"`
	UsageKeys            []string            `json:"usageKeys"`
	UsageParams          []schema.UsageParam `json:"usageParams"`
}

func (c *UsageAPIClient) buildUsageQuantitiesQuery(vars UsageQuantitiesQueryVariables) GraphQLQuery {
	v := interfaceToMap(vars)

	query := `
		query($repoUrl: String!, $project: String!, $resourceType: String!, $address: String!, $usageKeys: [String!]!, $usageParams: [UsageParamInput!]) {
			usageQuantities(repoUrl: $repoUrl, project: $project, resourceType: $resourceType, address: $address, usageKeys: $usageKeys, usageParams: $usageParams) {
    			address
				usageKey
				monthlyQuantity
			}
		}
	`

	return GraphQLQuery{query, v}
}

type CloudResourceIDVariables struct {
	RepoURL              string              `json:"repoUrl"`
	ProjectWithWorkspace string              `json:"project"`
	ResourceIDAddresses  []ResourceIDAddress `json:"addressResourceIds"`
}

type ResourceIDAddress struct {
	Address    string `json:"address"`
	ResourceID string `json:"resourceId"`
}

// UploadCloudResourceIDs uploads cloud resource IDs to the Infracost Cloud Usage API, so they may be
// used to calculate usage estimates.
func (c *UsageAPIClient) UploadCloudResourceIDs(vars CloudResourceIDVariables) error {
	if len(vars.ResourceIDAddresses) == 0 {
		logging.Logger.Debugf("No cloud resource IDs to upload for %s %s", vars.RepoURL, vars.ProjectWithWorkspace)
		return nil
	}

	query := GraphQLQuery{
		Query: `
			mutation($repoUrl: String!, $project: String!, $addressResourceIds: [AddressResourceIdInput!]!) {
				addAddressResourceIds(repoUrl: $repoUrl, project: $project, addressResourceIds: $addressResourceIds) {
					newCount
				} 
			}
		`,
		Variables: interfaceToMap(vars),
	}

	logging.Logger.Debugf("Uploading cloud resource IDs to %s for %s %s", c.endpoint, vars.RepoURL, vars.ProjectWithWorkspace)

	results, err := c.doQueries([]GraphQLQuery{query})
	if err != nil {
		return err
	} else if len(results) > 0 && results[0].Get("errors").Exists() {
		return fmt.Errorf("graphql error: %s", results[0].Get("errors").String())
	}

	newCount := results[0].Get("data.addAddressResourceIds.newCount").Int()

	logging.Logger.WithField("newCount", newCount).Debugf("Uploaded cloud resource IDs to %s for %s %s", c.endpoint, vars.RepoURL, vars.ProjectWithWorkspace)

	return nil
}

func interfaceToMap(in interface{}) map[string]interface{} {
	out := map[string]interface{}{}
	b, _ := json.Marshal(in)
	_ = json.Unmarshal(b, &out)
	return out
}
