package apiclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	json "github.com/json-iterator/go"

	"github.com/hashicorp/go-cleanhttp"
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
			logging.Logger.Err(err).Msgf("Error reading CA cert file %s", ctx.Config.TLSCACertFile)
		} else {
			ok := rootCAs.AppendCertsFromPEM(caCerts)

			if !ok {
				logging.Logger.Warn().Msg("No CA certs appended, only using system certs")
			} else {
				logging.Logger.Debug().Msgf("Loaded CA certs from %s", ctx.Config.TLSCACertFile)
			}
		}

		tlsConfig.RootCAs = rootCAs
	}

	if ctx.Config.TLSInsecureSkipVerify != nil {
		tlsConfig.InsecureSkipVerify = *ctx.Config.TLSInsecureSkipVerify // nolint: gosec
	}

	httpClient := cleanhttp.DefaultClient()
	httpClient.Transport.(*http.Transport).TLSClientConfig = &tlsConfig

	return &UsageAPIClient{
		APIClient: APIClient{
			httpClient: httpClient,
			endpoint:   ctx.Config.UsageAPIEndpoint,
			apiKey:     ctx.Config.APIKey,
			uuid:       ctx.UUID(),
		},
		Currency: currency,
	}
}

// ListActualCosts queries the Infracost Cloud Usage API to retrieve any cloud provider
// reported costs associated with the resource.
func (c *UsageAPIClient) ListActualCosts(vars ActualCostsQueryVariables) ([]ActualCostsResult, error) {
	query := c.buildActualCostsQuery(vars)

	logging.Logger.Debug().Msgf("Getting actual costs from %s for %s", c.endpoint, vars.Address)

	results, err := c.DoQueries([]GraphQLQuery{query})
	if err != nil {
		return nil, err
	} else if len(results) > 0 && results[0].Get("errors").Exists() {
		return nil, fmt.Errorf("graphql error: %s", results[0].Get("errors").String())
	}

	if len(results) == 0 {
		return nil, nil
	}

	result := results[0]

	actualCostResults := make([]ActualCostsResult, 0)
	for _, ac := range result.Get("data.actualCostsList").Array() {
		acr := ActualCostsResult{
			Address:        ac.Get("address").String(),
			ResourceID:     ac.Get("resourceId").String(),
			StartTimestamp: ac.Get("startAt").Time(),
			EndTimestamp:   ac.Get("endAt").Time(),
		}

		for _, cc := range ac.Get("costComponents").Array() {
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

		actualCostResults = append(actualCostResults, acr)
	}

	return actualCostResults, nil
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
			actualCostsList(repoUrl: $repoUrl, project: $project, address: $address, currency: $currency) {
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
func (c *UsageAPIClient) ListUsageQuantities(vars []*UsageQuantitiesQueryVariables) ([]*schema.UsageData, error) {
	queries := make([]GraphQLQuery, 0, len(vars))
	for _, v := range vars {
		logging.Logger.Debug().Msgf("Getting usage quantities from %s for %s %s %v", c.endpoint, v.ResourceType, v.Address, v.UsageKeys)
		queries = append(queries, c.buildUsageQuantitiesQuery(*v))
	}

	results, err := c.DoQueries(queries)
	if err != nil {
		return nil, err
	} else if len(results) > 0 && results[0].Get("errors").Exists() {
		return nil, fmt.Errorf("graphql error: %s", results[0].Get("errors").String())
	}

	attribsByAddress := make(map[string]map[string]any)
	for _, result := range results {
		for _, q := range result.Get("data.usageQuantities").Array() {
			address := q.Get("address").String()
			if attribsByAddress[address] == nil {
				attribsByAddress[address] = make(map[string]any)
			}

			usageKey := q.Get("usageKey").String()
			unflattenUsageKey(attribsByAddress[address], usageKey, q.Get("monthlyQuantity").String())
		}
	}

	var ud = make([]*schema.UsageData, 0, len(attribsByAddress))
	for address, attribs := range attribsByAddress {
		// now that we have converted the attribs to account for any flattened keys, convert the
		// structure to json so we can return it as the gjson.Result required by for UsageData.Attributes
		attribsJson, err := json.Marshal(attribs)
		if err != nil {
			return nil, err
		}

		ud = append(ud, &schema.UsageData{
			Address:    address,
			Attributes: gjson.ParseBytes(attribsJson).Map(),
		})
	}

	return ud, nil
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
func unflattenUsageKey(attribs map[string]any, usageKey string, value string) {
	split := strings.SplitN(usageKey, ".", 2)
	if len(split) <= 1 {
		attribs[usageKey] = value
		return
	}

	var childAttribs map[string]any
	if val, ok := attribs[split[0]]; ok {
		childAttribs = val.(map[string]any)
	} else {
		// sub attrib map doesn't already exist so add it to the parent
		childAttribs = make(map[string]any)
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
		logging.Logger.Debug().Msgf("No cloud resource IDs to upload for %s %s", vars.RepoURL, vars.ProjectWithWorkspace)
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

	logging.Logger.Debug().Msgf("Uploading cloud resource IDs to %s for %s %s", c.endpoint, vars.RepoURL, vars.ProjectWithWorkspace)

	results, err := c.DoQueries([]GraphQLQuery{query})
	if err != nil {
		return err
	} else if len(results) > 0 && results[0].Get("errors").Exists() {
		return fmt.Errorf("graphql error: %s", results[0].Get("errors").String())
	}

	newCount := results[0].Get("data.addAddressResourceIds.newCount").Int()

	logging.Logger.Debug().Str("newCount", fmt.Sprintf("%d", newCount)).Msgf("Uploaded cloud resource IDs to %s for %s %s", c.endpoint, vars.RepoURL, vars.ProjectWithWorkspace)

	return nil
}

func interfaceToMap(in any) map[string]any {
	out := map[string]any{}
	b, _ := json.Marshal(in)
	_ = json.Unmarshal(b, &out)
	return out
}
