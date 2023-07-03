package prices

import (
	"fmt"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"

	"github.com/tidwall/gjson"
)

var (
	excludedEnv = map[string]struct{}{
		"repoMetadata": {},
	}
)

type PricingClient interface {
}

type PricingAPIClient struct {
	apiclient.APIClient

	Currency       string
	EventsDisabled bool
}

type PriceQueryKey struct {
	Resource      *schema.Resource
	CostComponent *schema.CostComponent
}

type PriceQueryResult struct {
	PriceQueryKey
	Result gjson.Result
}

func NewPricingAPIClient(ctx *config.RunContext) *PricingAPIClient {
	currency := ctx.Config.Currency
	if currency == "" {
		currency = "USD"
	}

	return &PricingAPIClient{
		APIClient:      apiclient.NewTLSEnabledClient(ctx),
		Currency:       currency,
		EventsDisabled: ctx.Config.EventsDisabled,
	}
}

func (c *PricingAPIClient) AddEvent(name string, env map[string]interface{}) error {
	if c.EventsDisabled {
		return nil
	}

	filtered := make(map[string]interface{})
	for k, v := range env {
		if _, ok := excludedEnv[k]; ok {
			continue
		}

		filtered[k] = v
	}

	d := map[string]interface{}{
		"event": name,
		"env":   filtered,
	}

	_, err := c.DoRequest("POST", "/event", d)
	return err
}

func (c *PricingAPIClient) RunQueries(r *schema.Resource) ([]PriceQueryResult, error) {
	keys, queries := c.batchQueries(r)

	if len(queries) == 0 {
		logging.Logger.Debugf("Skipping getting pricing details for %s since there are no queries to run", r.Name)
		return []PriceQueryResult{}, nil
	}

	logging.Logger.Debugf("Getting pricing details for %s", r.Name)

	results, err := c.DoQueries(queries)
	if err != nil {
		return []PriceQueryResult{}, err
	}

	return c.zipQueryResults(keys, results), nil
}

func (c *PricingAPIClient) buildQuery(product *schema.ProductFilter, price *schema.PriceFilter) apiclient.GraphQLQuery {
	v := map[string]interface{}{}
	v["productFilter"] = product
	v["priceFilter"] = price

	query := fmt.Sprintf(`
		query($productFilter: ProductFilter!, $priceFilter: PriceFilter) {
			products(filter: $productFilter) {
				prices(filter: $priceFilter) {
					priceHash
					%s
				}
			}
		}
	`, c.Currency)

	return apiclient.GraphQLQuery{query, v}
}

// Batch all the queries for this resource so we can use one GraphQL call.
// Use PriceQueryKeys to keep track of which query maps to which sub-resource and price component.
func (c *PricingAPIClient) batchQueries(r *schema.Resource) ([]PriceQueryKey, []apiclient.GraphQLQuery) {
	keys := make([]PriceQueryKey, 0)
	queries := make([]apiclient.GraphQLQuery, 0)

	for _, component := range r.CostComponents {
		keys = append(keys, PriceQueryKey{r, component})
		queries = append(queries, c.buildQuery(component.ProductFilter, component.PriceFilter))
	}

	for _, subresource := range r.FlattenedSubResources() {
		for _, component := range subresource.CostComponents {
			keys = append(keys, PriceQueryKey{subresource, component})
			queries = append(queries, c.buildQuery(component.ProductFilter, component.PriceFilter))
		}
	}

	return keys, queries
}

func (c *PricingAPIClient) zipQueryResults(k []PriceQueryKey, r []gjson.Result) []PriceQueryResult {
	res := make([]PriceQueryResult, 0, len(k))

	for i, k := range k {
		res = append(res, PriceQueryResult{
			PriceQueryKey: k,
			Result:        r[i],
		})
	}

	return res
}
