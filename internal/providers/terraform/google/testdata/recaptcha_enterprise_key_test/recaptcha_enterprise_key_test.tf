provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "global"
}

resource "google_recaptcha_enterprise_key" "default" {
  display_name      = "test-recaptcha"
  web_settings {
    integration_type = "SCORE"
  }
}

resource "google_recaptcha_enterprise_key" "android" {
  display_name = "android-recaptcha"
  android_settings {}
}
