# Dedicated host (専有ホストプラン)
# class = "dynamic" (standard) | "windows"
resource "sakura_private_host" "host" {
  name  = "dedicated-host"
  zone  = "is1a"
  class = "dynamic" # <<< Try "windows" to compare costs
}
