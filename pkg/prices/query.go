package prices

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/infracost/infracost/pkg/config"
	"github.com/infracost/infracost/pkg/schema"
	"github.com/infracost/infracost/pkg/version"

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

	req, err := http.NewRequest("POST", q.endpoint, bytes.NewBuffer([]byte(queriesBody)))
	if err != nil {
		return results, err
	}

	req.Header.Set("content-type", "application/json")
	req.Header.Set("User-Agent", getUserAgent())

	client := http.Client{}
	resp, err := client.Do(req)
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
func (q *GraphQLQueryRunner) batchQueries(resource *schema.Resource) ([]queryKey, []GraphQLQuery) {
	queryKeys := make([]queryKey, 0)
	queries := make([]GraphQLQuery, 0)

	for _, costComponent := range resource.CostComponents {
		queryKeys = append(queryKeys, queryKey{resource, costComponent})
		queries = append(queries, q.buildQuery(costComponent.ProductFilter, costComponent.PriceFilter))
	}

	for _, subResource := range resource.FlattenedSubResources() {
		for _, costComponent := range subResource.CostComponents {
			queryKeys = append(queryKeys, queryKey{subResource, costComponent})
			queries = append(queries, q.buildQuery(costComponent.ProductFilter, costComponent.PriceFilter))
		}
	}

	return queryKeys, queries
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

func getUserAgent() string {
	userAgent := "infracost"
	if version.Version != "" {
		userAgent += fmt.Sprintf("-%s", version.Version)

	}
	infracostEnv := getInfracostEnv()

	if infracostEnv != "" {
		userAgent += fmt.Sprintf(" (%s)", infracostEnv)
	}

	return userAgent
}

func getInfracostEnv() string {
	if os.Getenv("INFRACOST_ENV") == "test" || isTesting() {
		return "test"
	} else if os.Getenv("INFRACOST_ENV") == "dev" {
		return "dev"
	}
	return ""
}

func isTesting() bool {
	return strings.HasSuffix(os.Args[0], ".test")
}
