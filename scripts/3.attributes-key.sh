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
}" | jq -r '.data.products[] | @base64' | head -2
)

i=0
for x in $attribs; do
    echo $x | base64 --decode | jq '.attributes | map ({ (.key): .value} ) | add' | jq ".$3"  | tee tmp-$i >/dev/null
    i=$(expr $i + 1)
done
echo Found $i different products
echo "\n#####################################\n"

if [ $i -ne 1 ]; then
     diff -U1 tmp-0 tmp-1 | grep -v "__typename"
fi

