package base

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"infracost/pkg/config"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

type GraphQLQuery struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type ResourceQueryResultMap map[*Resource]map[*PriceComponent]gjson.Result

type QueryRunner interface {
	RunQueries(resource Resource) (ResourceQueryResultMap, error)
}

type GraphQLQueryRunner struct {
	endpoint string
}

func NewGraphQLQueryRunner(endpoint string) *GraphQLQueryRunner {
	return &GraphQLQueryRunner{
		endpoint: endpoint,
	}
}

func (r *GraphQLQueryRunner) buildQuery(filters []Filter) GraphQLQuery {
	variables := map[string]interface{}{}
	variables["filter"] = map[string]interface{}{}
	variables["filter"].(map[string]interface{})["attributes"] = filters

	query := `
		query($filter: ProductFilter!) {
			products(
				filter: $filter,
			) {
				onDemandPricing {
					priceDimensions {
						unit
						pricePerUnit {
							USD
						}
					}
				}
			}
		}
	`
	return GraphQLQuery{query, variables}
}

func (r *GraphQLQueryRunner) getQueryResults(queries []GraphQLQuery) ([]gjson.Result, error) {
	results := make([]gjson.Result, 0, len(queries))

	queriesBody, err := json.Marshal(queries)
	if err != nil {
		return results, err
	}

	client := http.Client{}
	resp, err := client.Post(r.endpoint, "application/json", bytes.NewBuffer([]byte(queriesBody)))
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
func (r *GraphQLQueryRunner) batchQueries(resource Resource) ([]queryKey, []GraphQLQuery) {
	queryKeys := make([]queryKey, 0)
	queries := make([]GraphQLQuery, 0)

	for _, priceComponent := range resource.PriceComponents() {
		queryKeys = append(queryKeys, queryKey{resource, priceComponent})
		queries = append(queries, r.buildQuery(priceComponent.Filters()))
	}

	for _, subResource := range FlattenSubResources(resource) {
		for _, priceComponent := range subResource.PriceComponents() {
			queryKeys = append(queryKeys, queryKey{subResource, priceComponent})
			queries = append(queries, r.buildQuery(priceComponent.Filters()))
		}
	}

	return queryKeys, queries
}

// Unpack the query results into a map so we can find by resource and price component
func (r *GraphQLQueryRunner) unpackQueryResults(queryKeys []queryKey, queryResults []gjson.Result) ResourceQueryResultMap {
	resourceResults := make(ResourceQueryResultMap)

	for i, queryResult := range queryResults {
		resource := queryKeys[i].Resource
		priceComponent := queryKeys[i].PriceComponent

		if _, ok := resourceResults[&resource]; !ok {
			resourceResults[&resource] = make(map[*PriceComponent]gjson.Result)
		}
		resourceResults[&resource][&priceComponent] = queryResult
	}

	return resourceResults
}

func (r *GraphQLQueryRunner) RunQueries(resource Resource) (ResourceQueryResultMap, error) {
	queryKeys, queries := r.batchQueries(resource)

	log.Debugf("Getting pricing details from %s for %s", config.Config.ApiUrl, resource.Address())
	queryResults, err := r.getQueryResults(queries)
	if err != nil {
		return ResourceQueryResultMap{}, err
	}

	return r.unpackQueryResults(queryKeys, queryResults), nil
}
