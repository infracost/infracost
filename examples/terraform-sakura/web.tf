# Web servers × 2 (standard plan)
resource "sakura_server" "web" {
  count  = 2
  name   = "web-${count.index + 1}"
  zone   = "is1a"
  core   = 4 # <<< Try changing to 2 or 8 to compare costs
  memory = 8
}

resource "sakura_disk" "web_boot" {
  count = 2
  name  = "web-boot-${count.index + 1}"
  zone  = "is1a"
  plan  = "ssd"
  size  = 100
}
