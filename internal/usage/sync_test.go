package usage

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/stretchr/testify/assert"
)

func TestReplaceResourceUsages(t *testing.T) {
	dest := &ResourceUsage{
		Name: "resource",
		Items: []*schema.UsageItem{
			{
				Key:          "attr_dest_default",
				ValueType:    schema.Int64,
				Value:        int64(10),
				DefaultValue: int64(1),
			},
			{
				Key:          "attr_dest_nil_val",
				ValueType:    schema.Int64,
				Value:        nil,
				DefaultValue: int64(2),
			},
			{
				Key:          "attr_dest_nil_default",
				ValueType:    schema.Int64,
				Value:        int64(30),
				DefaultValue: nil,
			},
			{
				Key:          "attr_src_not_exists",
				ValueType:    schema.Int64,
				Value:        int64(40),
				DefaultValue: int64(4),
			},
		},
	}

	src := &ResourceUsage{
		Name: "resource",
		Items: []*schema.UsageItem{
			{
				Key:          "attr_dest_default",
				ValueType:    schema.Int64,
				Value:        int64(10),
				DefaultValue: int64(1),
			},
			{
				Key:          "attr_dest_nil_val",
				ValueType:    schema.Int64,
				Value:        int64(20),
				DefaultValue: int64(2),
			},
			{
				Key:          "attr_dest_nil_default",
				ValueType:    schema.Int64,
				Value:        int64(30),
				DefaultValue: int64(3),
			},
			{
				Key:          "attr_dest_not_exists",
				ValueType:    schema.Int64,
				Value:        int64(50),
				DefaultValue: int64(5),
			},
		},
	}

	replaceResourceUsages(dest, src, ReplaceResourceUsagesOpts{})

	assert.Len(t, dest.Items, 5)
	assert.Equal(t, int64(1), dest.Items[0].DefaultValue.(int64))
	assert.Equal(t, int64(10), dest.Items[0].Value.(int64))
	assert.Equal(t, int64(2), dest.Items[1].DefaultValue.(int64))
	assert.Equal(t, int64(20), dest.Items[1].Value.(int64))
	assert.Equal(t, int64(3), dest.Items[2].DefaultValue.(int64))
	assert.Equal(t, int64(30), dest.Items[2].Value.(int64))
	assert.Equal(t, int64(4), dest.Items[3].DefaultValue.(int64))
	assert.Equal(t, int64(40), dest.Items[3].Value.(int64))
	assert.Equal(t, int64(5), dest.Items[4].DefaultValue.(int64))
	assert.Equal(t, int64(50), dest.Items[4].Value.(int64))
}

func TestReplaceResourceUsagesSubResources(t *testing.T) {
	dest := &ResourceUsage{
		Name: "resource",
		Items: []*schema.UsageItem{
			{
				Key:       "subresource_1",
				ValueType: schema.SubResourceUsage,
				Value: &ResourceUsage{
					Name: "subresource_1",
					Items: []*schema.UsageItem{
						{
							Key:          "attr_dest_default",
							ValueType:    schema.Int64,
							Value:        int64(10),
							DefaultValue: int64(1),
						},
						{
							Key:          "attr_dest_nil_val",
							ValueType:    schema.Int64,
							Value:        nil,
							DefaultValue: int64(2),
						},
						{
							Key:          "attr_dest_nil_default",
							ValueType:    schema.Int64,
							Value:        int64(30),
							DefaultValue: nil,
						},
						{
							Key:          "attr_src_not_exists",
							ValueType:    schema.Int64,
							Value:        int64(40),
							DefaultValue: int64(4),
						},
					},
				},
			},
			{
				Key:       "subresource_2",
				ValueType: schema.SubResourceUsage,
				Value: &ResourceUsage{
					Name: "subresource_2",
					Items: []*schema.UsageItem{
						{
							Key:          "attr_1",
							ValueType:    schema.Int64,
							Value:        int64(10),
							DefaultValue: int64(1),
						},
					},
				},
			},
		},
	}

	src := &ResourceUsage{
		Name: "resource",
		Items: []*schema.UsageItem{
			{
				Key:       "subresource_1",
				ValueType: schema.SubResourceUsage,
				Value: &ResourceUsage{
					Name: "subresource_1",
					Items: []*schema.UsageItem{
						{
							Key:          "attr_dest_default",
							ValueType:    schema.Int64,
							Value:        int64(10),
							DefaultValue: nil,
						},
						{
							Key:          "attr_dest_nil_val",
							ValueType:    schema.Int64,
							Value:        int64(20),
							DefaultValue: int64(2),
						},
						{
							Key:          "attr_dest_nil_default",
							ValueType:    schema.Int64,
							Value:        int64(30),
							DefaultValue: int64(3),
						},
						{
							Key:          "attr_dest_not_exists",
							ValueType:    schema.Int64,
							Value:        int64(50),
							DefaultValue: int64(5),
						},
					},
				},
			},
			{
				Key:       "subresource_3",
				ValueType: schema.SubResourceUsage,
				Value: &ResourceUsage{
					Name: "subresource_3",
					Items: []*schema.UsageItem{
						{
							Key:          "attr_1",
							ValueType:    schema.Int64,
							Value:        int64(10),
							DefaultValue: int64(1),
						},
					},
				},
			},
		},
	}

	replaceResourceUsages(dest, src, ReplaceResourceUsagesOpts{})

	assert.Len(t, dest.Items, 3)

	subResource1 := dest.Items[0].Value.(*ResourceUsage)
	assert.Len(t, subResource1.Items, 5)
	assert.Equal(t, int64(1), subResource1.Items[0].DefaultValue.(int64))
	assert.Equal(t, int64(10), subResource1.Items[0].Value.(int64))
	assert.Equal(t, int64(2), subResource1.Items[1].DefaultValue.(int64))
	assert.Equal(t, int64(20), subResource1.Items[1].Value.(int64))
	assert.Equal(t, int64(3), subResource1.Items[2].DefaultValue.(int64))
	assert.Equal(t, int64(30), subResource1.Items[2].Value.(int64))
	assert.Equal(t, int64(4), subResource1.Items[3].DefaultValue.(int64))
	assert.Equal(t, int64(40), subResource1.Items[3].Value.(int64))
	assert.Equal(t, int64(5), subResource1.Items[4].DefaultValue.(int64))
	assert.Equal(t, int64(50), subResource1.Items[4].Value.(int64))

	subResource2 := dest.Items[1].Value.(*ResourceUsage)
	assert.Len(t, subResource2.Items, 1)
	assert.Equal(t, int64(1), subResource2.Items[0].DefaultValue.(int64))
	assert.Equal(t, int64(10), subResource2.Items[0].Value.(int64))

	subResource3 := dest.Items[1].Value.(*ResourceUsage)
	assert.Len(t, subResource3.Items, 1)
	assert.Equal(t, int64(1), subResource3.Items[0].DefaultValue.(int64))
	assert.Equal(t, int64(10), subResource3.Items[0].Value.(int64))
}

func TestReplaceResourceUsagesTypes(t *testing.T) {
	newDest := func() *ResourceUsage {
		return &ResourceUsage{
			Name: "resource",
			Items: []*schema.UsageItem{
				{
					Key:          "attr_different_type",
					ValueType:    schema.Int64,
					Value:        int64(10),
					DefaultValue: int64(1),
				},
			},
		}
	}

	src := &ResourceUsage{
		Name: "resource",
		Items: []*schema.UsageItem{
			{
				Key:          "attr_different_type",
				ValueType:    schema.Float64,
				Value:        float64(10),
				DefaultValue: float64(1),
			},
		},
	}

	dest := newDest()
	replaceResourceUsages(dest, src, ReplaceResourceUsagesOpts{OverrideValueType: true})

	assert.Len(t, dest.Items, 1)
	assert.Equal(t, schema.Float64, dest.Items[0].ValueType)
	assert.Equal(t, float64(1), dest.Items[0].DefaultValue.(float64))
	assert.Equal(t, float64(10), dest.Items[0].Value.(float64))

	dest = newDest()
	replaceResourceUsages(dest, src, ReplaceResourceUsagesOpts{OverrideValueType: false})

	assert.Len(t, dest.Items, 1)
	assert.Equal(t, schema.Int64, dest.Items[0].ValueType)
	assert.Equal(t, float64(1), dest.Items[0].DefaultValue.(float64))
	assert.Equal(t, float64(10), dest.Items[0].Value.(float64))
}

func TestReplaceResourceUsagesDescription(t *testing.T) {
	dest := &ResourceUsage{
		Name: "resource",
		Items: []*schema.UsageItem{
			{
				Key:         "attr_dest_description",
				ValueType:   schema.Int64,
				Description: "Dest description 1",
			},
			{
				Key:         "attr_no_dest_description",
				ValueType:   schema.Int64,
				Description: "",
			},
			{
				Key:         "attr_no_src_description",
				ValueType:   schema.Int64,
				Description: "Dest description 3",
			},
		},
	}

	src := &ResourceUsage{
		Name: "resource",
		Items: []*schema.UsageItem{
			{
				Key:         "attr_dest_description",
				ValueType:   schema.Int64,
				Description: "Src description 1",
			},
			{
				Key:         "attr_no_dest_description",
				ValueType:   schema.Int64,
				Description: "Src description 2",
			},
			{
				Key:         "attr_no_src_description",
				ValueType:   schema.Int64,
				Description: "",
			},
		},
	}

	replaceResourceUsages(dest, src, ReplaceResourceUsagesOpts{})

	assert.Len(t, dest.Items, 3)
	assert.Equal(t, "Src description 1", dest.Items[0].Description)
	assert.Equal(t, "Src description 2", dest.Items[1].Description)
	assert.Equal(t, "Dest description 3", dest.Items[2].Description)
}

func TestReplaceResourceUsageWithUsageData(t *testing.T) {
	dest := &ResourceUsage{
		Name: "resource",
		Items: []*schema.UsageItem{
			{
				Key:       "attr_dest_default",
				ValueType: schema.Int64,
				Value:     int64(10),
			},
			{
				Key:       "attr_dest_nil_val",
				ValueType: schema.Int64,
				Value:     nil,
			},
			{
				Key:       "attr_src_not_exists",
				ValueType: schema.Int64,
				Value:     int64(30),
			},
		},
	}

	usageData := schema.NewUsageData(
		"resource",
		schema.ParseAttributes(map[string]any{
			"attr_dest_default":   int64(100),
			"attr_dest_nil_val":   int64(200),
			"attr_dest_not_exist": int64(400), // This should be skipped
		}),
	)

	mergeResourceUsageWithUsageData(dest, usageData)

	assert.Len(t, dest.Items, 3)
	assert.Equal(t, int64(100), dest.Items[0].Value.(int64))
	assert.Equal(t, int64(200), dest.Items[1].Value.(int64))
	assert.Equal(t, int64(30), dest.Items[2].Value.(int64))
}

func TestReplaceResourceUsageWithUsageDataDeep(t *testing.T) {
	dest := &ResourceUsage{
		Name: "resource",
		Items: []*schema.UsageItem{
			{
				Key:       "subresource_1",
				ValueType: schema.SubResourceUsage,
				Value: &ResourceUsage{
					Name: "subresource_1",
					Items: []*schema.UsageItem{
						{
							Key:       "attr_dest_default",
							ValueType: schema.Int64,
							Value:     int64(10),
						},
						{
							Key:       "attr_dest_nil_val",
							ValueType: schema.Int64,
							Value:     nil,
						},
						{
							Key:       "attr_src_not_exists",
							ValueType: schema.Int64,
							Value:     int64(30),
						},
					},
				},
			},
			{
				Key:       "subresource_2",
				ValueType: schema.SubResourceUsage,
				Value: &ResourceUsage{
					Name: "subresource_2",
					Items: []*schema.UsageItem{
						{
							Key:          "attr_1",
							ValueType:    schema.Int64,
							Value:        int64(10),
							DefaultValue: int64(1),
						},
					},
				},
			},
		},
	}

	usageData := schema.NewUsageData(
		"resource",
		schema.ParseAttributes(map[string]any{
			"subresource_1": map[string]any{
				"attr_dest_default":   int64(100),
				"attr_dest_nil_val":   int64(200),
				"attr_dest_not_exist": int64(400), // This should be skipped
			},
			"subresource_3": map[string]any{ // This should be skipped
				"attr_1": int64(100),
			},
		}),
	)

	mergeResourceUsageWithUsageData(dest, usageData)

	assert.Len(t, dest.Items, 2)

	subResource1 := dest.Items[0].Value.(*ResourceUsage)
	assert.Len(t, subResource1.Items, 3)
	assert.Equal(t, int64(100), subResource1.Items[0].Value.(int64))
	assert.Equal(t, int64(200), subResource1.Items[1].Value.(int64))
	assert.Equal(t, int64(30), subResource1.Items[2].Value.(int64))

	subResource2 := dest.Items[1].Value.(*ResourceUsage)
	assert.Len(t, subResource2.Items, 1)
	assert.Equal(t, int64(10), subResource2.Items[0].Value.(int64))
}
