#!/bin/sh
echo "List of Product Families for $1:"
gq https://pricing.api.infracost.io/graphql -H "x-api-key: $INFRACOST_API_KEY" -q "
query {
    products (
    filter: {
      vendorName: \"aws\"
      region: \"us-east-1\"
      service: \"$1\"
    }
  ){
        productFamily
    }
}" | jq '.data.products | map ({ (.productFamily): .__typename} ) | add' | jq "keys"
