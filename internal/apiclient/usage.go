package apiclient

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/tidwall/gjson"
	"io/ioutil"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"

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
func (c *UsageAPIClient) ListActualCosts(repoURL, projectID string, r *schema.Resource) ([]ActualCostResult, error) {
	query := c.buildActualCostsQuery(repoURL, projectID, r)

	logging.Logger.Debugf("Getting actual costs from %s for %s", c.endpoint, r.Name)

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

func (c *UsageAPIClient) buildActualCostsQuery(repoURL, projectID string, r *schema.Resource) GraphQLQuery {
	v := map[string]interface{}{}
	v["repoUrl"] = repoURL
	v["project"] = projectID
	v["address"] = r.Name
	v["currency"] = c.Currency

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
func (c *UsageAPIClient) ListUsageQuantities(repoURL, projectID, resourceType, address string, usageKeys []string) (map[string]gjson.Result, error) {
	query := c.buildUsageQuantitiesQuery(repoURL, projectID, resourceType, address, usageKeys)

	logging.Logger.Debugf("Getting usage quantities from %s for %s %s %v", c.endpoint, resourceType, address, usageKeys)

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

func (c *UsageAPIClient) buildUsageQuantitiesQuery(repoURL, projectID, resourceType, address string, usageKeys []string) GraphQLQuery {
	v := map[string]interface{}{}
	v["repoUrl"] = repoURL
	v["project"] = projectID
	v["resourceType"] = resourceType
	v["address"] = address
	v["usageKeys"] = usageKeys

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
