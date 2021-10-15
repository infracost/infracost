package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestResourceDataEmpty(t *testing.T) {
	r := NewResourceData("somettype", "someprovider", "some.address", map[string]string{},
		gjson.Result{
			Type: gjson.JSON,
			Raw: `{	
					"someresource": {
						"number": 0,
						"string": "string",
						"empty_string": "",
						"null": nil,
						"nested": {
							"number": 0,
							"string": "string",
							"empty_string": "",
							"null": nil,
						}
					}
				}`,
		})

	tests := []struct {
		key string
		want bool
	}{
		{ key: "someresource.missing", want: true },
		{ key: "someresource.number", want: false },
		{ key: "someresource.string", want: false },
		{ key: "someresource.empty_string", want: true },
		{ key: "someresource.null", want: true },

		// nested keys don't work in usage, so these are all expected to return false
		{ key: "someresource.nested.missing", want: true },
		{ key: "someresource.nested.number", want: false },
		{ key: "someresource.nested.string", want: false },
		{ key: "someresource.nested.empty_string", want: true },
		{ key: "someresource.nested.null", want: true },
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.key, func(t *testing.T) {
			assert.Equal(t, r.Empty(tt.key), tt.want)
		})
	}

}
