package hcl

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/assert"
)

func TestBlock_LocalName(t *testing.T) {
	tests := []struct {
		name  string
		block *Block
		want  string
	}{
		{
			name: "resource Block with empty labels will return empty local name",
			block: &Block{
				hclBlock: &hcl.Block{
					Type:   "resource",
					Labels: nil,
				},
				logger: newDiscardLogger(),
			},
			want: "",
		},
		{
			name: "resource Block with valid labels will return reference without resource type",
			block: &Block{
				hclBlock: &hcl.Block{
					Type:   "resource",
					Labels: []string{"my-resource", "my-name"},
				},
				logger: newDiscardLogger(),
			},
			want: "my-resource.my-name",
		},
		{
			name: "data Block with valid labels will return reference with Block type",
			block: &Block{
				hclBlock: &hcl.Block{
					Type:   "data",
					Labels: []string{"my-block", "my-name"},
				},
				logger: newDiscardLogger(),
			},
			want: "data.my-block.my-name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.block.LocalName(), "LocalName()")
		})
	}
}
