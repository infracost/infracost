package apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/infracost/infracost/internal/config"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

type APIClient struct {
	endpoint string
	apiKey   string
}

type GraphQLQuery struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type GraphQLError struct {
	err error
	msg string
}

func (e *GraphQLError) Error() string {
	return fmt.Sprintf("%s: %v", e.msg, e.err.Error())
}

type GraphQLErrorResponse struct {
	Error string `json:"error"`
}

var ErrInvalidAPIKey = errors.New("Invalid API key")

func (c *APIClient) doQueries(queries []GraphQLQuery) ([]gjson.Result, error) {
	if len(queries) == 0 {
		log.Debug("Skipping GraphQL request as no queries have been specified")
		return []gjson.Result{}, nil
	}

	reqBody, err := json.Marshal(queries)
	if err != nil {
		return []gjson.Result{}, errors.Wrap(err, "Error generating GraphQL query body")
	}

	req, err := http.NewRequest("POST", c.endpoint+"/graphql", bytes.NewBuffer(reqBody))
	if err != nil {
		return []gjson.Result{}, errors.Wrap(err, "Error generating GraphQL request")
	}

	config.AddAuthHeaders(c.apiKey, req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []gjson.Result{}, errors.Wrap(err, "Error sending API request")
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []gjson.Result{}, &GraphQLError{err, "Invalid API response"}
	}

	if resp.StatusCode != 200 {
		var r GraphQLErrorResponse
		err = json.Unmarshal(respBody, &r)
		if err != nil {
			return []gjson.Result{}, &GraphQLError{err, "Invalid API response"}
		}

		if r.Error == "Invalid API key" {
			return []gjson.Result{}, ErrInvalidAPIKey
		}

		return []gjson.Result{}, &GraphQLError{errors.New(r.Error), "Received error from API"}
	}

	return gjson.ParseBytes(respBody).Array(), nil
}
