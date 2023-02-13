package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestParseAttributes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args interface{}
		want map[string]gjson.Result
	}{
		{
			name: "nested attributes are returned as nested gjson",
			args: map[interface{}]interface{}{
				"standard": map[interface{}]interface{}{
					"storage_gb":              1000,
					"monthly_tier_1_requests": 200,
				},
				"intelligent_tiering": map[string]interface{}{
					"monitored_objects":       500,
					"monthly_tier_2_requests": "a test string",
					"some_more_nesting": map[interface{}]interface{}{
						"so_nested": true,
					},
				},
			},
			want: map[string]gjson.Result{
				"standard": {
					Type: gjson.JSON,
					Raw:  `{"storage_gb": 1000, "monthly_tier_1_requests": 200}`,
				},
				"intelligent_tiering": {
					Type: gjson.JSON,
					Raw:  `{"monitored_objects": 500, "monthly_tier_2_requests": "a test string", "some_more_nesting": { "so_nested": true }}`,
				},
			},
		},
		{
			name: "single attributes are returned as gjson values",
			args: map[interface{}]interface{}{
				"num":    100,
				"bool":   true,
				"null":   nil,
				"string": "some string",
			},
			want: map[string]gjson.Result{
				"num": {
					Type:  gjson.Number,
					Raw:   "100",
					Str:   "",
					Num:   100,
					Index: 0,
				},
				"bool": {
					Type:  gjson.True,
					Raw:   "true",
					Str:   "",
					Num:   0,
					Index: 0,
				},
				"null": {
					Type:  gjson.Null,
					Raw:   "null",
					Str:   "",
					Num:   0,
					Index: 0,
				},
				"string": {
					Type:  gjson.String,
					Raw:   `"some string"`,
					Str:   "some string",
					Num:   0,
					Index: 0,
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := ParseAttributes(tt.args)
			assert.Len(t, got, len(tt.want))

			// range over the keys so we can do deep matching if it's json
			// we can't just do a basic compare as the map is unsorted and
			// matching will fail.
			for k, v := range tt.want {
				g, ok := got[k]
				require.True(t, ok)

				if v.Type == gjson.JSON {
					assert.JSONEq(t, v.Raw, g.Raw)

					// unset the json on both and now we can compare rest of
					// the fields freely
					v.Raw = ""
					g.Raw = ""
					assert.Equal(t, v, g)
					continue
				}

				assert.Equal(t, v, g)
			}
		})
	}
}

func TestUsageMap_Get(t *testing.T) {
	tests := []struct {
		address string
		usage   UsageMap
		want    *UsageData
	}{
		{
			address: "aws_lambda_function.hello_world",
			usage: NewUsageMapFromInterface(map[string]interface{}{"aws_lambda_function.hello_world": map[string]interface{}{
				"test": "this",
			}, "aws_lambda_function.hello_world2": ""}),
			want: &UsageData{
				Attributes: map[string]gjson.Result{
					"test": gjson.Parse(`"this"`),
				},
				Address: "aws_lambda_function.hello_world",
			},
		},
		{
			address: `aws_lambda_function.hello_world["foo"]`,
			usage: NewUsageMapFromInterface(map[string]interface{}{`aws_lambda_function.hello_world["foo"]`: map[string]interface{}{
				"test": "this",
			}, `aws_lambda_function.hello_world["bar"]`: ""}),
			want: &UsageData{
				Attributes: map[string]gjson.Result{
					"test": gjson.Parse(`"this"`),
				},
				Address: `aws_lambda_function.hello_world["foo"]`,
			},
		},
		{
			address: `aws_lambda_function.hello_world["foo"]`,
			usage: NewUsageMapFromInterface(map[string]interface{}{`aws_lambda_function.hello_world[*]`: map[string]interface{}{
				"test": "this",
			}, `aws_lambda_function.bar`: ""}),
			want: &UsageData{
				Attributes: map[string]gjson.Result{
					"test": gjson.Parse(`"this"`),
				},
				Address: `aws_lambda_function.hello_world[*]`,
			},
		},
		{
			address: `module.some_mod["foo"].aws_lambda_function.hello_world["bar"]`,
			usage: NewUsageMapFromInterface(map[string]interface{}{`module.some_mod["foo"].aws_lambda_function.hello_world["bar"]`: map[string]interface{}{
				"test": "this",
			}, `aws_lambda_function.hello_world`: ""}),
			want: &UsageData{
				Attributes: map[string]gjson.Result{
					"test": gjson.Parse(`"this"`),
				},
				Address: `module.some_mod["foo"].aws_lambda_function.hello_world["bar"]`,
			},
		},
		{
			address: `module.some_mod["foo"].aws_lambda_function.hello_world["bar"]`,
			usage: NewUsageMapFromInterface(map[string]interface{}{`module.some_mod["foo"].aws_lambda_function.hello_world[*]`: map[string]interface{}{
				"test": "this",
			}, `aws_lambda_function.hello_world`: ""}),
			want: &UsageData{
				Attributes: map[string]gjson.Result{
					"test": gjson.Parse(`"this"`),
				},
				Address: `module.some_mod["foo"].aws_lambda_function.hello_world[*]`,
			},
		},
		{
			address: `module.some_mod["foo"].module.some_bar["baz"].aws_lambda_function.hello_world["bar"]`,
			usage: NewUsageMapFromInterface(map[string]interface{}{`module.some_mod["foo"].module.some_bar[*].aws_lambda_function.hello_world[*]`: map[string]interface{}{
				"test": "this",
			}, `aws_lambda_function.hello_world`: ""}),
			want: &UsageData{
				Attributes: map[string]gjson.Result{
					"test": gjson.Parse(`"this"`),
				},
				Address: `module.some_mod["foo"].module.some_bar[*].aws_lambda_function.hello_world[*]`,
			},
		},
		{
			address: `module.mod["test2"].aws_lambda_function.test["foo"]`,
			usage: NewUsageMapFromInterface(map[string]interface{}{`module.mod["test2"].aws_lambda_function.test["foo"]`: map[string]interface{}{
				"test": "this",
			}, `module.mod[*].aws_lambda_function.test[*]`: ""}),
			want: &UsageData{
				Attributes: map[string]gjson.Result{
					"test": gjson.Parse(`"this"`),
				},
				Address: `module.mod["test2"].aws_lambda_function.test["foo"]`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.address, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.usage.Get(tt.address), "Get(%v)", tt.address)
		})
	}
}
