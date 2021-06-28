datacenter = "{{ .DCName }}"
data_dir = "/opt/consul"
{{ if .GossipKey }}
encrypt = "{{ .GossipKey }}"
{{ end }}
verify_incoming = true
verify_outgoing = true
verify_server_hostname = true
{{ if .CACertFile }}
ca_file = "/etc/consul.d/{{ .CACertFile }}"
{{ end }}
{{ if .CertFile }}
cert_file = "/etc/consul.d/{{ .CertFile }}"
{{ end }}
{{ if .KeyFile }}
key_file = "/etc/consul.d/{{ .KeyFile }}"
{{ end }}
client_addr = "127.0.0.1 {{ .Address }}"
acl = {
  {{ if .ACLEnabled }}
  enabled = {{ .ACLEnabled }}
  {{ end }}
  default_policy = "allow"
  enable_token_persistence = true
}
performance {
  raft_multiplier = 1
}
