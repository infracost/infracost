package hcl

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestParseVariable(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    cty.Value
		wantErr bool
	}{
		{
			name:    "string input",
			input:   "hello",
			want:    cty.StringVal("hello"),
			wantErr: false,
		},
		{
			name:    "integer input",
			input:   42,
			want:    cty.NumberIntVal(42),
			wantErr: false,
		},
		{
			name:    "float input",
			input:   3.14,
			want:    cty.NumberFloatVal(3.14),
			wantErr: false,
		},
		{
			name:    "boolean input",
			input:   true,
			want:    cty.BoolVal(true),
			wantErr: false,
		},
		{
			name:    "map with string keys input",
			input:   map[string]interface{}{"key": "value"},
			want:    cty.ObjectVal(map[string]cty.Value{"key": cty.StringVal("value")}),
			wantErr: false,
		},
		{
			name:    "map with integer keys input",
			input:   map[interface{}]interface{}{1: "one", 2: "two"},
			want:    cty.ObjectVal(map[string]cty.Value{"1": cty.StringVal("one"), "2": cty.StringVal("two")}),
			wantErr: false,
		},
		{
			name:    "nested map",
			input:   map[string]interface{}{"outer": map[string]interface{}{"inner": "value"}},
			want:    cty.ObjectVal(map[string]cty.Value{"outer": cty.ObjectVal(map[string]cty.Value{"inner": cty.StringVal("value")})}),
			wantErr: false,
		},
		{
			name:    "list input",
			input:   []interface{}{"one", "two"},
			want:    cty.TupleVal([]cty.Value{cty.StringVal("one"), cty.StringVal("two")}),
			wantErr: false,
		},
		{
			name:    "nested list",
			input:   []interface{}{[]interface{}{"one", "two"}, "three"},
			want:    cty.TupleVal([]cty.Value{cty.TupleVal([]cty.Value{cty.StringVal("one"), cty.StringVal("two")}), cty.StringVal("three")}),
			wantErr: false,
		},
		{
			name:    "HCL expression - simple map and list",
			input:   `{"Hello": "world", "Foo": ["bar", "baz"]}`,
			want:    cty.ObjectVal(map[string]cty.Value{"Hello": cty.StringVal("world"), "Foo": cty.TupleVal([]cty.Value{cty.StringVal("bar"), cty.StringVal("baz")})}),
			wantErr: false,
		},
		{
			name:    "HCL expression - map with complex keys and list",
			input:   `{"hello": "world", "foo": ["bar", "baz"]}`,
			want:    cty.ObjectVal(map[string]cty.Value{"hello": cty.StringVal("world"), "foo": cty.TupleVal([]cty.Value{cty.StringVal("bar"), cty.StringVal("baz")})}),
			wantErr: false,
		},
		{
			name:    "invalid input",
			input:   func() {},
			want:    cty.DynamicVal,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVariable(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVariable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !got.RawEquals(tt.want) {
				t.Errorf("ParseVariable() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertToStringKeyMap(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  interface{}
	}{
		{
			name:  "map with interface{} keys",
			input: map[interface{}]interface{}{1: "one", "two": 2},
			want:  map[string]interface{}{"1": "one", "two": 2},
		},
		{
			name:  "nested map with interface{} keys",
			input: map[interface{}]interface{}{"outer": map[interface{}]interface{}{1: "one"}},
			want:  map[string]interface{}{"outer": map[string]interface{}{"1": "one"}},
		},
		{
			name:  "slice of interface{}",
			input: []interface{}{map[interface{}]interface{}{1: "one"}},
			want:  []interface{}{map[string]interface{}{"1": "one"}},
		},
		{
			name:  "non-map and non-slice input",
			input: "string",
			want:  "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertToStringKeyMap(tt.input)
			if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", tt.want) {
				t.Errorf("convertToStringKeyMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
