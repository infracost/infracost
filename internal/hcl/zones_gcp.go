package hcl

import "github.com/zclconf/go-cty/cty"

var gcpZones = map[string]cty.Value{
	"asia-east1": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("asia-east1-a"), cty.StringVal("asia-east1-b"), cty.StringVal("asia-east1-c")}),
	}),
	"asia-east2": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("asia-east2-c"), cty.StringVal("asia-east2-b"), cty.StringVal("asia-east2-a")}),
	}),
	"asia-northeast1": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("asia-northeast1-a"), cty.StringVal("asia-northeast1-b"), cty.StringVal("asia-northeast1-c")}),
	}),
	"asia-northeast2": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("asia-northeast2-b"), cty.StringVal("asia-northeast2-c"), cty.StringVal("asia-northeast2-a")}),
	}),
	"asia-northeast3": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("asia-northeast3-a"), cty.StringVal("asia-northeast3-c"), cty.StringVal("asia-northeast3-b")}),
	}),
	"asia-south1": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("asia-south1-b"), cty.StringVal("asia-south1-a"), cty.StringVal("asia-south1-c")}),
	}),
	"asia-south2": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("asia-south2-a"), cty.StringVal("asia-south2-c"), cty.StringVal("asia-south2-b")}),
	}),
	"asia-southeast1": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("asia-southeast1-a"), cty.StringVal("asia-southeast1-b"), cty.StringVal("asia-southeast1-c")}),
	}),
	"asia-southeast2": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("asia-southeast2-a"), cty.StringVal("asia-southeast2-c"), cty.StringVal("asia-southeast2-b")}),
	}),
	"australia-southeast1": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("australia-southeast1-c"), cty.StringVal("australia-southeast1-a"), cty.StringVal("australia-southeast1-b")}),
	}),
	"australia-southeast2": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("australia-southeast2-a"), cty.StringVal("australia-southeast2-c"), cty.StringVal("australia-southeast2-b")}),
	}),
	"europe-central2": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("europe-central2-b"), cty.StringVal("europe-central2-c"), cty.StringVal("europe-central2-a")}),
	}),
	"europe-north1": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("europe-north1-b"), cty.StringVal("europe-north1-c"), cty.StringVal("europe-north1-a")}),
	}),
	"europe-southwest1": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("europe-southwest1-b"), cty.StringVal("europe-southwest1-a"), cty.StringVal("europe-southwest1-c")}),
	}),
	"europe-west1": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("europe-west1-b"), cty.StringVal("europe-west1-c"), cty.StringVal("europe-west1-d")}),
	}),
	"europe-west12": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("europe-west12-c"), cty.StringVal("europe-west12-a"), cty.StringVal("europe-west12-b")}),
	}),
	"europe-west2": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("europe-west2-a"), cty.StringVal("europe-west2-b"), cty.StringVal("europe-west2-c")}),
	}),
	"europe-west3": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("europe-west3-c"), cty.StringVal("europe-west3-a"), cty.StringVal("europe-west3-b")}),
	}),
	"europe-west4": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("europe-west4-c"), cty.StringVal("europe-west4-b"), cty.StringVal("europe-west4-a")}),
	}),
	"europe-west6": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("europe-west6-b"), cty.StringVal("europe-west6-c"), cty.StringVal("europe-west6-a")}),
	}),
	"europe-west8": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("europe-west8-a"), cty.StringVal("europe-west8-b"), cty.StringVal("europe-west8-c")}),
	}),
	"europe-west9": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("europe-west9-b"), cty.StringVal("europe-west9-a"), cty.StringVal("europe-west9-c")}),
	}),
	"me-central1": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("me-central1-a"), cty.StringVal("me-central1-b"), cty.StringVal("me-central1-c")}),
	}),
	"me-west1": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("me-west1-b"), cty.StringVal("me-west1-a"), cty.StringVal("me-west1-c")}),
	}),
	"northamerica-northeast1": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("northamerica-northeast1-a"), cty.StringVal("northamerica-northeast1-b"), cty.StringVal("northamerica-northeast1-c")}),
	}),
	"northamerica-northeast2": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("northamerica-northeast2-b"), cty.StringVal("northamerica-northeast2-a"), cty.StringVal("northamerica-northeast2-c")}),
	}),
	"southamerica-east1": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("southamerica-east1-a"), cty.StringVal("southamerica-east1-b"), cty.StringVal("southamerica-east1-c")}),
	}),
	"southamerica-west1": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("southamerica-west1-a"), cty.StringVal("southamerica-west1-b"), cty.StringVal("southamerica-west1-c")}),
	}),
	"us-central1": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("us-central1-a"), cty.StringVal("us-central1-b"), cty.StringVal("us-central1-c"), cty.StringVal("us-central1-f")}),
	}),
	"us-east1": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("us-east1-b"), cty.StringVal("us-east1-c"), cty.StringVal("us-east1-d")}),
	}),
	"us-east4": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("us-east4-a"), cty.StringVal("us-east4-b"), cty.StringVal("us-east4-c")}),
	}),
	"us-east5": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("us-east5-c"), cty.StringVal("us-east5-b"), cty.StringVal("us-east5-a")}),
	}),
	"us-south1": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("us-south1-c"), cty.StringVal("us-south1-a"), cty.StringVal("us-south1-b")}),
	}),
	"us-west1": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("us-west1-a"), cty.StringVal("us-west1-b"), cty.StringVal("us-west1-c")}),
	}),
	"us-west2": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("us-west2-c"), cty.StringVal("us-west2-b"), cty.StringVal("us-west2-a")}),
	}),
	"us-west3": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("us-west3-a"), cty.StringVal("us-west3-b"), cty.StringVal("us-west3-c")}),
	}),
	"us-west4": cty.ObjectVal(map[string]cty.Value{
		"names": cty.ListVal([]cty.Value{cty.StringVal("us-west4-c"), cty.StringVal("us-west4-a"), cty.StringVal("us-west4-b")}),
	}),
}
