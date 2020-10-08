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

3. Set the INFRACOST_API_KEY env variable:

  ```
  export INFRACOST_API_KEY=my_api_key_here
  ```

## Usage

* List all available AWS `service` names, or find a `service` name by searching for a keyword or acronym in the JSON file downloaded in the setup step. AWS use many acronyms so be sure to search for those too, e.g. "ES" returns "AmazonES" for ElasticSearch. The reason we do not query the pricing service graphQL API is that it only returns a limited number of responses, and as there are lots of products in AWS, our desired service name might not be included in the returned responses.

  List all available AWS service names:

  ```sh
  ./1_find_services.sh
  ```

  Search for service name using a keyword/acronym:

  ```sh
  ./1_find_services.sh <keyword>
  ```

  For example, running `./1_find_services.sh rds` returns:

  ```sh
  "offerCode" : "AmazonRDS",
  ```

* Find the desired `productFamily` using the `service` name found in the previous step:

  ```sh
  ./2_find_product_families.sh <service-name>
  ```

  For example, running `./2_find_product_families.sh AmazonRDS` returns:

    ```sh
    {
      "Aurora Global Database",
      "CPU Credits",
      "Database Instance"
    }
    ```

* Use scripts 3 with 4 to explore attribute keys and values separately, or script 5 to list all key-attribute pairs. This is useful if you want to grep for things such as `grep -i usagetype` or `grep -i engine`.

  * Find the desired `attributes` keys using the `service` name and `productFamily` found in previous steps:

    ```sh
    ./3_find_attribute_keys.sh <service-name> <product-family>
    ```

    For example, running `./3_find_attribute_keys.sh AmazonRDS "Database Instance"` returns the following abbreviated output, which shows there are different `instanceType` and `databaseEngine` attribute:

      ```sh
      Found 1000 different products
      ------------------
      
      @@ -18,3 +18,3 @@
          "key": "instanceType",
      -    "value": "db.m3.large",
      +    "value": "db.r4.4xlarge",
      @@ -28,3 +28,3 @@
          "key": "databaseEngine",
      -    "value": "MySQL",
      +    "value": "MariaDB",
      ...
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

  * List all key-value pairs for the desired `attributes` value using the `service` name and `productFamily`:

    ```sh
    ./5_find_all_attribute_key_values.sh <service-name> <product-family>
    ```

    For example, running `./5_find_all_attribute_key_values.sh AmazonRDS "Database Instance"` returns:

      ```sh
      clockSpeed=2.3 GHz
      ...
      databaseEdition=Standard
      ...
      databaseEngine=SQL Server
      ...
      usagetype=Multi-AZUsage:db.m6g.4xl
      ```

* When writing integration tests, a `PriceHash` is used to match the price of a cost component so the tests continue to pass even if the actual price value changes. To find a `PriceHash`:

  1. Install the [modheader](https://bewisse.com/modheader/) browser extension so you can set custom HTTP headers in your browser.
  2. Use the modheader browser extension to set `x-api-key` to your Infracost API key.
  3. Browse to [https://pricing.api.infracost.io/graphql](https://pricing.api.infracost.io/graphql) and use the following query with your desired variable values:

  * query:

  ```gql
  query($productFilter: ProductFilter!, $priceFilter: PriceFilter) {
    products(filter: $productFilter) {
      prices(filter: $priceFilter) {
        priceHash
        USD
      }
    }
  }
  ```

  * variables (edit as required):

  ```gql
  {
    "priceFilter": {
      "purchaseOption": "on_demand"
    },
    "productFilter": {
      "service": "AmazonRDS",
      "productFamily": "Database Instance",
      "region": "us-east-1",
      "attributeFilters": [
        {
          "key": "instanceType",
          "value": "db.t3.medium"
        },
        {
          "key": "databaseEngine",
          "value": "MariaDB"
        },
        {
          "key": "deploymentOption",
          "value": "Single-AZ"
        }
      ]
    }
  }
  ```

  The above returns the following result, which you can manually verify by browsing to the AWS pricing page for the above product/filters. For example, [https://aws.amazon.com/rds/mariadb/pricing/](https://aws.amazon.com/rds/mariadb/pricing/) shows the hourly price as $0.068, which matches the following JSON:

  ```json
  {
    "data": {
      "products": [
        {
          "prices": [
            {
              "priceHash": "72dd84252654545b6f19c2e553efecde-d2c98780d7b6e36641b521f1f8145c6f",
              "USD": "0.0680000000"
            }
          ]
        }
      ]
    }
  }
  ```
