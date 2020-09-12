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

func (q *GraphQLQueryRunner) RunQueries(resource *schema.Resource) ([]queryResult, error) {
	queryKeys, queries := q.batchQueries(resource)

	log.Debugf("Getting pricing details from %s for %s", config.Config.ApiUrl, resource.Name)
	results, err := q.getQueryResults(queries)
	if err != nil {
		return []queryResult{}, err
	}

	return q.zipQueryResults(queryKeys, results), nil
}

func (q *GraphQLQueryRunner) buildQuery(productFilter *schema.ProductFilter, priceFilter *schema.PriceFilter) GraphQLQuery {
	variables := map[string]interface{}{}
	variables["productFilter"] = productFilter
	variables["priceFilter"] = priceFilter

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
	return GraphQLQuery{query, variables}
}

func (q *GraphQLQueryRunner) getQueryResults(queries []GraphQLQuery) ([]gjson.Result, error) {
	results := make([]gjson.Result, 0, len(queries))

	queriesBody, err := json.Marshal(queries)
	if err != nil {
		return results, err
	}

	client := http.Client{}
	resp, err := client.Post(q.endpoint, "application/json", bytes.NewBuffer([]byte(queriesBody)))
	if err != nil {
		return results, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return results, err
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
