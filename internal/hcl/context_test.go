package hcl

import (
	"io"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

func Test_ContextVariables(t *testing.T) {
	underlying := &hcl.EvalContext{}
	ctx := NewContext(underlying, nil, newTestLogger())

	val, err := gocty.ToCtyValue("hello", cty.String)
	if err != nil {
		t.Fatal(err)
	}

	ctx.Set(val, "my", "value")
	value := underlying.Variables["my"].AsValueMap()["value"]
	assert.Equal(t, "hello", value.AsString())

}

func Test_ContextVariablesPreservation(t *testing.T) {

	underlying := &hcl.EvalContext{}
	underlying.Variables = make(map[string]cty.Value)
	underlying.Variables["x"], _ = gocty.ToCtyValue("does it work?", cty.String)
	str, _ := gocty.ToCtyValue("something", cty.String)
	underlying.Variables["my"] = cty.ObjectVal(map[string]cty.Value{
		"other": str,
		"obj": cty.ObjectVal(map[string]cty.Value{
			"another": str,
		}),
	})
	ctx := NewContext(underlying, nil, newTestLogger())

	val, err := gocty.ToCtyValue("hello", cty.String)
	if err != nil {
		t.Fatal(err)
	}

	ctx.Set(val, "my", "value")
	assert.Equal(t, "hello", underlying.Variables["my"].AsValueMap()["value"].AsString())
	assert.Equal(t, "something", underlying.Variables["my"].AsValueMap()["other"].AsString())
	assert.Equal(t, "something", underlying.Variables["my"].AsValueMap()["obj"].AsValueMap()["another"].AsString())
	assert.Equal(t, "does it work?", underlying.Variables["x"].AsString())

}

func Test_ContextVariablesPreservationByDot(t *testing.T) {

	underlying := &hcl.EvalContext{}
	underlying.Variables = make(map[string]cty.Value)
	underlying.Variables["x"], _ = gocty.ToCtyValue("does it work?", cty.String)
	str, _ := gocty.ToCtyValue("something", cty.String)
	underlying.Variables["my"] = cty.ObjectVal(map[string]cty.Value{
		"other": str,
		"obj": cty.ObjectVal(map[string]cty.Value{
			"another": str,
		}),
	})
	ctx := NewContext(underlying, nil, newTestLogger())

	val, err := gocty.ToCtyValue("hello", cty.String)
	if err != nil {
		t.Fatal(err)
	}

	ctx.SetByDot(val, "my.something.value")
	assert.Equal(t, "hello", underlying.Variables["my"].AsValueMap()["something"].AsValueMap()["value"].AsString())
	assert.Equal(t, "something", underlying.Variables["my"].AsValueMap()["other"].AsString())
	assert.Equal(t, "something", underlying.Variables["my"].AsValueMap()["obj"].AsValueMap()["another"].AsString())
	assert.Equal(t, "does it work?", underlying.Variables["x"].AsString())
}

func Test_ContextSetThenImmediateGet(t *testing.T) {

	underlying := &hcl.EvalContext{}

	entry := newTestLogger()
	ctx := NewContext(underlying, nil, entry)

	ctx.Set(cty.ObjectVal(map[string]cty.Value{
		"mod_result": cty.StringVal("ok"),
	}), "module", "modulename")

	val := ctx.Get("module", "modulename", "mod_result")
	assert.Equal(t, "ok", val.AsString())
}

func newTestLogger() zerolog.Logger {
	return zerolog.New(io.Discard)
}

func Test_ContextSetThenImmediateGetWithChild(t *testing.T) {

	underlying := &hcl.EvalContext{}

	ctx := NewContext(underlying, nil, newTestLogger())

	childCtx := ctx.NewChild()

	childCtx.Root().Set(cty.ObjectVal(map[string]cty.Value{
		"mod_result": cty.StringVal("ok"),
	}), "module", "modulename")

	val := ctx.Get("module", "modulename", "mod_result")
	assert.Equal(t, "ok", val.AsString())
}
