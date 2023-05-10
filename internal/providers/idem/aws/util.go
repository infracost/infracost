package aws

import (
	"github.com/awslabs/goformation/v4/cloudformation/tags"
)

func mapTags(cfTags []tags.Tag) map[string]string {
	mapped := make(map[string]string)
	for _, tag := range cfTags {
		mapped[tag.Key] = tag.Value
	}
	return mapped
}

func intPtr(i int64) *int64 {
	return &i
}

func strPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}
