# AppRun shared (共有プラン)
# Cost is estimated as: hourly_rate × 730h × max_scale (upper bound).
# Actual cost is lower when traffic is low and instances scale to zero.
resource "sakura_apprun_shared" "api" {
  name            = "api"
  timeout_seconds = 60
  port            = 8080
  min_scale       = 0 # 0 enables scale-to-zero
  max_scale       = 5 # <<< Try changing to compare upper-bound cost

  components = [{
    name       = "app"
    max_cpu    = "1"   # <<< 0.5 / 1 / 2
    max_memory = "2Gi" # <<< 1Gi / 2Gi / 4Gi
    deploy_source = {
      container_registry = {
        image = "example.sakuracr.jp/api:latest"
      }
    }
  }]
}
