package prices

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/infracost/infracost/pkg/config"
	"github.com/infracost/infracost/pkg/schema"
	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

var InvalidAPIKeyError = errors.New("Invalid API key")

type pricingAPIErrorResponse struct {
	Error string `json:"error"`
}

type queryKey struct {
	Resource      *schema.Resource
	CostComponent *schema.CostComponent
}

type queryResult struct {
	queryKey
	Result gjson.Result
}

type GraphQLQuery struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type QueryRunner interface {
	RunQueries(resource *schema.Resource) ([]queryResult, error)
}

type GraphQLQueryRunner struct {
	endpoint string
}

func NewGraphQLQueryRunner(endpoint string) *GraphQLQueryRunner {
	return &GraphQLQueryRunner{
		endpoint: endpoint,
	}
}

func (q *GraphQLQueryRunner) RunQueries(r *schema.Resource) ([]queryResult, error) {
	keys, queries := q.batchQueries(r)

	if len(queries) == 0 {
		log.Debugf("Skipping getting pricing details for %s since there are no queries to run", r.Name)
		return []queryResult{}, nil
	}

	log.Debugf("Getting pricing details from %s for %s", config.Config.PricingAPIEndpoint, r.Name)

	results, err := q.getQueryResults(queries)
	if err != nil {
		return []queryResult{}, err
	}

	return q.zipQueryResults(keys, results), nil
}

func (q *GraphQLQueryRunner) buildQuery(product *schema.ProductFilter, price *schema.PriceFilter) GraphQLQuery {
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

func (q *GraphQLQueryRunner) getQueryResults(queries []GraphQLQuery) ([]gjson.Result, error) {
	results := make([]gjson.Result, 0, len(queries))

	if len(queries) == 0 {
		log.Debug("skipping GraphQL request as no queries have been specified")
		return results, nil
	}

	queriesBody, err := json.Marshal(queries)
	if err != nil {
		return results, errors.Wrap(err, "error marshalling queries")
	}

	req, err := http.NewRequest("POST", q.endpoint, bytes.NewBuffer([]byte(queriesBody)))
	if err != nil {
		return results, err
	}

	req.Header.Set("content-type", "application/json")
	req.Header.Set("User-Agent", config.GetUserAgent())
	req.Header.Set("X-Api-Key", config.Config.ApiKey)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return results, errors.Wrap(err, "error contacting api")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return results, errors.Wrap(err, "error reading api response")
	}
	if resp.StatusCode != 200 {
		var r pricingAPIErrorResponse
		err = json.Unmarshal(body, &r)
		if err != nil {
			return results, fmt.Errorf("error unmarshalling api response error")
		}
		if r.Error == "Invalid API key" {
			return results, InvalidAPIKeyError
		}
		return results, fmt.Errorf(r.Error)
	}

	results = append(results, gjson.ParseBytes(body).Array()...)

	return results, nil
}

// Batch all the queries for this resource so we can use one GraphQL call
// Use queryKeys to keep track of which query maps to which sub-resource and price component
func (q *GraphQLQueryRunner) batchQueries(r *schema.Resource) ([]queryKey, []GraphQLQuery) {
	keys := make([]queryKey, 0)
	queries := make([]GraphQLQuery, 0)

	for _, c := range r.CostComponents {
		keys = append(keys, queryKey{r, c})
		queries = append(queries, q.buildQuery(c.ProductFilter, c.PriceFilter))
	}

	for _, r := range r.FlattenedSubResources() {
		for _, c := range r.CostComponents {
			keys = append(keys, queryKey{r, c})
			queries = append(queries, q.buildQuery(c.ProductFilter, c.PriceFilter))
		}
	}

	return keys, queries
}

func (q *GraphQLQueryRunner) zipQueryResults(k []queryKey, r []gjson.Result) []queryResult {
	res := make([]queryResult, 0, len(k))

	for i, k := range k {
		res = append(res, queryResult{
			queryKey: k,
			Result:   r[i],
		})
	}

	return res
}
