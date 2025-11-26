package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/schema"
)

type resourceWithFloat struct {
	MyFloat *float64 `infracost_usage:"my_float"`
}

type resourceWithString struct {
	MyString *string `infracost_usage:"my_string"`
}

type resourceWithInt struct {
	MyInt *int64 `infracost_usage:"my_int"`
}

type resourceWithSubUsage struct {
	MySubUsage *subUsageResource `infracost_usage:"my_sub_usage"`
}

type subUsageResource struct {
	MyInt    *int64  `infracost_usage:"my_int"`
	MyString *string `infracost_usage:"my_string"`
}

type args struct {
	args any
	u    *schema.UsageData
}

func TestPopulateArgsWithUsage(t *testing.T) {
	tests := []struct {
		name string
		args args
		want any
	}{
		{
			name: "parses float usage",
			args: args{
				args: &resourceWithFloat{},
				u: &schema.UsageData{
					Attributes: map[string]gjson.Result{
						"my_float": {
							Type: gjson.Number,
							Raw:  "1.4",
							Num:  1.4,
						},
					},
				},
			},
			want: &resourceWithFloat{
				MyFloat: newFloat(1.4),
			},
		},
		{
			name: "parses int usage",
			args: args{
				args: &resourceWithInt{},
				u: &schema.UsageData{
					Attributes: map[string]gjson.Result{
						"my_int": {
							Type: gjson.Number,
							Raw:  "2",
							Num:  2,
						},
					},
				},
			},
			want: &resourceWithInt{
				MyInt: newInt(2),
			},
		},
		{
			name: "parses string usage",
			args: args{
				args: &resourceWithString{},
				u: &schema.UsageData{
					Attributes: map[string]gjson.Result{
						"my_string": {
							Type: gjson.String,
							Raw:  "mystring",
							Str:  "mystring",
						},
					},
				},
			},
			want: &resourceWithString{
				MyString: newString("mystring"),
			},
		},
		{
			name: "parses sub resources usage",
			args: args{
				args: &resourceWithSubUsage{},
				u: &schema.UsageData{
					Attributes: map[string]gjson.Result{
						"my_sub_usage": {
							Type: gjson.JSON,
							Raw:  `{"my_int": 3, "my_string": "test"}`,
						},
					},
				},
			},
			want: &resourceWithSubUsage{
				MySubUsage: &subUsageResource{
					MyInt:    newInt(3),
					MyString: newString("test"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tt.args.args

			PopulateArgsWithUsage(v, tt.args.u)

			assert.Equal(t, tt.want, v)
		})
	}
}

func newString(s string) *string {
	return &s
}

func newInt(i int64) *int64 {
	return &i
}

func newFloat(f float64) *float64 {
	return &f
}
