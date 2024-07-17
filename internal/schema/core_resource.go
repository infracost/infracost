package schema

import (
	"github.com/tidwall/gjson"
)

type CoreResourceFunc func(*ResourceData) CoreResource

// CoreResource is the new/preferred way to represent provider-agnostic resources that
// support advanced features such as Infracost Cloud usage estimates and actual costs.
type CoreResource interface {
	CoreType() string
	UsageSchema() []*UsageItem
	PopulateUsage(u *UsageData)
	BuildResource() *Resource
}

// BlankCoreResource is a helper struct for NoPrice and Skipped resources that are evaluated as
// part of the Policy API. This implements the CoreResource interface and returns a skipped resource
// that doesn't affect the CLI output.
type BlankCoreResource struct {
	Name string
	Type string
}

func (b BlankCoreResource) CoreType() string           { return b.Type }
func (b BlankCoreResource) UsageSchema() []*UsageItem  { return nil }
func (b BlankCoreResource) PopulateUsage(u *UsageData) {}
func (b BlankCoreResource) BuildResource() *Resource {
	return &Resource{
		Name:      b.Name,
		IsSkipped: true,
		NoPrice:   true,
	}
}

type UsageParam struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// CoreResourceWithUsageParams is a CoreResource that sends additional
// parameters (e.g. Lambda memory size) to the Usage API when estimating usage.
type CoreResourceWithUsageParams interface {
	CoreResource
	UsageEstimationParams() []UsageParam
}

// PartialResource is used to collect all information required to construct a
// resource that is generated by provider parser and pass it back up to
// top level functions that can supply additional provider-agnostic information
// (such as Infracost Cloud usage estimates) before the resource is built.
type PartialResource struct {
	Type                        string
	Address                     string
	Tags                        *map[string]string
	DefaultTags                 *map[string]string
	ProviderSupportsDefaultTags bool
	ProviderLink                string
	TagPropagation              *TagPropagation
	UsageData                   *UsageData
	Metadata                    map[string]gjson.Result

	// CoreResource is the new/preferred struct for providing an intermediate-object
	// that contains all provider-derived information, but has not yet been built into
	// a Resource.
	CoreResource CoreResource

	// Resource field is provided for backward compatibility with provider resource builders
	// that have not yet been converted to build CoreResource's
	Resource *Resource

	// CloudResourceIDs are collected during parsing in case they need to be uploaded to the
	// Cloud Usage API to be used in the usage estimate calculations.
	CloudResourceIDs []string
}

func NewPartialResource(d *ResourceData, r *Resource, cr CoreResource, cloudResourceIds []string) *PartialResource {
	return &PartialResource{
		Type:                        d.Type,
		Address:                     d.Address,
		Tags:                        d.Tags,
		DefaultTags:                 d.DefaultTags,
		ProviderSupportsDefaultTags: d.ProviderSupportsDefaultTags,
		ProviderLink:                d.ProviderLink,
		TagPropagation:              d.TagPropagation,
		UsageData:                   d.UsageData,
		Metadata:                    d.Metadata,
		CoreResource:                cr,
		Resource:                    r,
		CloudResourceIDs:            cloudResourceIds,
	}
}

// BuildResource create a new Resource from the CoreResource, or (for backward compatibility) returns
// a previously built Resource
func BuildResource(partial *PartialResource, fetchedUsage *UsageData) *Resource {
	var res *Resource
	if partial.CoreResource != nil {
		u := partial.UsageData
		u = u.Merge(fetchedUsage)

		partial.CoreResource.PopulateUsage(u)
		res = partial.CoreResource.BuildResource()
	} else {
		res = partial.Resource
	}

	if res == nil {
		return &Resource{
			Name:         partial.Address,
			ResourceType: partial.Type,
			IsSkipped:    true,
			SkipMessage:  "This resource is not currently supported",
		}
	}

	res.ResourceType = partial.Type
	res.Tags = partial.Tags
	res.DefaultTags = partial.DefaultTags
	res.ProviderSupportsDefaultTags = partial.ProviderSupportsDefaultTags
	res.ProviderLink = partial.ProviderLink
	res.TagPropagation = partial.TagPropagation
	res.Metadata = partial.Metadata
	return res
}

func BuildResources(projects []*Project, projectPtrToUsageMap map[*Project]UsageMap) {
	for _, project := range projects {
		usageMap := projectPtrToUsageMap[project]

		project.BuildResources(usageMap)
	}
}
