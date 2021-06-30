datacenter = "{{ .DCName }}"
data_dir = "/opt/nomad"
bind_addr = "{{ .Address }}"
tls {
  http = false
  rpc = false
  verify_server_hostname = false
  verify_https_client = false
}
acl = {
  enabled = false
}