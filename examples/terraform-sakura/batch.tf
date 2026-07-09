# Dedicated AMD — batch server (コア専有プラン / AMD EPYC)
# Set cpu_model to an EPYC identifier to select the AMD dedicated plan.
resource "sakura_server" "batch" {
  name       = "batch"
  zone       = "is1a"
  core       = 32  # <<< Available: 32,64,128,192
  memory     = 120 # <<< Available: 120,240,480,1024
  commitment = "dedicated"
  cpu_model  = "epyc_7003" # <<< epyc_7003 / epyc_9004
}
