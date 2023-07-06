package apiclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"math"
	"net/http"
	"os"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

var (
	excludedEnv = map[string]struct{}{
		"repoMetadata": {},
	}
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

type BatchRequest struct {
	keys    []PriceQueryKey
	queries []GraphQLQuery
}

func NewPricingAPIClient(ctx *config.RunContext) *PricingAPIClient {
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
			log.Errorf("Error reading CA cert file %s: %v", ctx.Config.TLSCACertFile, err)
		} else {
			ok := rootCAs.AppendCertsFromPEM(caCerts)

			if !ok {
				log.Warningf("No CA certs appended, only using system certs")
			} else {
				log.Debugf("Loaded CA certs from %s", ctx.Config.TLSCACertFile)
			}
		}

		tlsConfig.RootCAs = rootCAs
	}

	if ctx.Config.TLSInsecureSkipVerify != nil {
		tlsConfig.InsecureSkipVerify = *ctx.Config.TLSInsecureSkipVerify // nolint: gosec
	}

	client := retryablehttp.NewClient()
	client.Logger = &LeveledLogger{Logger: logging.Logger.WithField("library", "retryablehttp")}
	client.HTTPClient.Transport.(*http.Transport).TLSClientConfig = &tlsConfig

	return &PricingAPIClient{
		APIClient: APIClient{
			httpClient: client.StandardClient(),
			endpoint:   ctx.Config.PricingAPIEndpoint,
			apiKey:     ctx.Config.APIKey,
			uuid:       ctx.UUID(),
		},
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

	_, err := c.doRequest("POST", "/event", d)
	return err
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

// Batch all the queries for these resources so we can use less GraphQL requests
// Use PriceQueryKeys to keep track of which query maps to which sub-resource and price component.
func (c *PricingAPIClient) BatchRequests(resources []*schema.Resource, batchSize int) []BatchRequest {
	reqs := make([]BatchRequest, 0)

	keys := make([]PriceQueryKey, 0)
	queries := make([]GraphQLQuery, 0)

	for _, r := range resources {
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
	}

	for i := 0; i < len(queries); i += batchSize {
		keysEnd := int64(math.Min(float64(i+batchSize), float64(len(keys))))
		queriesEnd := int64(math.Min(float64(i+batchSize), float64(len(queries))))

		reqs = append(reqs, BatchRequest{keys[i:keysEnd], queries[i:queriesEnd]})
	}

	return reqs
}

func (c *PricingAPIClient) PerformRequest(req BatchRequest) ([]PriceQueryResult, error) {
	log.Debugf("Getting pricing details for %d cost components from %s", len(req.queries), c.endpoint)

	results, err := c.doQueries(req.queries)
	if err != nil {
		return []PriceQueryResult{}, err
	}

	return c.zipQueryResults(req.keys, results), nil
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
