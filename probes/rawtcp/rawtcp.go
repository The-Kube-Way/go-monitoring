package rawtcp

import (
	"fmt"
	"net"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// Conf RawTCP probe config
type Conf struct {
	CheckInterval time.Duration `yaml:"check_interval"`
	Host          string        `yaml:"host"`
	Port          string        `yaml:"port"`
	Timeout       time.Duration `yaml:"timeout"`
}

// CheckRawTCP RawTCP probe
func CheckRawTCP(config Conf) []string {

	contextLogger := log.WithFields(log.Fields{
		"probe": "ping",
		"id":    net.JoinHostPort(config.Host, config.Port)})

	var errors []string

	contextLogger.Trace("Entering in CheckRawTCP")

	conn, err := net.DialTimeout(
		"tcp",
		net.JoinHostPort(config.Host, config.Port),
		config.Timeout)

	if err != nil {
		errors = append(
			errors,
			fmt.Sprintf("Fail to connect: %s", err.Error()))
		contextLogger.Warning(errors[len(errors)-1])
	}

	if conn != nil {
		defer conn.Close()
		contextLogger.Debug(fmt.Sprintf("TCP port %s is open", config.Port))
	}

	contextLogger.Debug("errors: ", errors)

	return errors
}

// Schedule a probe
func Schedule(config Conf, interval time.Duration, up *prometheus.GaugeVec) *time.Ticker {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				errors := CheckRawTCP(config)
				if len(errors) == 0 {
					up.WithLabelValues("raw_tcp", net.JoinHostPort(config.Host, config.Port)).Set(1)
				} else {
					up.WithLabelValues("raw_tcp", net.JoinHostPort(config.Host, config.Port)).Set(0)
				}
			}
		}
	}()
	return ticker
}
