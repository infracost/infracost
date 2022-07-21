package hcl

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		{
			name:  "cty null to int",
			value: cty.NullVal(cty.Number),
			want:  0,
		},
		{
			name:  "cty nil to int",
			value: cty.NilVal,
			want:  0,
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
		{
			name:  "cty null to string",
			value: cty.NullVal(cty.String),
			want:  "",
		},
		{
			name:  "cty nil to string",
			value: cty.NilVal,
			want:  "",
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

func TestAttributeValueWithIncompleteContextAndConditionalShouldNotPanic(t *testing.T) {
	p := hclparse.NewParser()
	f, diags := p.ParseHCL([]byte(`
locals {
  original_tags    = "test"
  transformed_tags = local.original_tags
  id = var.enabled ? local.transformed_tags : "test3"
}
`), "test")

	require.False(t, diags.HasErrors(), fmt.Sprintf("diags has unexpected error %s from parsing input string", diags.Error()))

	c, _, diags := f.Body.PartialContent(terraformSchemaV012)
	require.False(t, diags.HasErrors(), "diags has unexpected error %s from parsing body content", diags.Error())

	var block *hcl.Block
	for _, b := range c.Blocks {
		if b.Type == "locals" {
			block = b
		}
	}

	require.NotNil(t, block, "could not find required test block")

	attrs, diags := block.Body.JustAttributes()
	require.False(t, diags.HasErrors(), "diags has unexpected error %s fetching attributes", diags.Error())

	buf := bytes.NewBuffer([]byte{})
	l := logrus.New()
	l.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	l.SetOutput(buf)
	l.SetLevel(logrus.DebugLevel)
	logger := logrus.NewEntry(l)

	l2 := logrus.New()
	l2.SetOutput(io.Discard)
	discard := logrus.NewEntry(l2)

	tag := Attribute{
		HCLAttr: attrs["transformed_tags"],
		Ctx: &Context{ctx: &hcl.EvalContext{
			Variables: map[string]cty.Value{},
		}},
		Logger: discard,
	}

	attr := Attribute{
		HCLAttr: attrs["id"],
		Ctx: &Context{
			ctx: &hcl.EvalContext{
				Variables: map[string]cty.Value{
					"local": cty.ObjectVal(map[string]cty.Value{
						"original_tags":    cty.StringVal("test"),
						"transformed_tags": tag.Value(),
					}),
					"var": cty.ObjectVal(map[string]cty.Value{
						"enabled": cty.BoolVal(true),
					}),
				},
			},
			logger: logger,
		},
		Verbose: false,
		Logger:  logger,
	}

	v := attr.Value()
	assert.Equal(t, cty.DynamicVal, v)

	b, err := io.ReadAll(buf)
	require.NoError(t, err)

	assert.NotContains(t, string(b), "invalid memory address")
}
