# Hashicorp Nomad/Consul deployment tool
This is the CLI tool for deployment Hashicorp Nomad and Consul cluster, inspired by RKE
tool for Kubernetes provisioning. It provides interactive survey for generating
configuration file, and uses that configuration to spill Consul cluster
with optional gossip encryption enabled and securing API communication with https.

## Quick start

### Consul deployment
```console
    $ ./nomad-deploy consul config # take survey about target hosts and cluster options
    $ ./cat consul.yaml
    $ ./nomad-deploy consul up # deploy cluster
```

### Nomad deployment
Nomad deployment CLI is the same as for deploying consul:
```console
    $ ./nomad-deploy nomad config
    $ ./cat nomad.yaml
    $ ./nomad-deploy nomad up # deploy cluster
```