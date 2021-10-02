package schema_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/schema"
)

func TestUsageData_GetMap(t *testing.T) {
	t.Parallel()

	n2000 := gjson.Result{
		Type:  gjson.Number,
		Raw:   "2000",
		Str:   "",
		Num:   2000,
		Index: 0,
	}

	n1000 := gjson.Result{
		Type:  gjson.Number,
		Raw:   "1000",
		Str:   "",
		Num:   1000,
		Index: 0,
	}

	type fields struct {
		Attributes map[string]gjson.Result
	}

	type args struct {
		key string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]gjson.Result
	}{
		{
			name: "parses nested properties to map of gjson",
			fields: fields{
				Attributes: map[string]gjson.Result{
					"monthly_outbound_from_region_to_dx_connection_location.eu_west_1": n1000,
					"monthly_outbound_from_region_to_dx_connection_location.eu_east_2": n2000,
				},
			},
			args: args{
				key: "monthly_outbound_from_region_to_dx_connection_location",
			},
			want: map[string]gjson.Result{
				"eu_west_1": n1000,
				"eu_east_2": n2000,
			},
		},
		{
			name: "not found key returns nil",
			fields: fields{
				Attributes: map[string]gjson.Result{
					"monthly_outbound_from_region_to_dx_connection_location.eu_west_1": n1000,
					"monthly_outbound_from_region_to_dx_connection_location.eu_east_2": n2000,
				},
			},
			args: args{
				key: "not_found",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			u := &schema.UsageData{
				Attributes: tt.fields.Attributes,
			}

			assert.Equal(t, tt.want, u.GetMap(tt.args.key))
		})
	}
}
