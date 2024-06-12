package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestResourceDataEmpty(t *testing.T) {
	r := NewResourceData("somettype", "someprovider", "some.address", nil, gjson.Result{
		Type: gjson.JSON,
		Raw: `{	
					"someresource": {
						"number": 0,
						"string": "string",
						"object": { "a": 1 },
						"array": [ 1, 2, 3],
						"empty_string": "",
						"empty_object": {  
                        },
						"empty_array": [  
                        ],	
						"null": null,
						"nested": {
							"number": 0,
							"string": "string",
							"object": { "a": 1 },
							"array": [ 1, 2, 3],
							"empty_string": "",
							"empty_object": {
							},
							"empty_array": [
							],	
							"null": null,
						}
					}
				}`,
	})

	tests := []struct {
		key  string
		want bool
	}{
		{key: "someresource.missing", want: true},
		{key: "someresource.number", want: false},
		{key: "someresource.string", want: false},
		{key: "someresource.object", want: false},
		{key: "someresource.array", want: false},
		{key: "someresource.empty_string", want: true},
		{key: "someresource.empty_object", want: true},
		{key: "someresource.empty_array", want: true},
		{key: "someresource.null", want: true},

		{key: "someresource.nested.missing", want: true},
		{key: "someresource.nested.number", want: false},
		{key: "someresource.nested.string", want: false},
		{key: "someresource.nested.object", want: false},
		{key: "someresource.nested.array", want: false},
		{key: "someresource.nested.empty_string", want: true},
		{key: "someresource.nested.empty_object", want: true},
		{key: "someresource.nested.empty_array", want: true},
		{key: "someresource.nested.null", want: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.key, func(t *testing.T) {
			assert.Equal(t, tt.want, r.IsEmpty(tt.key))
		})
	}

}
