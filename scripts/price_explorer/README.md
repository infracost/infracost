# Scripts

When adding a new resource to infracost, a `productFilter` has to be added that uniquely identifies the product for pricing purposes (the filter is used to find a unique `productHash`). The filter needs a `service`, `productFamily` and `attribute` key/values to filter the prices. Because cloud prices have many parameter values to determine a price on a per-region or product type basis, querying for prices can look odd as many parameter values are duplicated. These scripts help to explore the pricing data and find the correct price by using [graphqurl](https://github.com/hasura/graphqurl) and `jq` to determine the `productFilter` that uniquely identifies the product.

## Setup

1. These scripts require `jq`, `node` and the npm module `graphqurl`. On MacOS, the following should work:

  ```sh
  brew install jq
  brew install node
  npm install -g graphqurl
  ```

2. Download the AWS pricing index file:

  ```sh
  wget https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/index.json
  ```

## Usage

* Find the `service` name by searching for a keyword for a name in the JSON file downloaded in the setup step. The reason we do not query the pricing service graphQL API is that it only returns a limited number of responses, and as there are lots of products in AWS, our desired resource service name might not be included in the returned responses.

  ```sh
  ./1_find_services.sh <key-word>
  ```

  For example, running `./1_find_services.sh rds` returns:

    ```
    "AmazonRDS" : {
      "offerCode" : "AmazonRDS",
      "versionIndexUrl" : "/offers/v1.0/aws/AmazonRDS/index.json",
      "currentVersionUrl" : "/offers/v1.0/aws/AmazonRDS/current/index.json",
      "currentRegionIndexUrl" : "/offers/v1.0/aws/AmazonRDS/current/region_index.json"
    ```

* Find the desired `productFamily` using the `service` name found in the previous step:

  ```sh
  ./2_find_product_families.sh <service-name>
  ```

  For example, running `./2_find_product_families.sh AmazonRDS` returns:

    ```sh
    {
      "Aurora Global Database": "Product",
      "CPU Credits": "Product",
      "Database Instance": "Product"
    }
    ```

* Find the desired `attributes` keys using the `service` name and `productFamily` found in previous steps:

  ```sh
  ./3_find_attribute_keys.sh <service-name> <product-family>
  ```

  For example, running `./3_find_attribute_keys.sh AmazonRDS "Database Instance"` returns the following abbreviated output, which shows there are different `instanceType` and `databaseEngine` attribute:

    ```sh
    -  "instanceType": "db.m3.large",
    +  "instanceType": "db.r4.4xlarge",
    ...
    -  "databaseEngine": "Oracle",
    +  "databaseEngine": "MariaDB",
    ```

* Identify desired `attributes` value using the `service` name, `productFamily` and `attribute key` found in previous steps:

  ```sh
  ./4_find_attribute_values.sh <service-name> <product-family> <attribute-key>
  ```

  For example, running `./4_find_attribute_values.sh AmazonRDS "Database Instance" databaseEngine` returns the unique values for the `databaseEngine` attribute:

    ```sh
    "Aurora MySQL"
    "Aurora PostgreSQL"
    "MariaDB"
    "MySQL (on-premise for Outpost)"
    "MySQL"
    "Oracle"
    "PostgreSQL (on-premise for Outposts)"
    "PostgreSQL"
    "SQL Server"
    ```
