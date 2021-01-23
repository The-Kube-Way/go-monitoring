# Go Monitoring

![Go](https://github.com/the-kube-way/go-monitoring/workflows/Go/badge.svg?branch=main)
[![Project Status: WIP  Initial development is in progress, but there has not yet been a stable, usable release suitable for the public.](https://www.repostatus.org/badges/latest/wip.svg)](https://www.repostatus.org/#wip)

This tool helps to monitor network services (at the moment HTTP, Ping and raw TCP)  
It exposes Prometheus timeseries to be used by Alertmanager for notification.

## Usage

Have a look to [kubernetes-manifests.yaml](kubernetes-manifests.yaml) and [example.yaml](example.yaml) 

## Exported Prometheus timeseries

In addition to [standard Go metrics](https://github.com/prometheus/client_golang), Go monitoring exports the following timeseries:

- **go_monitoring_up**: 1 if target is up, else 0
  - probe: name of the probe (http, ping, raw_tcp)
  - id: id of the target (url for http probe, host for ping, host:port for raw tcp)
