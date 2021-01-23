module main

replace github.com/the-kube-way/go-monitoring/probes/http => ./probes/http

replace github.com/the-kube-way/go-monitoring/probes/ping => ./probes/ping

replace github.com/the-kube-way/go-monitoring/probes/rawtcp => ./probes/rawtcp

go 1.15

require (
	github.com/go-ping/ping v0.0.0-20201115131931-3300c582a663 // indirect
	github.com/prometheus/client_golang v1.9.0
	github.com/sirupsen/logrus v1.7.0
	github.com/the-kube-way/go-monitoring/probes/http v0.0.0-00010101000000-000000000000
	github.com/the-kube-way/go-monitoring/probes/ping v0.0.0-00010101000000-000000000000
	github.com/the-kube-way/go-monitoring/probes/rawtcp v0.0.0-00010101000000-000000000000
	gopkg.in/yaml.v2 v2.4.0
)
