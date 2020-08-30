#!/bin/sh
attribs=$(
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
}" | jq -r '.data.products[] | @base64'
)

i=0
for x in $attribs; do
    values=$(echo $x | base64 --decode | jq '.attributes | map ({ (.key): .value} ) | add' | jq ".$3")
    i=$(expr $i + 1)
done

echo $values | sort -u
