package apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/version"
)

type APIClient struct {
	httpClient *http.Client
	endpoint   string
	apiKey     string
	uuid       uuid.UUID
}

type GraphQLQuery struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type APIError struct {
	err error
	msg string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %v", e.msg, e.err.Error())
}

type APIErrorResponse struct {
	Error string `json:"error"`
}

var ErrInvalidAPIKey = errors.New("Invalid API key")

func (c *APIClient) doQueries(queries []GraphQLQuery) ([]gjson.Result, error) {
	if len(queries) == 0 {
		log.Debug("Skipping GraphQL request as no queries have been specified")
		return []gjson.Result{}, nil
	}

	respBody, err := c.doRequest("POST", "/graphql", queries)
	return gjson.ParseBytes(respBody).Array(), err
}

func (c *APIClient) doRequest(method string, path string, d interface{}) ([]byte, error) {
	logging.Logger.Debugf("'%s' request to '%s' using trace_id: '%s'", method, path, c.uuid.String())

	reqBody, err := json.Marshal(d)
	if err != nil {
		return []byte{}, errors.Wrap(err, "Error generating request body")
	}

	req, err := http.NewRequest(method, c.endpoint+path, bytes.NewBuffer(reqBody))
	if err != nil {
		return []byte{}, errors.Wrap(err, "Error generating request")
	}

	c.AddAuthHeaders(req)

	client := c.httpClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, errors.Wrap(err, "Error sending API request")
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, &APIError{err, "Invalid API response"}
	}

	if resp.StatusCode != 200 {
		var r APIErrorResponse

		err = json.Unmarshal(respBody, &r)
		if err != nil {
			return []byte{}, &APIError{fmt.Errorf(resp.Status), "Invalid API response"}
		}

		if r.Error == "Invalid API key" {
			return []byte{}, ErrInvalidAPIKey
		}
		return []byte{}, &APIError{fmt.Errorf("%v %v", resp.Status, r.Error), "Received error from API"}
	}

	return respBody, nil
}

func (c *APIClient) AddDefaultHeaders(req *http.Request) {
	req.Header.Set("content-type", "application/json")
	req.Header.Set("User-Agent", userAgent())
}

func (c *APIClient) AddAuthHeaders(req *http.Request) {
	c.AddDefaultHeaders(req)
	req.Header.Set("X-Api-Key", c.apiKey)
	if c.uuid != uuid.Nil {
		req.Header.Set("X-Infracost-Trace-Id", fmt.Sprintf("cli=%s", c.uuid.String()))
	}
}

func userAgent() string {
	userAgent := "infracost"

	if version.Version != "" {
		userAgent += fmt.Sprintf("-%s", version.Version)
	}

	return userAgent
}
