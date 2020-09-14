#!/bin/bash
gq https://pricing.infracost.io/graphql -q "
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
}" | jq -rn '[inputs] | map(.data.products[].attributes) | flatten | map(.key+"="+.value) | unique | join("\n")'
