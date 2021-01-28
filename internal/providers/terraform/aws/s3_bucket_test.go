package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestS3Bucket(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_s3_bucket" "bucket1" {
			bucket = "bucket1"

			lifecycle_rule {
				enabled = true
				tags = {
					Key = "value"
				}

				transition {
					storage_class = "INTELLIGENT_TIERING"
				}
				transition {
					storage_class = "ONEZONE_IA"
				}
				transition {
					storage_class = "STANDARD_IA"
				}
				transition {
					storage_class = "GLACIER"
				}
				transition {
					storage_class = "DEEP_ARCHIVE"
				}
			}
		}
		`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_s3_bucket.bucket1",
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "Standard",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:             "Storage",
							PriceHash:        "4ba817554541b6117c5552e95da3b08f-ee3dd7e4624338037ca6fea0933a662f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "PUT, COPY, POST, LIST requests",
							PriceHash:        "3fe44c30e48417d7bd3cdf4cf7dbc9dc-4a9dfd3965ffcbab75845ead7a27fd47",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "GET, SELECT, and all other requests",
							PriceHash:        "ab4ea46a6f426c447d2a36320c5022a8-4a9dfd3965ffcbab75845ead7a27fd47",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Select data scanned",
							PriceHash:        "1b6fcf7c66df50085e1826b97b561a9e-b1ae3861dc57e2db217fa83a7420374f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Select data returned",
							PriceHash:        "e957e3b952731768ec3540012b7f595e-b1ae3861dc57e2db217fa83a7420374f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
					},
				},
				{
					Name: "Intelligent tiering",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:             "Storage (frequent access)",
							PriceHash:        "07a918cb6025bc4db80666ef0e66e0b5-ee3dd7e4624338037ca6fea0933a662f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Storage (infrequent access)",
							PriceHash:        "5c80e08ae0fc926e1adb9112a6efcb2e-ee3dd7e4624338037ca6fea0933a662f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Monitoring and automation",
							PriceHash:        "6b6cf991e1be847e4bdbd53967993953-262e24dae0e085b444e6d3d16fd79991",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "PUT, COPY, POST, LIST requests",
							PriceHash:        "d9fa06614ed222dad3872b565efc737a-4a9dfd3965ffcbab75845ead7a27fd47",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "GET, SELECT, and all other requests",
							PriceHash:        "1ff61e05cac488b66b6246ea9ad0a363-4a9dfd3965ffcbab75845ead7a27fd47",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Lifecycle transition",
							PriceHash:        "d93a2e6333a84b980013338f23ab8469-4a9dfd3965ffcbab75845ead7a27fd47",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Select data scanned",
							PriceHash:        "f68c09c5b9469794cba7a64eb13a1d7b-b1ae3861dc57e2db217fa83a7420374f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Select data returned",
							PriceHash:        "6be95b5dda932a3fe7ff1a18cd5eaefd-b1ae3861dc57e2db217fa83a7420374f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Early delete (within 30 days)",
							PriceHash:        "dd7cde29892824e45609de4c9f425dc9-ee3dd7e4624338037ca6fea0933a662f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
					},
				},
				{
					Name: "Standard - infrequent access",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:             "Storage",
							PriceHash:        "fb14131eed9cc90c1a1b39af0241c9c1-ee3dd7e4624338037ca6fea0933a662f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "PUT, COPY, POST, LIST requests",
							PriceHash:        "446853cc37182cdc098bcfddeedbefea-4a9dfd3965ffcbab75845ead7a27fd47",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "GET, SELECT, and all other requests",
							PriceHash:        "1dcad450c58fe98b413a3a8570487bb6-4a9dfd3965ffcbab75845ead7a27fd47",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Lifecycle transition",
							PriceHash:        "d93a2e6333a84b980013338f23ab8469-4a9dfd3965ffcbab75845ead7a27fd47",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Retrievals",
							PriceHash:        "9bad25b8553bfa85e759c45bfb1cb3ed-b1ae3861dc57e2db217fa83a7420374f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Select data scanned",
							PriceHash:        "86473721cbbf40120e222f2b642530a4-b1ae3861dc57e2db217fa83a7420374f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Select data returned",
							PriceHash:        "087d756dfa45c575c6549a8417735cb4-b1ae3861dc57e2db217fa83a7420374f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
					},
				},
				{
					Name: "One zone - infrequent access",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:             "Storage",
							PriceHash:        "d4ca3a7b652c57630e78ed91933d869f-ee3dd7e4624338037ca6fea0933a662f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "PUT, COPY, POST, LIST requests",
							PriceHash:        "d3387b07454cd9fcfa7e87cc975e93fc-4a9dfd3965ffcbab75845ead7a27fd47",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "GET, SELECT, and all other requests",
							PriceHash:        "6209cc28c45a4fe4d7c3c48a67927880-4a9dfd3965ffcbab75845ead7a27fd47",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Lifecycle transition",
							PriceHash:        "d93a2e6333a84b980013338f23ab8469-4a9dfd3965ffcbab75845ead7a27fd47",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Retrievals",
							PriceHash:        "e68e8504841bd960121ff7990856ed42-b1ae3861dc57e2db217fa83a7420374f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Select data scanned",
							PriceHash:        "f4eb8a1865fe84b11c7cc3e26758d28a-b1ae3861dc57e2db217fa83a7420374f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Select data returned",
							PriceHash:        "98e3336b80553b54ae34e57a6cafd153-b1ae3861dc57e2db217fa83a7420374f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
					},
				},
				{
					Name: "Glacier",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:             "Storage",
							PriceHash:        "91a960e141a8c8ae5ac93089e4be4fd8-ee3dd7e4624338037ca6fea0933a662f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "PUT, COPY, POST, LIST requests",
							PriceHash:        "02e559a7fcbd024b096e7f9b0db98641-4a9dfd3965ffcbab75845ead7a27fd47",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "GET, SELECT, and all other requests",
							PriceHash:        "ea2236cb67bbac53413cd6256bdc0ba1-4a9dfd3965ffcbab75845ead7a27fd47",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Lifecycle transition",
							PriceHash:        "abe557ab4bd5a4e641a60a8fa8afdb06-4a9dfd3965ffcbab75845ead7a27fd47",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Retrieval requests (standard)",
							PriceHash:        "5d96c9b8fc8ca9c1d8703d567a8fc434-4a9dfd3965ffcbab75845ead7a27fd47",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Retrievals (standard)",
							PriceHash:        "da657ccc1280c9a13ca71774fa3f51f0-b1ae3861dc57e2db217fa83a7420374f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Select data scanned (standard)",
							PriceHash:        "e55a46fd87bf1b72140651e2b9cfb03a-b1ae3861dc57e2db217fa83a7420374f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Select data returned (standard)",
							PriceHash:        "3b592b5c70031c358750a4ccdbd16283-b1ae3861dc57e2db217fa83a7420374f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Retrieval requests (expedited)",
							PriceHash:        "b09d424f71a23453052c628130370119-4a9dfd3965ffcbab75845ead7a27fd47",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Retrievals (expedited)",
							PriceHash:        "75747e85ffac1c226f7c9190906e1f4d-b1ae3861dc57e2db217fa83a7420374f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Select data scanned (expedited)",
							PriceHash:        "f1e9ce6c7901c74746be2246423ab404-b1ae3861dc57e2db217fa83a7420374f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Select data returned (expedited)",
							PriceHash:        "9ece5b97875e992a43f56191699637da-b1ae3861dc57e2db217fa83a7420374f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Retrieval requests (bulk)",
							PriceHash:        "a9a178624e239644ce18f4c25832bd75-4a9dfd3965ffcbab75845ead7a27fd47",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Retrievals (bulk)",
							PriceHash:        "08c5d2effc2eafae2bcb360edfaf2d4a-b1ae3861dc57e2db217fa83a7420374f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Select data scanned (bulk)",
							PriceHash:        "257a754d6bac69b65dee527ee33e4a88-b1ae3861dc57e2db217fa83a7420374f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Select data returned (bulk)",
							PriceHash:        "f5ac4db39024038525176ed3bd064788-b1ae3861dc57e2db217fa83a7420374f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Early delete (within 90 days)",
							PriceHash:        "baf40d845c528f1c2bdfe7527c76570a-ee3dd7e4624338037ca6fea0933a662f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
					},
				},
				{
					Name: "Glacier deep archive",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:             "Storage",
							PriceHash:        "e20face2e1716d8b58fec20e7ac85098-ee3dd7e4624338037ca6fea0933a662f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "PUT, COPY, POST, LIST requests",
							PriceHash:        "6fefcf21312acdbdee7dfe139154077e-4a9dfd3965ffcbab75845ead7a27fd47",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "GET, SELECT, and all other requests",
							PriceHash:        "20b342876db2209498bf0eb7a70514c9-4a9dfd3965ffcbab75845ead7a27fd47",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Lifecycle transition",
							PriceHash:        "20713d1325a3376f28f974ae8a82d98f-4a9dfd3965ffcbab75845ead7a27fd47",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Retrieval requests (standard)",
							PriceHash:        "95057fe74650d73b6e19595d5f6a08ac-4a9dfd3965ffcbab75845ead7a27fd47",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Retrievals (standard)",
							PriceHash:        "4e43fda9d741fd54b738af650fe1cb9f-b1ae3861dc57e2db217fa83a7420374f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Retrieval requests (bulk)",
							PriceHash:        "92f98e6a8b3759c23f4e358da7ed4b7c-4a9dfd3965ffcbab75845ead7a27fd47",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Retrievals (bulk)",
							PriceHash:        "b5b39edbf4353c1344e944384e98f1f1-b1ae3861dc57e2db217fa83a7420374f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
						{
							Name:             "Early delete (within 180 days)",
							PriceHash:        "0ae73df3791e4d7eae2e06d30ed29359-ee3dd7e4624338037ca6fea0933a662f",
							MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
						},
					},
				},
			},
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Object tagging",
					PriceHash:        "64620ac9fef800f3dc85d1c151b32739-433973798398d654710ea15482359393",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
