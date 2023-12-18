provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_kms_crypto_key" "my_keys" {
  name     = "crypto-key-example"
  key_ring = ""
}

resource "google_kms_crypto_key" "my_keys_withUsage" {
  name     = "crypto-key-example"
  key_ring = ""

  version_template {
    algorithm        = "EC_SIGN_P256_SHA256"
    protection_level = "HSM"
  }
}
resource "google_kms_crypto_key" "with_rotate_withUsage" {
  name            = "crypto-key-example"
  key_ring        = ""
  rotation_period = "100000s"
}
