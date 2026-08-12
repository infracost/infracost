# Free resource (no cost)
resource "sakura_ssh_key" "deploy" {
  name       = "deploy-key"
  public_key = "ssh-rsa AAAA..."
}
