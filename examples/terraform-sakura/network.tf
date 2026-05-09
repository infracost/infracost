# Router + Switch (250Mbps)
resource "sakura_internet" "main" {
  name       = "main-router"
  zone       = "is1a"
  band_width = 250 # <<< Try changing to 100, 500, 1000 to compare costs
}

resource "sakura_switch" "internal" {
  name = "internal-sw"
  zone = "is1a"
}
