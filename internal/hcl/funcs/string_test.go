package funcs

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestEndsWith(t *testing.T) {
	tests := []struct {
		String cty.Value
		Suffix cty.Value
		Want   cty.Value
		Err    bool
	}{
		{ // Prefix matches
			cty.StringVal("hello"),
			cty.StringVal("llo"),
			cty.BoolVal(true),
			false,
		},
		{ // Prefix doesn't match
			cty.StringVal("helloz"),
			cty.StringVal("llo"),
			cty.BoolVal(false),
			false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("endswith(%#v, %#v)", test.String, test.Suffix), func(t *testing.T) {
			got, err := EndsWith(test.String, test.Suffix)

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

func TestReplace(t *testing.T) {
	tests := []struct {
		String  cty.Value
		Substr  cty.Value
		Replace cty.Value
		Want    cty.Value
		Err     bool
	}{
		{ // Regular search and replace
			cty.StringVal("hello"),
			cty.StringVal("hel"),
			cty.StringVal("bel"),
			cty.StringVal("bello"),
			false,
		},
		{ // Search string doesn't match
			cty.StringVal("hello"),
			cty.StringVal("nope"),
			cty.StringVal("bel"),
			cty.StringVal("hello"),
			false,
		},
		{ // Regular expression
			cty.StringVal("hello"),
			cty.StringVal("/l/"),
			cty.StringVal("L"),
			cty.StringVal("heLLo"),
			false,
		},
		{
			cty.StringVal("helo"),
			cty.StringVal("/(l)/"),
			cty.StringVal("$1$1"),
			cty.StringVal("hello"),
			false,
		},
		{ // Bad regexp
			cty.StringVal("hello"),
			cty.StringVal("/(l/"),
			cty.StringVal("$1$1"),
			cty.UnknownVal(cty.String),
			true,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("replace(%#v, %#v, %#v)", test.String, test.Substr, test.Replace), func(t *testing.T) {
			got, err := Replace(test.String, test.Substr, test.Replace)

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

func TestStartsWith(t *testing.T) {
	tests := []struct {
		String cty.Value
		Prefix cty.Value
		Want   cty.Value
		Err    bool
	}{
		{ // Prefix matches
			cty.StringVal("hello"),
			cty.StringVal("hel"),
			cty.BoolVal(true),
			false,
		},
		{ // Prefix doesn't match
			cty.StringVal("zhello"),
			cty.StringVal("hel"),
			cty.BoolVal(false),
			false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("startswith(%#v, %#v)", test.String, test.Prefix), func(t *testing.T) {
			got, err := StartsWith(test.String, test.Prefix)

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

func TestStrContains(t *testing.T) {
	tests := []struct {
		String cty.Value
		Substr cty.Value
		Want   cty.Value
		Err    bool
	}{
		{ // Prefix matches
			cty.StringVal("hello"),
			cty.StringVal("hel"),
			cty.BoolVal(true),
			false,
		},
		{ // Sub-string matches doesn't match
			cty.StringVal("zhello"),
			cty.StringVal("hel"),
			cty.BoolVal(true),
			false,
		},
		{ // Suffix matches doesn't match
			cty.StringVal("hello"),
			cty.StringVal("llo"),
			cty.BoolVal(true),
			false,
		},
		{ // No match
			cty.StringVal("hello"),
			cty.StringVal("ezl"),
			cty.BoolVal(false),
			false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("strcontains(%#v, %#v)", test.String, test.Substr), func(t *testing.T) {
			got, err := StrContains(test.String, test.Substr)

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
