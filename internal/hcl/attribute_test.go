package hcl

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"
)

func TestAttribute_AsInt(t *testing.T) {
	tests := []struct {
		name  string
		value cty.Value
		want  int64
	}{
		{
			name:  "cty number to int",
			value: cty.NumberIntVal(66),
			want:  66,
		},
		{
			name:  "cty string to int",
			value: cty.StringVal("66"),
			want:  66,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr := &Attribute{
				Ctx: &Context{
					ctx: &hcl.EvalContext{
						Variables: map[string]cty.Value{},
					},
					logger: newDiscardLogger(),
				},
				HCLAttr: &hcl.Attribute{
					Expr: hcl.StaticExpr(tt.value, hcl.Range{}),
				},
				Logger: newDiscardLogger(),
			}

			actual := attr.AsInt()
			assert.Equalf(t, tt.want, actual, "AsInt()")
		})
	}
}

func TestAttribute_AsString(t *testing.T) {
	tests := []struct {
		name  string
		value cty.Value
		want  string
	}{
		{
			name:  "cty string to string",
			value: cty.StringVal("test"),
			want:  "test",
		},
		{
			name:  "cty int to string",
			value: cty.NumberIntVal(1),
			want:  "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr := &Attribute{
				Ctx: &Context{
					ctx: &hcl.EvalContext{
						Variables: map[string]cty.Value{},
					},
					logger: newDiscardLogger(),
				},
				HCLAttr: &hcl.Attribute{
					Expr: hcl.StaticExpr(tt.value, hcl.Range{}),
				},
				Logger: newDiscardLogger(),
			}

			actual := attr.AsString()
			assert.Equalf(t, tt.want, actual, "AsString()")
		})
	}
}
