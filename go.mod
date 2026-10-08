module main

replace github.com/the-kube-way/go-monitoring/probes/http => ./probes/http

replace github.com/the-kube-way/go-monitoring/probes/ping => ./probes/ping

replace github.com/the-kube-way/go-monitoring/probes/rawtcp => ./probes/rawtcp

go 1.26.0

require (
	github.com/prometheus/client_golang v1.25.0
	github.com/sirupsen/logrus v1.9.3
	github.com/the-kube-way/go-monitoring/probes/http v0.0.0-20250406112039-67c78299763a
	github.com/the-kube-way/go-monitoring/probes/ping v0.0.0-20250406112039-67c78299763a
	github.com/the-kube-way/go-monitoring/probes/rawtcp v0.0.0-20250406112039-67c78299763a
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus-community/pro-bing v0.9.1 // indirect
	github.com/prometheus/client_model v0.6.3 // indirect
	github.com/prometheus/common v0.72.0 // indirect
	github.com/prometheus/procfs v0.22.0 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	golang.org/x/net v0.59.0 // indirect
	golang.org/x/sync v0.23.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.42.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)
