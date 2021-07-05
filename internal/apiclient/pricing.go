package apiclient

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

type PricingAPIClient struct {
	APIClient
}

type PriceQueryKey struct {
	Resource      *schema.Resource
	CostComponent *schema.CostComponent
}

type PriceQueryResult struct {
	PriceQueryKey
	Result gjson.Result
}

func NewPricingAPIClient(cfg *config.Config) *PricingAPIClient {
	return &PricingAPIClient{
		APIClient{
			endpoint: cfg.PricingAPIEndpoint,
			apiKey:   cfg.APIKey,
		},
	}
}

func (c *PricingAPIClient) RunQueries(r *schema.Resource) ([]PriceQueryResult, error) {
	keys, queries := c.batchQueries(r)

	if len(queries) == 0 {
		log.Debugf("Skipping getting pricing details for %s since there are no queries to run", r.Name)
		return []PriceQueryResult{}, nil
	}

	log.Debugf("Getting pricing details from %s for %s", c.endpoint, r.Name)

	results, err := c.doQueries(queries)
	if err != nil {
		return []PriceQueryResult{}, err
	}

	return c.zipQueryResults(keys, results), nil
}

func (c *PricingAPIClient) buildQuery(product *schema.ProductFilter, price *schema.PriceFilter) GraphQLQuery {
	v := map[string]interface{}{}
	v["productFilter"] = product
	v["priceFilter"] = price

	query := `
		query($productFilter: ProductFilter!, $priceFilter: PriceFilter) {
			products(filter: $productFilter) {
				prices(filter: $priceFilter) {
					priceHash
					USD
				}
			}
		}
	`

	return GraphQLQuery{query, v}
}

// Batch all the queries for this resource so we can use one GraphQL call.
// Use PriceQueryKeys to keep track of which query maps to which sub-resource and price component.
func (c *PricingAPIClient) batchQueries(r *schema.Resource) ([]PriceQueryKey, []GraphQLQuery) {
	keys := make([]PriceQueryKey, 0)
	queries := make([]GraphQLQuery, 0)

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
