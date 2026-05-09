terraform {
  required_providers {
    sakura = {
      source  = "sacloud/sakura"
      version = "~> 0.1"
    }
  }
}

provider "sakura" {
  zone = "is1a"
}
