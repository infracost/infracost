package apiclient

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"github.com/tidwall/gjson"
	"io/ioutil"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
)

type UsageAPIClient struct {
	APIClient
	Currency string
}

// ActualCostResult contains the cost component information of actual costs retrieved from
// the Infracost Cloud Usage API
type ActualCostResult struct {
	Address         string
	UsageType       string
	Description     string
	MonthlyCost     string
	MonthlyQuantity string
	Price           string
	Unit            string
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

		caCerts, err := ioutil.ReadFile(ctx.Config.TLSCACertFile)
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
func (c *UsageAPIClient) ListActualCosts(vars ActualCostsQueryVariables) ([]ActualCostResult, error) {
	query := c.buildActualCostsQuery(vars)

	logging.Logger.Debugf("Getting actual costs from %s for %s", c.endpoint, vars.Address)

	var actualCosts []ActualCostResult

	results, err := c.doQueries([]GraphQLQuery{query})
	if err != nil {
		return actualCosts, err
	}

	for _, result := range results {
		for _, ac := range result.Get("data.actualCosts").Array() {
			actualCosts = append(actualCosts, ActualCostResult{
				Address:         ac.Get("address").String(),
				UsageType:       ac.Get("usageType").String(),
				Description:     ac.Get("description").String(),
				Unit:            ac.Get("unit").String(),
				Price:           ac.Get("price").String(),
				MonthlyCost:     ac.Get("monthlyCost").String(),
				MonthlyQuantity: ac.Get("monthlyQuantity").String(),
			})
		}
	}
	return actualCosts, nil
}

type ActualCostsQueryVariables struct {
	RepoURL  string `json:"repoUrl"`
	Project  string `json:"project"`
	Address  string `json:"address"`
	Currency string `json:"currency"`
}

func (c *UsageAPIClient) buildActualCostsQuery(vars ActualCostsQueryVariables) GraphQLQuery {
	v := interfaceToMap(vars)

	query := `
		query($repoUrl: String!, $project: String!, $address: String!, $currency: String!) {
			actualCosts(repoUrl: $repoUrl, project: $project, address: $address, currency: $currency) {
    			address
				usageType
				description
    			currency
    			monthlyCost
				monthlyQuantity
				price
                unit
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

	attribs := make(map[string]gjson.Result)

	results, err := c.doQueries([]GraphQLQuery{query})
	if err != nil {
		return nil, err
	}

	for _, result := range results {
		for _, q := range result.Get("data.usageQuantities").Array() {
			usageKey := q.Get("usageKey").String()
			attribs[usageKey] = q.Get("monthlyQuantity")
		}
	}

	return attribs, nil
}

type UsageQuantitiesQueryVariables struct {
	RepoURL      string   `json:"repoUrl"`
	Project      string   `json:"project"`
	ResourceType string   `json:"resourceType"`
	Address      string   `json:"address"`
	UsageKeys    []string `json:"usageKeys"`
}

func (c *UsageAPIClient) buildUsageQuantitiesQuery(vars UsageQuantitiesQueryVariables) GraphQLQuery {
	v := interfaceToMap(vars)

	query := `
		query($repoUrl: String!, $project: String!, $resourceType: String!, $address: String!, $usageKeys: [String!]!) {
			usageQuantities(repoUrl: $repoUrl, project: $project, resourceType: $resourceType, address: $address, usageKeys: $usageKeys) {
    			address
				usageKey
				monthlyQuantity
			}
		}
	`

	return GraphQLQuery{query, v}
}

func interfaceToMap(in interface{}) map[string]interface{} {
	out := map[string]interface{}{}
	b, _ := json.Marshal(in)
	_ = json.Unmarshal(b, &out)
	return out
}
