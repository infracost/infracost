package apiclient

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	json "github.com/json-iterator/go"

	"github.com/google/uuid"
	"github.com/pkg/errors"
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
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

var (
	ErrorCodeExceededQuota = "above_quota"
	ErrorCodeAPIKeyInvalid = "invalid_api_key"
)

// APIError defines an api error that is designed to be showed to the user in a
// output form (comment/stdout/html).
type APIError struct {
	err error
	// Msg defines a human-readable string that is safe to show to the user
	// to give more context about an error.
	Msg string
	// Code is the original StatusCode of the error.
	Code int
	// ErrorCode is the internal error code that can accompany errors from different status codes.
	ErrorCode string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %v", e.Msg, e.err.Error())
}

type APIErrorResponse struct {
	Error     string `json:"error"`
	ErrorCode string `json:"error_code"`
}

func (c *APIClient) DoQueries(queries []GraphQLQuery) ([]gjson.Result, error) {
	if len(queries) == 0 {
		logging.Logger.Debug().Msg("Skipping GraphQL request as no queries have been specified")
		return []gjson.Result{}, nil
	}

	respBody, err := c.doRequest("POST", "/graphql", queries)
	return gjson.ParseBytes(respBody).Array(), err
}

func (c *APIClient) doRequest(method string, path string, d any) ([]byte, error) {
	logging.Logger.Debug().Msgf("'%s' request to '%s' using trace_id: '%s'", method, path, c.uuid.String())

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
		return []byte{}, &APIError{err: err, Msg: fmt.Sprintf("Invalid API response %s %s", method, path)}
	}

	if resp.StatusCode != 200 {
		var r APIErrorResponse

		err = json.Unmarshal(respBody, &r)
		if err != nil {
			return []byte{}, &APIError{err: err, Msg: fmt.Sprintf("Invalid API response %q %q body: %q", method, path, respBody), Code: resp.StatusCode}
		}

		if r.ErrorCode != "" {
			return []byte{}, &APIError{err: fmt.Errorf("%v %v", resp.Status, r.Error), Msg: r.Error, Code: resp.StatusCode, ErrorCode: r.ErrorCode}
		}

		// handle legacy cloud pricing apis which don't have the new `error_code` field.
		if r.Error == "Invalid API key" {
			return []byte{}, &APIError{err: fmt.Errorf("%v %v", resp.Status, r.Error), Msg: "Invalid API Key", Code: resp.StatusCode, ErrorCode: ErrorCodeAPIKeyInvalid}
		}

		return []byte{}, &APIError{err: fmt.Errorf("%v %v", resp.Status, r.Error), Msg: "Received error from API", Code: resp.StatusCode}
	}

	return respBody, nil
}

func (c *APIClient) AddDefaultHeaders(req *http.Request) {
	req.Header.Set("content-type", "application/json")
	req.Header.Set("User-Agent", userAgent())
}

func (c *APIClient) AddAuthHeaders(req *http.Request) {
	c.AddDefaultHeaders(req)
	if strings.HasPrefix(c.apiKey, "ics") {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	} else {
		req.Header.Set("X-Api-Key", c.apiKey)
	}

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
