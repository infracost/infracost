package funcs

import "github.com/zclconf/go-cty/cty"

func refineNonNull(b *cty.RefinementBuilder) *cty.RefinementBuilder {
	return b.NotNull()
}
