package hcl

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/spacelift-io/spacectl/client"
	"github.com/spacelift-io/spacectl/client/session"
	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"

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
	}{
		{
			name: "return variables from spacelift context",
			fields: fields{
				Metadata: vcs.Metadata{
					Remote: vcs.Remote{
						Name: "test/test",
					},
				},
			},
			args: args{
				options: RemoteVarLoaderOptions{
					EnvName: "dev",
				},
			},
			want: map[string]cty.Value{
				"test":  cty.StringVal("1"),
				"test2": cty.StringVal("foo"),
			},
			wantErr: assert.NoError,
			testServerFunc: func(t *testing.T) http.HandlerFunc {
				// test that the /graqphl endpoint has been called and the correct query has been sent
				return func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "POST", r.Method)
					s, err := io.ReadAll(r.Body)
					assert.NoError(t, err)
					assert.JSONEq(t, `{"query":"query($input:SearchInput!){searchStacks(input: $input){edges{node{id,name,config{id,value}}},pageInfo{endCursor,hasNextPage,hasPreviousPage}}}","variables":{"input":{"first":1,"after":null,"fullTextSearch":null,"predicates":[{"field":"repository","constraint":{"booleanEquals":null,"enumEquals":null,"stringMatches":["test/test"]}},{"field":"name","constraint":{"booleanEquals":null,"enumEquals":null,"stringMatches":["dev"]}}],"orderBy":null}}}`, string(s))
					assert.Equal(t, "Bearer test", r.Header.Get("Authorization"))

					_, err = w.Write([]byte(`{"data":{"searchStacks":{"edges":[{"node":{"id":"dev","name":"dev","config":[{"id":"test","value":"1"},{"id":"test2","value":"foo"}]}}],"pageInfo":{"endCursor":"MDFKNEtQMzVETjFYRDcxNTY1UkJCMzBITUQ=","hasNextPage":true,"hasPreviousPage":false}}}}`))
					assert.NoError(t, err)
				}
			},
		},
		{
			name: "returns empty map if no variables found",
			fields: fields{
				Metadata: vcs.Metadata{
					Remote: vcs.Remote{
						Name: "test/test",
					},
				},
			},
			args: args{
				options: RemoteVarLoaderOptions{
					EnvName: "dev",
				},
			},
			want:    nil,
			wantErr: assert.NoError,
			testServerFunc: func(t *testing.T) http.HandlerFunc {
				// test that the /graqphl endpoint has been called and the correct query has been sent
				return func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "POST", r.Method)
					s, err := io.ReadAll(r.Body)
					assert.NoError(t, err)
					assert.JSONEq(t, `{"query":"query($input:SearchInput!){searchStacks(input: $input){edges{node{id,name,config{id,value}}},pageInfo{endCursor,hasNextPage,hasPreviousPage}}}","variables":{"input":{"first":1,"after":null,"fullTextSearch":null,"predicates":[{"field":"repository","constraint":{"booleanEquals":null,"enumEquals":null,"stringMatches":["test/test"]}},{"field":"name","constraint":{"booleanEquals":null,"enumEquals":null,"stringMatches":["dev"]}}],"orderBy":null}}}`, string(s))
					assert.Equal(t, "Bearer test", r.Header.Get("Authorization"))

					_, err = w.Write([]byte(`{"data":{"searchStacks":{"edges":[{"node":{"id":"dev","name":"dev","config":[]}}],"pageInfo":{"endCursor":"MDFKNEtQMzVETjFYRDcxNTY1UkJCMzBITUQ=","hasNextPage":true,"hasPreviousPage":false}}}}`))
					assert.NoError(t, err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.testServerFunc(t))
			defer ts.Close()

			s := &SpaceliftRemoteVariableLoader{
				Client: client.New(http.DefaultClient, &stubSession{
					token:    "test",
					endpoint: ts.URL,
				}),
				Metadata: tt.fields.Metadata,
				cache:    &sync.Map{},
			}

			got, err := s.Load(tt.args.options)
			if !tt.wantErr(t, err, fmt.Sprintf("Load(%v)", tt.args.options)) {
				return
			}
			assert.Equalf(t, tt.want, got, "Load(%v)", tt.args.options)
		})
	}
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
