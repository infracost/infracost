package schema

type Provider interface {
	LoadResources(map[string]*UsageData) (*Project, error)
}
