client {
  options {
    "docker.privileged.enabled" = "true"
  }
  enabled = true
}
plugin "raw_exec" {
  config {
    enabled = true
  }
}