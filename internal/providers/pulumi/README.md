# Pulumi Support for Infracost

This directory contains the implementation for Pulumi support in Infracost. Pulumi is an Infrastructure as Code tool that allows users to define cloud resources using general-purpose programming languages.

## Architecture

The Pulumi support is implemented with the following components:

1. **Preview JSON Provider** - Parses the output of `pulumi preview --json` to extract resource information.
2. **Resource Implementations** - Converted from Terraform resources to handle Pulumi's camelCase attributes.
3. **Registry Files** - Maps Pulumi resource types (e.g. `aws:s3/bucket:Bucket`) to their corresponding resource implementations.

## Testing

We provide two ways to test Pulumi resources:

### 1. Resource Tests

Use the `putest.ResourceTests()` function to create inline tests with JSON and expected cost component checks:

```go
func TestComputeInstance(t *testing.T) {
    pulumiJSON := `{
        "steps": [{
            "resource": {
                "type": "gcp:compute/instance:Instance",
                "name": "my-instance",
                "properties": {
                    // Pulumi resource properties
                }
            }
        }]
    }`

    usage := schema.NewUsageMap(map[string]interface{}{
        "my-instance": map[string]interface{}{
            "monthly_hours": 730,
        },
    })

    resourceChecks := []testutil.ResourceCheck{
        // Expected cost components
    }

    putest.ResourceTests(t, pulumiJSON, usage, resourceChecks)
}
```

### 2. Golden File Tests

Use the `putest.GoldenFileResourceTests()` function to create tests that compare against golden files:

```go
func TestS3BucketGoldenFile(t *testing.T) {
    putest.GoldenFileResourceTests(t, "s3_bucket")
}
```

This requires a directory structure:
```
testdata/
  s3_bucket/
    s3_bucket.json       # Pulumi preview JSON file
    s3_bucket.usage.yml  # Optional usage file
    s3_bucket.golden     # Golden file output to compare against
```

## Running Tests

To run the tests for a specific provider:

```shell
go test -v ./internal/providers/pulumi/aws/...
go test -v ./internal/providers/pulumi/azure/...
go test -v ./internal/providers/pulumi/google/...
```

Or to run all Pulumi tests:

```shell
go test -v ./internal/providers/pulumi/...
```

## Differences from Terraform

1. **Attribute Names**: Pulumi uses camelCase for attribute names, while Terraform uses snake_case.
2. **URN References**: Pulumi uses URNs (Uniform Resource Names) instead of IDs for references.
3. **JSON Format**: The Pulumi preview JSON format is different from Terraform plan JSON.
4. **Cloud Provider Prefixes**: Pulumi uses prefixes like `aws:`, `azure:`, `gcp:` in resource types.