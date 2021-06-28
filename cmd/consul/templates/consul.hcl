datacenter = "dc1"
data_dir = "/opt/consul"
encrypt = "{{ .GossipKey }}"
verify_incoming = true
verify_outgoing = true
verify_server_hostname = true
ca_file = "/etc/consul.d/{{ .CACertFile }}"
cert_file = "/etc/consul.d/{{ .CertFile }}"
key_file = "/etc/consul.d/{{ .KeyFile }}"
client_addr = "127.0.0.1 {{ .Hostname }}"
acl = {
  enabled = false
  default_policy = "allow"
  enable_token_persistence = true
}
performance {
  raft_multiplier = 1
}