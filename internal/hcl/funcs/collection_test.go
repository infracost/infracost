package funcs

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestCoalesce(t *testing.T) {
	tests := []struct {
		Values []cty.Value
		Want   cty.Value
		Err    bool
	}{
		{
			[]cty.Value{cty.StringVal("first"), cty.StringVal("second"), cty.StringVal("third")},
			cty.StringVal("first"),
			false,
		},
		{
			[]cty.Value{cty.StringVal(""), cty.StringVal("second"), cty.StringVal("third")},
			cty.StringVal("second"),
			false,
		},
		{
			[]cty.Value{cty.StringVal(""), cty.StringVal("")},
			cty.NilVal,
			true,
		},
		{
			[]cty.Value{cty.True},
			cty.True,
			false,
		},
		{
			[]cty.Value{cty.NullVal(cty.Bool), cty.True},
			cty.True,
			false,
		},
		{
			[]cty.Value{cty.NullVal(cty.Bool), cty.False},
			cty.False,
			false,
		},
		{
			[]cty.Value{cty.NullVal(cty.Bool), cty.False, cty.StringVal("hello")},
			cty.StringVal("false"),
			false,
		},
		{
			[]cty.Value{cty.True, cty.UnknownVal(cty.Bool)},
			cty.True,
			false,
		},
		{
			[]cty.Value{cty.UnknownVal(cty.Bool), cty.True},
			cty.UnknownVal(cty.Bool),
			false,
		},
		{
			[]cty.Value{cty.UnknownVal(cty.Bool), cty.StringVal("hello")},
			cty.UnknownVal(cty.String),
			false,
		},
		{
			[]cty.Value{cty.DynamicVal, cty.True},
			cty.UnknownVal(cty.Bool),
			false,
		},
		{
			[]cty.Value{cty.DynamicVal},
			cty.DynamicVal,
			false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Coalesce(%#v...)", test.Values), func(t *testing.T) {
			got, err := Coalesce(test.Values...)

			if test.Err {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}
