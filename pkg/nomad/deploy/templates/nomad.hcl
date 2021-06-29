datacenter = "{{ .DCName }}"
data_dir = "/opt/nomad"
bind_addr = "0.0.0.0"
encrypt = "{{ .GossipKey }}"
tls {
  http = false
  rpc = false
  verify_server_hostname = false
  verify_https_client = false
}
acl = {
  enabled = false
}
