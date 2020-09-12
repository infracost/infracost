package prices

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/infracost/infracost/pkg/config"
	"github.com/infracost/infracost/pkg/schema"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

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

	log.Debugf("Getting pricing details from %s for %s", config.Config.ApiUrl, r.Name)

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

	queriesBody, err := json.Marshal(queries)
	if err != nil {
		return results, errors.Wrap(err, "error marshaling queries")
	}

	client := http.Client{}
	resp, err := client.Post(q.endpoint, "application/json", bytes.NewBuffer([]byte(queriesBody)))
	if err != nil {
		return results, errors.Wrap(err, "error contacting api")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return results, errors.Wrap(err, "error reading api response")
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

func (q *GraphQLQueryRunner) zipQueryResults(queryKeys []queryKey, results []gjson.Result) []queryResult {
	queryResults := make([]queryResult, 0, len(queryKeys))

	for i, queryKey := range queryKeys {
		queryResults = append(queryResults, queryResult{
			queryKey: queryKey,
			Result:   results[i],
		})
	}

	return queryResults
}
