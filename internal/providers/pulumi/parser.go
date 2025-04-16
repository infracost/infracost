package pulumi

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

// Parser parses a Pulumi preview.json file
type Parser struct {
	previewJSON         []byte
	ctx                 *config.ProjectContext
	usage               schema.UsageMap
	includePastResources bool
}

// parseResources extracts resources from a Pulumi preview.json file
func (p *Parser) parseResources() ([]*schema.Resource, error) {
	var out []*schema.Resource

	j := gjson.ParseBytes(p.previewJSON)
	
	// Check if the JSON format is valid
	if !j.IsObject() {
		return out, fmt.Errorf("invalid Pulumi preview JSON file format")
	}

	// Extract resources from the preview JSON
	resources := j.Get("steps").Array()
	for _, res := range resources {
		// Skip if not a resource operation (could be a stack or other operation)
		if !res.Get("resource").Exists() {
			continue
		}
		
		// Get resource type
		resourceType := res.Get("resource.type").String()
		
		// Skip resources we don't support
		item := p.lookupRegistry(resourceType)
		if item == nil {
			log.Debugf("Skipping resource %s", resourceType)
			continue
		}
		
		// Create resource data
		addr := res.Get("resource.name").String()
		d := schema.NewResourceDataFromProviderPulumi(
			resourceType,
			addr,
			res,
		)
		d.PulumiUrn = res.Get("resource.urn").String()
		
		// Find actual resource function and build the resource 
		if item.RFunc != nil {
			if u := p.usage[addr]; u != nil {
				out = append(out, item.RFunc(d, u))
			} else {
				out = append(out, item.RFunc(d, nil))
			}
		}
	}

	return out, nil
}

func (p *Parser) lookupRegistry(resourceType string) *schema.RegistryItem {
	// Try direct lookup
	for _, r := range ResourceRegistry {
		if resourceType == r.Name {
			return r
		}
	}
	
	// Try normalized lookup (convert aws:s3/bucket:Bucket to aws_s3_bucket)
	normalizedType := strings.ReplaceAll(resourceType, ":", "_")
	normalizedType = strings.ReplaceAll(normalizedType, "/", "_")
	
	for _, r := range ResourceRegistry {
		if normalizedType == r.Name {
			return r
		}
	}
	
	return nil
}
