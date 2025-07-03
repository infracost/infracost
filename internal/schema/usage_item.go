package schema

type UsageVariableType int

const (
	Int64 UsageVariableType = iota
	String
	Float64
	StringArray
	SubResourceUsage
	KeyValueMap
	Bool
)

type UsageItem struct {
	Key          string
	DefaultValue interface{}
	Value        interface{}
	ValueType    UsageVariableType
	Description  string
}
