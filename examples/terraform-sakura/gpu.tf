# GPU server — 高火力VRT / H100 (24 core / 240GB, Ishikari Zone 1 only)
resource "sakura_server" "gpu" {
  name      = "gpu"
  zone      = "is1a"
  core      = 24
  memory    = 240
  gpu       = 1
  gpu_model = "H100" # <<< H100 or V100
}
