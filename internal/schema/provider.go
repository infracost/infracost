package schema

type Provider interface {
	Type() string
	DisplayType() string
	AddMetadata(*ProjectMetadata)
	LoadResources(UsageMap) ([]*Project, error)
}
