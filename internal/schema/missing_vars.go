package schema

func ExtractMissingVarsCausingMissingAttributeKeys(r *ResourceData, attribute string) []string {
	var missing []string
	if raw := r.Metadata["attributesWithUnknownKeys"]; raw.IsArray() {
		for _, el := range raw.Array() {
			if el.Get("attribute").String() == attribute {
				if vars := el.Get("missingVariables"); vars.IsArray() {
					for _, v := range vars.Array() {
						missing = append(missing, v.String())
					}
				}
			}
		}
	}
	return missing
}
