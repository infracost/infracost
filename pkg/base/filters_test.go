package base

import (
	"fmt"
	"reflect"
	"testing"
)

func TestMappedValue(t *testing.T) {
	var valueMapping *ValueMapping
	var result string

	valueMapping = &ValueMapping{"from", "to", nil}
	result = valueMapping.MappedValue("val")
	if (result != "val") {
		t.Error("got wrong mapped value")
	}

	valueMapping = &ValueMapping{"from", "to", func(fromVal interface{}) string { return fmt.Sprintf("%s.mapped", fromVal) }}
	result = valueMapping.MappedValue("val")
	if (result != "val.mapped") {
		t.Error("got wrong mapped value when a function is used")
	}
}

func TestMergeFilters(t *testing.T) {
	var filtersA = []Filter{
		{Key: "key1", Value: "val1"},
		{Key: "key2", Value: "val2"},
	}

	var filtersB = []Filter{
		{Key: "key3", Value: "val3"},
		{Key: "key1", Value: "val1-updated"},
	}

	result := MergeFilters(filtersA, filtersB)

	expected := []Filter{
		{Key: "key1", Value: "val1-updated"},
		{Key: "key2", Value: "val2"},
		{Key: "key3", Value: "val3"},
	}

	if (!reflect.DeepEqual(result, expected)) {
		t.Error("did not get the expected output", result)
	}
}

func TestMapFilters(t *testing.T) {
	valueMappings := []ValueMapping{
		{"fromKey1", "toKey1", nil},
		{"fromKey2", "toKey2", func(fromVal interface{}) string { return fmt.Sprintf("%s.mapped", fromVal) }},
	}

	values := map[string]interface{}{
		"fromKey1": 1,
		"fromKey2": "val2",
		"fromKey3": "val3",
	}

	result := MapFilters(valueMappings, values)

	expected := []Filter{
		{Key: "toKey1", Value: "1"},
		{Key: "toKey2", Value: "val2.mapped"},
	}

	if (!reflect.DeepEqual(result, expected)) {
		t.Error("did not get the expected output", result)
	}
}
