package apiclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"

	"github.com/tidwall/gjson"
)

type PricingAPIClient struct {
	APIClient
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
	currency := ctx.Config().Currency
	if currency == "" {
		currency = "USD"
	}

	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	if ctx.Config().TLSCACertFile != "" {
		caCerts, err := ioutil.ReadFile(ctx.Config().TLSCACertFile)
		if err != nil {
			ctx.Logger().Error().Err(err).Msgf("Error reading CA cert file %s", ctx.Config().TLSCACertFile)
		} else {
			ok := rootCAs.AppendCertsFromPEM(caCerts)

			if !ok {
				ctx.Logger().Warn().Msg("No CA certs appended, only using system certs")
			} else {
				ctx.Logger().Debug().Msgf("Loaded CA certs from %s", ctx.Config().TLSCACertFile)
			}
		}
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: ctx.Config().TLSInsecureSkipVerify, // nolint: gosec
		RootCAs:            rootCAs,
	}

	return &PricingAPIClient{
		APIClient: APIClient{
			endpoint:  ctx.Config().PricingAPIEndpoint,
			apiKey:    ctx.Config().APIKey,
			tlsConfig: tlsConfig,
		},
		Currency:       currency,
		EventsDisabled: ctx.Config().EventsDisabled,
	}
}

func (c *PricingAPIClient) AddEvent(name string, env map[string]interface{}) error {
	if c.EventsDisabled {
		return nil
	}

	d := map[string]interface{}{
		"event": name,
		"env":   env,
	}

	_, err := c.doRequest("POST", "/event", d)
	return err
}

func (c *PricingAPIClient) RunQueries(ctx *config.RunContext, r *schema.Resource) ([]PriceQueryResult, error) {
	keys, queries := c.batchQueries(r)

	if len(queries) == 0 {
		ctx.Logger().Debug().Str("resource", r.Name).Msg("Skipping getting pricing details since there are no queries to run")
		return []PriceQueryResult{}, nil
	}

	ctx.Logger().Debug().Str("resource", r.Name).Msg("Getting pricing details")

	results, err := c.doQueries(ctx, queries)
	if err != nil {
		return []PriceQueryResult{}, err
	}

	return c.zipQueryResults(keys, results), nil
}

func (c *PricingAPIClient) buildQuery(product *schema.ProductFilter, price *schema.PriceFilter) GraphQLQuery {
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
