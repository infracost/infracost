package costs

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"infracost/pkg/config"
	"infracost/pkg/resource"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

type GraphQLQuery struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type ResourceQueryResultMap map[resource.Resource]map[resource.PriceComponent]gjson.Result

type QueryRunner interface {
	RunQueries(r resource.Resource) (ResourceQueryResultMap, error)
}

type GraphQLQueryRunner struct {
	endpoint string
}

func NewGraphQLQueryRunner(endpoint string) *GraphQLQueryRunner {
	return &GraphQLQueryRunner{
		endpoint: endpoint,
	}
}

func (q *GraphQLQueryRunner) buildQuery(productFilter *resource.ProductFilter, priceFilter *resource.PriceFilter) GraphQLQuery {
	variables := map[string]interface{}{}
	variables["productFilter"] = productFilter
	variables["priceFilter"] = priceFilter

	query := `
		query($productFilter: ProductFilter!, $priceFilter: PriceFilter) {
			products(filter: $productFilter) {
				prices(filter: $priceFilter) {
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
func (q *GraphQLQueryRunner) batchQueries(r resource.Resource) ([]queryKey, []GraphQLQuery) {
	queryKeys := make([]queryKey, 0)
	queries := make([]GraphQLQuery, 0)

	for _, priceComponent := range r.PriceComponents() {
		queryKeys = append(queryKeys, queryKey{r, priceComponent})
		queries = append(queries, q.buildQuery(priceComponent.ProductFilter(), priceComponent.PriceFilter()))
	}

	for _, subResource := range resource.FlattenSubResources(r) {
		for _, priceComponent := range subResource.PriceComponents() {
			queryKeys = append(queryKeys, queryKey{subResource, priceComponent})
			queries = append(queries, q.buildQuery(priceComponent.ProductFilter(), priceComponent.PriceFilter()))
		}
	}

	return queryKeys, queries
}

// Unpack the query results into a map so we can find by resource and price component
func (q *GraphQLQueryRunner) unpackQueryResults(queryKeys []queryKey, queryResults []gjson.Result) ResourceQueryResultMap {
	resourceResults := make(ResourceQueryResultMap)

	for i, queryResult := range queryResults {
		r := queryKeys[i].Resource
		priceComponent := queryKeys[i].PriceComponent

		if _, ok := resourceResults[r]; !ok {
			resourceResults[r] = make(map[resource.PriceComponent]gjson.Result)
		}
		resourceResults[r][priceComponent] = queryResult
	}

	return resourceResults
}

func (q *GraphQLQueryRunner) RunQueries(r resource.Resource) (ResourceQueryResultMap, error) {
	queryKeys, queries := q.batchQueries(r)

	log.Debugf("Getting pricing details from %s for %s", config.Config.ApiUrl, r.Address())
	queryResults, err := q.getQueryResults(queries)
	if err != nil {
		return ResourceQueryResultMap{}, err
	}

	return q.unpackQueryResults(queryKeys, queryResults), nil
}
