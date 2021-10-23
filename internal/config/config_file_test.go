package config

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYamlError_Error(t *testing.T) {
	type fields struct {
		raw    error
		base   string
		errors []error
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "should return raw error message",
			fields: fields{
				raw: errors.New("raw message"),
			},
			want: "raw message",
		},
		{
			name: "should return formatted with base and sub errors",
			fields: fields{
				base: "top message",
				errors: []error{
					errors.New("child 1"),
					errors.New("child 2"),
				},
			},
			want: `top message:
	child 1
	child 2`,
		},
		{
			name: "should return with recursive levels",
			fields: fields{
				base: "top message",
				errors: []error{
					&YamlError{
						base: "child message",
						errors: []error{
							errors.New("child child error 1"),
							&YamlError{
								base: "infant message",
								errors: []error{
									errors.New("infant error 1"),
								},
							},
						},
					},
				},
			},
			want: `top message:
	child message:
		child child error 1
		infant message:
			infant error 1`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			y := &YamlError{
				raw:    tt.fields.raw,
				base:   tt.fields.base,
				errors: tt.fields.errors,
			}
			assert.Equal(t, tt.want, y.Error())
		})
	}
}
