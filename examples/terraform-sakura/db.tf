# Dedicated Intel — DB server (コア専有プラン / Intel Xeon)
# commitment = "dedicated" selects the dedicated core plan.
resource "sakura_server" "db" {
  name       = "db"
  zone       = "is1a"
  core       = 8  # <<< Available: 2,4,6,8,10,24
  memory     = 32 # <<< Available: 4,8,16,24,32,48,96,192
  commitment = "dedicated"
}

resource "sakura_disk" "db_boot" {
  name = "db-boot"
  zone = "is1a"
  plan = "ssd"
  size = 40
}

resource "sakura_disk" "db_data" {
  name = "db-data"
  zone = "is1a"
  plan = "ssd"
  size = 500
}
