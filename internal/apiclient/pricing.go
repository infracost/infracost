package apiclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/infracost/infracost/internal/config"
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

func NewPricingAPIClient(ctx *config.RunContext) *PricingAPIClient {
	currency := ctx.Config.Currency
	if currency == "" {
		currency = "USD"
	}

	tlsConfig := tls.Config{
		MinVersion: tls.VersionTLS12,
	} // nolint: gosec

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

	// disallow this setting
	// if ctx.Config.TLSInsecureSkipVerify != nil {
	//	tlsConfig.InsecureSkipVerify = *ctx.Config.TLSInsecureSkipVerify
	//}

	var iamAuthenticator *core.IamAuthenticator = nil
	authenticatorBuilder := core.NewIamAuthenticatorBuilder()
	if ctx.Config.IBMCloudIAMUrl != "" {
		fmt.Println("Configured IAM URL", ctx.Config.IBMCloudIAMUrl)
		authenticatorBuilder.SetURL(ctx.Config.IBMCloudIAMUrl)
	} else {
		fmt.Println("No IBM_CLOUD_IAM_URL credential set, defaults to production.")
	}
	if ctx.Config.IBMCloudApiKey != "" {
		if len(ctx.Config.IBMCloudApiKey) != 44 {
			fmt.Println("IBM_CLOUD_API_KEY's length is not 44... Is this a proper IAM api key?")
		}
		authenticatorBuilder.SetApiKey(ctx.Config.IBMCloudApiKey)
		authenticator, err := authenticatorBuilder.Build()
		if err != nil {
			log.Error("Unable to init authenticator", err)
		}
		iamAuthenticator = authenticator
	} else {
		fmt.Println("No IBM_CLOUD_API_KEY credential set")
	}

	if ctx.Config.APIKey == "" && iamAuthenticator == nil {
		fmt.Println("No authentication method specified")
	}

	return &PricingAPIClient{
		APIClient: APIClient{
			endpoint:         ctx.Config.PricingAPIEndpoint,
			apiKey:           ctx.Config.APIKey,
			ibmAuthenticator: iamAuthenticator,
			tlsConfig:        &tlsConfig,
			uuid:             ctx.UUID(),
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

	query := fmt.Sprintf(`
		query($productFilter: ProductFilter!, $priceFilter: PriceFilter) {
			products(filter: $productFilter) {
				prices(filter: $priceFilter) {
					priceHash
					%s
					startUsageAmount
					endUsageAmount
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
