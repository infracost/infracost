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
