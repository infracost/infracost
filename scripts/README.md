# Scripts

For adding a new resource to supported resources, you have to create a productFilter which uniquely identifies the product (a unique productHash). For this end, you need to add service, product family and attribute key values to filter the products.

But for a new terraform resource, the exact available values for filtering the query is unknown. If you use graphQL to query the resources, you would receive duplicate values for each query, and can not easily identify the available options for each key. Although it is possible to use deduplicators such as [graphql-deduplicator](https://github.com/gajus/graphql-deduplicator) to modify the backend schema, but it is also possible to easily remove the duplicates from the received response in the client side with jq tool.

These scripts use [graphqurl](https://github.com/hasura/graphqurl) and jq, to ease determining the productFilter which uniquely identifies the product.

## Usage

* On MacOS prepare the requirements with 
```sh
source requrements.sh
```

* Identify the `service` name with searching some keywords for the resource name in the json file downloaded from aws api. The reason we do not query the graphQL API, is that it only returns a limited number of responses, and as there are LOTS of products in aws, our desired resource service name might not be included in the returned responses (try with `AmazonRDS` and you'll see).

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
