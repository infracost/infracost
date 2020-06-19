package base

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/tidwall/gjson"
)

type GraphQLQuery struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

func BuildQuery(filters []Filter) GraphQLQuery {
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

func GetQueryResults(queries []GraphQLQuery) ([]gjson.Result, error) {
	results := make([]gjson.Result, 0, len(queries))

	queriesBody, err := json.Marshal(queries)
	if err != nil {
		return results, err
	}

	client := http.Client{}
	resp, err := client.Post("http://localhost:4000", "application/json", bytes.NewBuffer([]byte(queriesBody)))
	if err != nil {
		return results, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return results, err
	}

	for _, result := range gjson.ParseBytes(body).Array() {
		results = append(results, result)
	}

	return results, nil
}
