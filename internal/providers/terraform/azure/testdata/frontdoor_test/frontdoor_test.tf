provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "FrontDoorExampleResourceGroup"
  location = "West Europe"
}

resource "azurerm_frontdoor" "frontdoor_without_usage" {
  name                = "exampleFrontdoor"
  resource_group_name = azurerm_resource_group.example.name

  routing_rule {
    name               = "exampleRoutingRule1"
    accepted_protocols = ["Http", "Https"]
    patterns_to_match  = ["/*"]
    frontend_endpoints = ["exampleFrontendEndpoint1"]
    forwarding_configuration {
      forwarding_protocol = "MatchRequest"
      backend_pool_name   = "exampleBackendBing"
    }
  }

  routing_rule {
    name               = "exampleDisabledRoutingRule"
    enabled            = false
    accepted_protocols = ["Http", "Https"]
    patterns_to_match  = ["/*"]
    frontend_endpoints = ["exampleFrontendEndpoint1"]
    forwarding_configuration {
      forwarding_protocol = "MatchRequest"
      backend_pool_name   = "exampleBackendBing"
    }
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint1"
    host_name = "example-frontdoor1.example.net"
  }

  # --- Required attributes ---
  backend_pool_load_balancing {
    name = "exampleLoadBalancingSettings1"
  }

  backend_pool_health_probe {
    name = "exampleHealthProbeSetting1"
  }

  backend_pool {
    name = "exampleBackendBing"
    backend {
      host_header = "www.bing.com"
      address     = "www.bing.com"
      http_port   = 80
      https_port  = 443
    }

    load_balancing_name = "exampleLoadBalancingSettings1"
    health_probe_name   = "exampleHealthProbeSetting1"
  }
}

resource "azurerm_frontdoor" "frontdoor_with_usage" {
  name                = "exampleFrontdoorWithUsage"
  resource_group_name = azurerm_resource_group.example.name

  routing_rule {
    name               = "exampleRoutingRule1"
    accepted_protocols = ["Http", "Https"]
    patterns_to_match  = ["/*"]
    frontend_endpoints = ["exampleFrontendEndpoint1"]
    forwarding_configuration {
      forwarding_protocol = "MatchRequest"
      backend_pool_name   = "exampleBackendBing"
    }
  }

  routing_rule {
    name               = "exampleRoutingRule2"
    accepted_protocols = ["Http", "Https"]
    patterns_to_match  = ["/*"]
    frontend_endpoints = ["exampleFrontendEndpoint1"]
    forwarding_configuration {
      forwarding_protocol = "MatchRequest"
      backend_pool_name   = "exampleBackendBing"
    }
  }

  routing_rule {
    name               = "exampleRoutingRule3"
    accepted_protocols = ["Http", "Https"]
    patterns_to_match  = ["/*"]
    frontend_endpoints = ["exampleFrontendEndpoint1"]
    forwarding_configuration {
      forwarding_protocol = "MatchRequest"
      backend_pool_name   = "exampleBackendBing"
    }
  }

  routing_rule {
    name               = "exampleRoutingRule4"
    accepted_protocols = ["Http", "Https"]
    patterns_to_match  = ["/*"]
    frontend_endpoints = ["exampleFrontendEndpoint1"]
    forwarding_configuration {
      forwarding_protocol = "MatchRequest"
      backend_pool_name   = "exampleBackendBing"
    }
  }

  routing_rule {
    name               = "exampleRoutingRule5"
    accepted_protocols = ["Http", "Https"]
    patterns_to_match  = ["/*"]
    frontend_endpoints = ["exampleFrontendEndpoint1"]
    forwarding_configuration {
      forwarding_protocol = "MatchRequest"
      backend_pool_name   = "exampleBackendBing"
    }
  }

  routing_rule {
    name               = "exampleRoutingRule6"
    accepted_protocols = ["Http", "Https"]
    patterns_to_match  = ["/*"]
    frontend_endpoints = ["exampleFrontendEndpoint1"]
    forwarding_configuration {
      forwarding_protocol = "MatchRequest"
      backend_pool_name   = "exampleBackendBing"
    }
  }

  routing_rule {
    name               = "exampleDisabledRoutingRule"
    enabled            = false
    accepted_protocols = ["Http", "Https"]
    patterns_to_match  = ["/*"]
    frontend_endpoints = ["exampleFrontendEndpoint1"]
    forwarding_configuration {
      forwarding_protocol = "MatchRequest"
      backend_pool_name   = "exampleBackendBing"
    }
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint1"
    host_name = "example-frontdoor1.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint2"
    host_name = "example-frontdoor2.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint3"
    host_name = "example-frontdoor3.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint4"
    host_name = "example-frontdoor4.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint5"
    host_name = "example-frontdoor5.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint6"
    host_name = "example-frontdoor6.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint7"
    host_name = "example-frontdoor7.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint8"
    host_name = "example-frontdoor8.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint9"
    host_name = "example-frontdoor9.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint10"
    host_name = "example-frontdoor10.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint11"
    host_name = "example-frontdoor11.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint12"
    host_name = "example-frontdoor12.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint13"
    host_name = "example-frontdoor13.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint14"
    host_name = "example-frontdoor14.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint15"
    host_name = "example-frontdoor15.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint16"
    host_name = "example-frontdoor16.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint17"
    host_name = "example-frontdoor17.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint18"
    host_name = "example-frontdoor18.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint19"
    host_name = "example-frontdoor19.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint20"
    host_name = "example-frontdoor20.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint21"
    host_name = "example-frontdoor21.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint22"
    host_name = "example-frontdoor22.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint23"
    host_name = "example-frontdoor23.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint24"
    host_name = "example-frontdoor24.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint25"
    host_name = "example-frontdoor25.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint26"
    host_name = "example-frontdoor26.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint27"
    host_name = "example-frontdoor27.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint28"
    host_name = "example-frontdoor28.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint29"
    host_name = "example-frontdoor29.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint30"
    host_name = "example-frontdoor30.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint31"
    host_name = "example-frontdoor31.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint32"
    host_name = "example-frontdoor32.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint33"
    host_name = "example-frontdoor33.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint34"
    host_name = "example-frontdoor34.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint35"
    host_name = "example-frontdoor35.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint36"
    host_name = "example-frontdoor36.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint37"
    host_name = "example-frontdoor37.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint38"
    host_name = "example-frontdoor38.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint39"
    host_name = "example-frontdoor39.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint40"
    host_name = "example-frontdoor40.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint41"
    host_name = "example-frontdoor41.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint42"
    host_name = "example-frontdoor42.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint43"
    host_name = "example-frontdoor43.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint44"
    host_name = "example-frontdoor44.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint45"
    host_name = "example-frontdoor45.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint46"
    host_name = "example-frontdoor46.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint47"
    host_name = "example-frontdoor47.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint48"
    host_name = "example-frontdoor48.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint49"
    host_name = "example-frontdoor49.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint50"
    host_name = "example-frontdoor50.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint51"
    host_name = "example-frontdoor51.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint52"
    host_name = "example-frontdoor52.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint53"
    host_name = "example-frontdoor53.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint54"
    host_name = "example-frontdoor54.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint55"
    host_name = "example-frontdoor55.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint56"
    host_name = "example-frontdoor56.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint57"
    host_name = "example-frontdoor57.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint58"
    host_name = "example-frontdoor58.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint59"
    host_name = "example-frontdoor59.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint60"
    host_name = "example-frontdoor60.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint61"
    host_name = "example-frontdoor61.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint62"
    host_name = "example-frontdoor62.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint63"
    host_name = "example-frontdoor63.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint64"
    host_name = "example-frontdoor64.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint65"
    host_name = "example-frontdoor65.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint66"
    host_name = "example-frontdoor66.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint67"
    host_name = "example-frontdoor67.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint68"
    host_name = "example-frontdoor68.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint69"
    host_name = "example-frontdoor69.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint70"
    host_name = "example-frontdoor70.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint71"
    host_name = "example-frontdoor71.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint72"
    host_name = "example-frontdoor72.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint73"
    host_name = "example-frontdoor73.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint74"
    host_name = "example-frontdoor74.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint75"
    host_name = "example-frontdoor75.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint76"
    host_name = "example-frontdoor76.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint77"
    host_name = "example-frontdoor77.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint78"
    host_name = "example-frontdoor78.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint79"
    host_name = "example-frontdoor79.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint80"
    host_name = "example-frontdoor80.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint81"
    host_name = "example-frontdoor81.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint82"
    host_name = "example-frontdoor82.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint83"
    host_name = "example-frontdoor83.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint84"
    host_name = "example-frontdoor84.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint85"
    host_name = "example-frontdoor85.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint86"
    host_name = "example-frontdoor86.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint87"
    host_name = "example-frontdoor87.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint88"
    host_name = "example-frontdoor88.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint89"
    host_name = "example-frontdoor89.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint90"
    host_name = "example-frontdoor90.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint91"
    host_name = "example-frontdoor91.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint92"
    host_name = "example-frontdoor92.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint93"
    host_name = "example-frontdoor93.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint94"
    host_name = "example-frontdoor94.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint95"
    host_name = "example-frontdoor95.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint96"
    host_name = "example-frontdoor96.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint97"
    host_name = "example-frontdoor97.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint98"
    host_name = "example-frontdoor98.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint99"
    host_name = "example-frontdoor99.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint100"
    host_name = "example-frontdoor100.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint101"
    host_name = "example-frontdoor101.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint102"
    host_name = "example-frontdoor102.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint103"
    host_name = "example-frontdoor103.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint104"
    host_name = "example-frontdoor104.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint105"
    host_name = "example-frontdoor105.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint106"
    host_name = "example-frontdoor106.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint107"
    host_name = "example-frontdoor107.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint108"
    host_name = "example-frontdoor108.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint109"
    host_name = "example-frontdoor109.example.net"
  }

  frontend_endpoint {
    name      = "exampleFrontendEndpoint110"
    host_name = "example-frontdoor110.example.net"
  }

  # --- Required attributes ---
  backend_pool_load_balancing {
    name = "exampleLoadBalancingSettings1"
  }

  backend_pool_health_probe {
    name = "exampleHealthProbeSetting1"
  }

  backend_pool {
    name = "exampleBackendBing"
    backend {
      host_header = "www.bing.com"
      address     = "www.bing.com"
      http_port   = 80
      https_port  = 443
    }

    load_balancing_name = "exampleLoadBalancingSettings1"
    health_probe_name   = "exampleHealthProbeSetting1"
  }
}
