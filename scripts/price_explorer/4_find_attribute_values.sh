#!/bin/sh
attribs=$(
gq https://pricing.api.infracost.io/graphql -H "x-api-key: $INFRACOST_API_KEY" -q "
query {
    products (
    filter: {
      vendorName: \"aws\"
      region: \"us-east-1\"
      service: \"$1\"
      productFamily: \"$2\"
    }
  ){
    	productHash
    	attributes { key , value }
    }
}" | jq -r --arg k1 "$3" '.data.products[].attributes[] | select(.key==$k1)' | jq '.value'
)

echo "$attribs" | sort | uniq
