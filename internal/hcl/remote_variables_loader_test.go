package hcl

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/rs/zerolog"
	"github.com/spacelift-io/spacectl/client"
	"github.com/spacelift-io/spacectl/client/session"
	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/vcs"
)

func TestSpaceliftRemoteVariableLoader_Load(t *testing.T) {
	type fields struct {
		Metadata vcs.Metadata
	}
	type args struct {
		options RemoteVarLoaderOptions
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		want           map[string]cty.Value
		wantErr        assert.ErrorAssertionFunc
		testServerFunc func(t *testing.T) http.HandlerFunc
		verifyLog      func(t *testing.T, log string)
	}{
		{
			name: "return variables from SpaceLift",
			fields: fields{
				Metadata: vcs.Metadata{
					Remote: vcs.Remote{
						Name: "test/test",
					},
				},
			},
			args: args{
				options: RemoteVarLoaderOptions{
					ModulePath:  "module/path",
					Environment: "dev",
				},
			},
			want: map[string]cty.Value{
				"context_var": cty.StringVal("context_value"),
				"config_var":  cty.StringVal("config_value"),
				"runtime_var": cty.StringVal("runtime_value"),
			},
			wantErr: assert.NoError,
			testServerFunc: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					assert.Equal(t, "POST", r.Method)
					s, err := io.ReadAll(r.Body)
					assert.NoError(t, err)

					expectedJSON := `{
						"query": "query($input:SearchInput!){searchStacks(input: $input){edges{node{id,name,projectRoot,config{id,value},runtimeConfig{element{id,value}},attachedContexts{id,contextName}}},pageInfo{endCursor,hasNextPage,hasPreviousPage}}}",
						"variables": {
							"input": {
								"first": 10,
								"after": null,
								"fullTextSearch": null,
								"orderBy": null,
								"predicates": [
									{"field": "repository", "constraint": {"stringMatches": ["test/test"], "booleanEquals": null, "enumEquals": null}},
									{"field": "projectRoot", "constraint": {"stringMatches": ["module/path"], "booleanEquals": null, "enumEquals": null}}
								]
							}
						}
					}`
					assert.JSONEq(t, expectedJSON, string(s))

					response := `{
						"data": {
							"searchStacks": {
								"edges": [
									{
										"node": {
											"id": "module-path-dev",
											"name": "module-path:dev",
											"projectRoot": "module/path",
											"attachedContexts": [
												{"id": "TF_VAR_context_var", "contextName": "context_value"},
												{"id": "CONTEXT_VAR", "contextName": "non_tf_context_value"}
											],
											"config": [
												{"id": "TF_VAR_config_var", "value": "config_value"},
												{"id": "CONFIG_VAR", "value": "non_tf_config_value"}
											],
											"runtimeConfig": [
												{"element": {"id": "TF_VAR_runtime_var", "value": "runtime_value"}},
												{"element": {"id": "RUNTIME_VAR", "value": "non_tf_runtime_value"}}
											]
										}
									}
								],
								"pageInfo": {
									"endCursor": "MDFKNEtQMzVETjFYRDcxNTY1UkJCMzBITUQ=",
									"hasNextPage": false,
									"hasPreviousPage": false
								}
							}
						}
					}`
					_, err = w.Write([]byte(response))
					assert.NoError(t, err)
				}
			},
		},
		{
			name: "returns the correct variables when multiple stacks match ModulePath but only one has the correct environment",
			fields: fields{
				Metadata: vcs.Metadata{
					Remote: vcs.Remote{
						Name: "test/test",
					},
				},
			},
			args: args{
				options: RemoteVarLoaderOptions{
					ModulePath:  "module/path",
					Environment: "prod",
				},
			},
			want: map[string]cty.Value{
				"runtime_var_prod": cty.StringVal("runtime_value_prod"),
				"config_var_prod":  cty.StringVal("config_value_prod"),
				"context_var_prod": cty.StringVal("context_value_prod"),
			},
			wantErr: assert.NoError,
			testServerFunc: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					assert.Equal(t, "POST", r.Method)
					s, err := io.ReadAll(r.Body)
					assert.NoError(t, err)

					expectedJSON := `{
						"query": "query($input:SearchInput!){searchStacks(input: $input){edges{node{id,name,projectRoot,config{id,value},runtimeConfig{element{id,value}},attachedContexts{id,contextName}}},pageInfo{endCursor,hasNextPage,hasPreviousPage}}}",
						"variables": {
							"input": {
								"first": 10,
								"after": null,
								"fullTextSearch": null,
								"orderBy": null,
								"predicates": [
									{"field": "repository", "constraint": {"stringMatches": ["test/test"], "booleanEquals": null, "enumEquals": null}},
									{"field": "projectRoot", "constraint": {"stringMatches": ["module/path"], "booleanEquals": null, "enumEquals": null}}
								]
							}
						}
					}`
					assert.JSONEq(t, expectedJSON, string(s))

					response := `{
						"data": {
							"searchStacks": {
								"edges": [
									{
										"node": {
											"id": "module-path-staging",
											"name": "module-path-:staging",
											"projectRoot": "module/path",
											"attachedContexts": [
												{"id": "TF_VAR_context_var_staging", "contextName": "context_value_staging"}
											],
											"config": [
												{"id": "TF_VAR_config_var_staging", "value": "config_value_staging"}
											],
											"runtimeConfig": [
												{"element": {"id": "TF_VAR_runtime_var_staging", "value": "runtime_value_staging"}}
											]
										}
									},
									{
										"node": {
											"id": "module-path-prod",
											"name": "module-path:prod",
											"projectRoot": "module/path",
											"attachedContexts": [
												{"id": "TF_VAR_context_var_prod", "contextName": "context_value_prod"}
											],
											"config": [
												{"id": "TF_VAR_config_var_prod", "value": "config_value_prod"}
											],
											"runtimeConfig": [
												{"element": {"id": "TF_VAR_runtime_var_prod", "value": "runtime_value_prod"}}
											]
										}
									}
								],
								"pageInfo": {
									"endCursor": "MDFKNEtQMzVETjFYRDcxNTY1UkJCMzBITUQ=",
									"hasNextPage": false,
									"hasPreviousPage": false
								}
							}
						}
					}`
					_, err = w.Write([]byte(response))
					assert.NoError(t, err)
				}
			},
		},
		{
			name: "returns the correct variables according to the precedence of the config and context values",
			fields: fields{
				Metadata: vcs.Metadata{
					Remote: vcs.Remote{
						Name: "test/test",
					},
				},
			},
			args: args{
				options: RemoteVarLoaderOptions{
					ModulePath:  "module/path",
					Environment: "dev",
				},
			},
			want: map[string]cty.Value{
				"var1": cty.StringVal("runtime_value"),
				"var2": cty.StringVal("config_value"),
			},
			wantErr: assert.NoError,
			testServerFunc: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					assert.Equal(t, "POST", r.Method)
					s, err := io.ReadAll(r.Body)
					assert.NoError(t, err)

					expectedJSON := `{
						"query": "query($input:SearchInput!){searchStacks(input: $input){edges{node{id,name,projectRoot,config{id,value},runtimeConfig{element{id,value}},attachedContexts{id,contextName}}},pageInfo{endCursor,hasNextPage,hasPreviousPage}}}",
						"variables": {
							"input": {
								"first": 10,
								"after": null,
								"fullTextSearch": null,
								"orderBy": null,
								"predicates": [
									{"field": "repository", "constraint": {"stringMatches": ["test/test"], "booleanEquals": null, "enumEquals": null}},
									{"field": "projectRoot", "constraint": {"stringMatches": ["module/path"], "booleanEquals": null, "enumEquals": null}}
								]
							}
						}
					}`
					assert.JSONEq(t, expectedJSON, string(s))

					response := `{
						"data": {
							"searchStacks": {
								"edges": [
									{
										"node": {
											"id": "module-path-dev",
											"name": "module-path:dev",
											"projectRoot": "module/path",
											"attachedContexts": [
												{"id": "TF_VAR_var1", "contextName": "context_value"},
												{"id": "TF_VAR_var2", "contextName": "context_value"}
											],
											"config": [
												{"id": "TF_VAR_var1", "value": "config_value"},
												{"id": "TF_VAR_var2", "value": "config_value"}
											],
											"runtimeConfig": [
												{"element": {"id": "TF_VAR_var1", "value": "runtime_value"}}
											]
										}
									}
								],
								"pageInfo": {
									"endCursor": null,
									"hasNextPage": false,
									"hasPreviousPage": false
								}
							}
						}
					}`
					_, err = w.Write([]byte(response))
					assert.NoError(t, err)
				}
			},
		},
		{
			name: "returns nil if no stacks match ModulePath",
			fields: fields{
				Metadata: vcs.Metadata{
					Remote: vcs.Remote{
						Name: "test/test",
					},
				},
			},
			args: args{
				options: RemoteVarLoaderOptions{
					ModulePath:  "nonexistent/path",
					Environment: "prod",
				},
			},
			want:    nil,
			wantErr: assert.NoError,
			testServerFunc: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					assert.Equal(t, "POST", r.Method)
					s, err := io.ReadAll(r.Body)
					assert.NoError(t, err)

					expectedJSON := `{
						"query": "query($input:SearchInput!){searchStacks(input: $input){edges{node{id,name,projectRoot,config{id,value},runtimeConfig{element{id,value}},attachedContexts{id,contextName}}},pageInfo{endCursor,hasNextPage,hasPreviousPage}}}",
						"variables": {
							"input": {
								"first": 10,
								"after": null,
								"fullTextSearch": null,
								"orderBy": null,
								"predicates": [
									{"field": "repository", "constraint": {"stringMatches": ["test/test"], "booleanEquals": null, "enumEquals": null}},
									{"field": "projectRoot", "constraint": {"stringMatches": ["nonexistent/path"], "booleanEquals": null, "enumEquals": null}}
								]
							}
						}
					}`
					assert.JSONEq(t, expectedJSON, string(s))

					response := `{
						"data": {
							"searchStacks": {
								"edges": [],
								"pageInfo": {
									"endCursor": null,
									"hasNextPage": false,
									"hasPreviousPage": false
								}
							}
						}
					}`
					_, err = w.Write([]byte(response))
					assert.NoError(t, err)
				}
			},
		},
		{
			name: "returns nil if ModulePath is missing",
			fields: fields{
				Metadata: vcs.Metadata{
					Remote: vcs.Remote{
						Name: "test/test",
					},
				},
			},
			args: args{
				options: RemoteVarLoaderOptions{
					ModulePath: "",
				},
			},
			want:    nil,
			wantErr: assert.NoError,
			testServerFunc: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					// This test should not reach the server as ModulePath is empty
					assert.FailNow(t, "HTTP server should not be called when ModulePath is empty")
				}
			},
		},
		{
			name: "warns if ModulePath matches but no environment matches",
			fields: fields{
				Metadata: vcs.Metadata{
					Remote: vcs.Remote{
						Name: "test/test",
					},
				},
			},
			args: args{
				options: RemoteVarLoaderOptions{
					ModulePath:  "module/path",
					Environment: "nonexistent-env",
				},
			},
			want:    nil,
			wantErr: assert.NoError,
			testServerFunc: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					assert.Equal(t, "POST", r.Method)
					s, err := io.ReadAll(r.Body)
					assert.NoError(t, err)

					expectedJSON := `{
						"query": "query($input:SearchInput!){searchStacks(input: $input){edges{node{id,name,projectRoot,config{id,value},runtimeConfig{element{id,value}},attachedContexts{id,contextName}}},pageInfo{endCursor,hasNextPage,hasPreviousPage}}}",
						"variables": {
							"input": {
								"first": 10,
								"after": null,
								"fullTextSearch": null,
								"orderBy": null,
								"predicates": [
									{"field": "repository", "constraint": {"stringMatches": ["test/test"], "booleanEquals": null, "enumEquals": null}},
									{"field": "projectRoot", "constraint": {"stringMatches": ["module/path"], "booleanEquals": null, "enumEquals": null}}
								]
							}
						}
					}`
					assert.JSONEq(t, expectedJSON, string(s))

					response := `{
						"data": {
							"searchStacks": {
								"edges": [
									{
										"node": {
											"id": "module-path-dev",
											"name": "module-path:dev",
											"projectRoot": "module/path",
											"attachedContexts": [
												{"id": "TF_VAR_context_var", "contextName": "context_value"}
											],
											"config": [
												{"id": "TF_VAR_config_var", "value": "config_value"}
											],
											"runtimeConfig": [
												{"element": {"id": "TF_VAR_runtime_var", "value": "runtime_value"}}
											]
										}
									}
								],
								"pageInfo": {
									"endCursor": null,
									"hasNextPage": false,
									"hasPreviousPage": false
								}
							}
						}
					}`
					_, err = w.Write([]byte(response))
					assert.NoError(t, err)
				}
			},
			verifyLog: func(t *testing.T, logOutput string) {
				assert.Contains(t, logOutput, "no Spacelift stack found for module path")
			},
		},
		{
			name: "warns if multiple stacks are found and no environment is specified",
			fields: fields{
				Metadata: vcs.Metadata{
					Remote: vcs.Remote{
						Name: "test/test",
					},
				},
			},
			args: args{
				options: RemoteVarLoaderOptions{
					ModulePath:  "module/path",
					Environment: "",
				},
			},
			want:    nil,
			wantErr: assert.NoError,
			testServerFunc: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					assert.Equal(t, "POST", r.Method)
					s, err := io.ReadAll(r.Body)
					assert.NoError(t, err)

					expectedJSON := `{
						"query": "query($input:SearchInput!){searchStacks(input: $input){edges{node{id,name,projectRoot,config{id,value},runtimeConfig{element{id,value}},attachedContexts{id,contextName}}},pageInfo{endCursor,hasNextPage,hasPreviousPage}}}",
						"variables": {
							"input": {
								"first": 10,
								"after": null,
								"fullTextSearch": null,
								"orderBy": null,
								"predicates": [
									{"field": "repository", "constraint": {"stringMatches": ["test/test"], "booleanEquals": null, "enumEquals": null}},
									{"field": "projectRoot", "constraint": {"stringMatches": ["module/path"], "booleanEquals": null, "enumEquals": null}}
								]
							}
						}
					}`
					assert.JSONEq(t, expectedJSON, string(s))

					response := `{
						"data": {
							"searchStacks": {
								"edges": [
									{
										"node": {
											"id": "module-path-stack1",
											"name": "module-path-stack1",
											"projectRoot": "module/path",
											"attachedContexts": [
												{"id": "TF_VAR_var1", "contextName": "context_value1"}
											],
											"config": [
												{"id": "TF_VAR_var1", "value": "config_value1"}
											],
											"runtimeConfig": [
												{"element": {"id": "TF_VAR_var1", "value": "runtime_value1"}}
											]
										}
									},
									{
										"node": {
											"id": "module-path-stack2",
											"name": "module-path-stack2",
											"projectRoot": "module/path",
											"attachedContexts": [
												{"id": "TF_VAR_var2", "contextName": "context_value2"}
											],
											"config": [
												{"id": "TF_VAR_var2", "value": "config_value2"}
											],
											"runtimeConfig": [
												{"element": {"id": "TF_VAR_var2", "value": "runtime_value2"}}
											]
										}
									}
								],
								"pageInfo": {
									"endCursor": null,
									"hasNextPage": false,
									"hasPreviousPage": false
								}
							}
						}
					}`
					_, err = w.Write([]byte(response))
					assert.NoError(t, err)
				}
			},
			verifyLog: func(t *testing.T, logOutput string) {
				assert.Contains(t, logOutput, "found multiple Spacelift stacks for module path")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.testServerFunc(t))
			defer ts.Close()

			var logBuf bytes.Buffer
			log := zerolog.New(&logBuf).With().Timestamp().Logger()

			oldLogger := logging.Logger
			logging.Logger = log
			defer func() { logging.Logger = oldLogger }()

			s := &SpaceliftRemoteVariableLoader{
				Client: client.New(http.DefaultClient, &stubSession{
					token:    "test",
					endpoint: ts.URL,
				}),
				Metadata: tt.fields.Metadata,
				cache:    &sync.Map{},
				hitMap:   &sync.Map{},
			}

			got, err := s.Load(tt.args.options)
			if !tt.wantErr(t, err, fmt.Sprintf("Load(%v)", tt.args.options)) {
				return
			}
			assert.Equalf(t, tt.want, got, "Load(%v)", tt.args.options)

			if tt.verifyLog != nil {
				tt.verifyLog(t, logBuf.String())
			}
		})
	}
}

func TestSpaceliftRemoteVariableLoader_Load_ReturnsNilIfAlreadyLoaded(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This server should only be called once
		w.Header().Set("Content-Type", "application/json")
		response := `{
			"data": {
				"searchStacks": {
					"edges": [
						{
							"node": {
								"id": "module-path-dev",
								"name": "module-path:dev",
								"projectRoot": "module/path",
								"attachedContexts": [
									{"id": "TF_VAR_test", "contextName": "test_value"}
								],
								"config": [],
								"runtimeConfig": []
							}
						}
					],
					"pageInfo": {
						"endCursor": null,
						"hasNextPage": false,
						"hasPreviousPage": false
					}
				}
			}
		}`
		_, _ = w.Write([]byte(response))
	}))
	defer ts.Close()

	s := &SpaceliftRemoteVariableLoader{
		Client: client.New(http.DefaultClient, &stubSession{
			token:    "test",
			endpoint: ts.URL,
		}),
		Metadata: vcs.Metadata{
			Remote: vcs.Remote{
				Name: "test/test",
			},
		},
		cache:  &sync.Map{},
		hitMap: &sync.Map{},
	}

	options := RemoteVarLoaderOptions{
		ModulePath:  "module/path",
		Environment: "dev",
	}

	// First call should return variables
	got, err := s.Load(options)
	assert.NoError(t, err)
	assert.Equal(t, map[string]cty.Value{
		"test": cty.StringVal("test_value"),
	}, got)

	// Second call should return nil
	got, err = s.Load(options)
	assert.NoError(t, err)
	assert.Nil(t, got)
}

type stubSession struct {
	token    string
	endpoint string
}

func (s stubSession) BearerToken(ctx context.Context) (string, error) {
	return s.token, nil
}

func (s stubSession) Endpoint() string {
	return s.endpoint
}

func (s stubSession) Type() session.CredentialsType {
	return session.CredentialsTypeAPIToken
}
