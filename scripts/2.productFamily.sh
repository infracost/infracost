#!/bin/sh
gq https://pricing.infracost.io/graphql -q "
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
}" | jq '.data.products | map ({ (.productFamily): .__typename} ) | add'
