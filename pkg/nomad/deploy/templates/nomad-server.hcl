server {
    enabled = true
    encrypt = "{{ .GossipKey }}"
    bootstrap_expect = 1
}