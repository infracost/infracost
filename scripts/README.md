# Scripts

When adding a new resource to infracost, a `productFilter` has to be added that uniquely identifies the product (a unique `productHash`) for pricing purposes. The filter needs a `service`, `productFamily` and `attribute` key/values to filter the prices. Because cloud prices have many parameter values to determine a price on a per-region or product type basis, querying for prices can look odd as many parameter values are duplicated. These scripts help find the correct price by using [graphqurl](https://github.com/hasura/graphqurl) and `jq` to determe the `productFilter` that uniquely identifies the product.

## Usage

* On MacOS prepare the requirements with 
```sh
source requrements.sh
```

* Identify the `service` name by searching some keywords for the resource name in the json file downloaded from the AWS API. The reason we do not query the graphQL API, is that it only returns a limited number of responses, and as there are lots of products in AWS, our desired resource service name might not be included in the returned responses (try with `AmazonRDS` and you'll see).

```sh
./1.service.sh <key-word> ## e.g. rds
```

* Identify desired `productFamily` using the `service` name found in the prev step:

```sh
./2.productFamily.sh <service-name>
```

* Identify desired `attributes` key using the `service` name and `productFamily` found in prev steps:

```sh
./3.attributes-key.sh <service-name> <product-family>
```

* Identify desired `attributes` value using the `service` name, `productFamily` and `attribute key` found in prev steps:

```sh
./4.attributes-value.sh <service-name> <product-family> <attribute-key>
```
