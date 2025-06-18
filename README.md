# Go Monitoring

![CI](https://github.com/the-kube-way/go-monitoring/workflows/CI/badge.svg?branch=main)
[![Project Status: Active – The project has reached a stable, usable state and is being actively developed.](https://www.repostatus.org/badges/latest/active.svg)](https://www.repostatus.org/#active)

This tool helps to monitor network services (at the moment HTTP, Ping and raw TCP)  
It exposes Prometheus timeseries that can be used by Alertmanager for notification.
This is inspired by [Blackbox Exporter](https://github.com/prometheus/blackbox_exporter).

## Usage

Docker:
```bash
docker run -p 8080:8080 -v $PWD/config.yaml:/config/config.yaml ghcr.io/the-kube-way/go-monitoring:latest
```

Kubernetes:
Check [deploy/kubernetes/manifests.yaml](deploy/kubernetes/manifests.yaml).


## Exported Prometheus timeseries

In addition to [standard Go metrics](https://github.com/prometheus/client_golang), Go monitoring exports the following timeseries:

- **go_monitoring_up**: 1 if target is up, else 0
  - probe: name of the probe (http, ping, raw_tcp)
  - id: id of the target (url for http probe, host for ping, host:port for raw tcp)
  - name: name of the target (as specified in the config, "" if not specified)

## Configuration

The configuration is done via a YAML file:
```yaml
global:
  check_interval: 5m  # Default value if not overwritten per target

http:
  - url: https://example.com
    name: monitor_example_com
    expected_status_code: 200
  
  - url: https://example.com/search
    name: monitor_example_com_search
    check_interval: 30s
    method: POST
    body: '{"test": "test"}'
    skip_verify: false
    headers:
      Authorization: Basic xxx
      Content-Type: application/json
    expected_status_code: 200
    expected_response_body: "OK"

ping:
  - host: example.com
    name: example_server
    check_interval: 1m

raw_tcp:
  - host: ssh.example.com
    port: 22
    name: ssh_example_com
    check_interval: 10m
```

## License

[MIT](LICENSE)
