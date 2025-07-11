package rawtcp

import (
	"fmt"
	"net"
	"time"
	"math/rand"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/the-kube-way/go-monitoring/notifications"
	log "github.com/sirupsen/logrus"
)

// Conf RawTCP probe config
type Conf struct {
	CheckInterval time.Duration `yaml:"check_interval"`
	Name          string        `yaml:"name"`
	Host          string        `yaml:"host"`
	Port          string        `yaml:"port"`
	Timeout       time.Duration `yaml:"timeout"`
}

// CheckRawTCP RawTCP probe
func CheckRawTCP(config Conf) []string {

	contextLogger := log.WithFields(log.Fields{
		"probe": "raw_tcp",
		"name":  config.Name,
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
func Schedule(config Conf, interval time.Duration, up *prometheus.GaugeVec, filename string, oncallOffer string) *time.Ticker {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				// Wait between 0 and the interval to spread the load
				waitTime := time.Duration(rand.Int63n(int64(interval)))
				time.Sleep(waitTime)

				errors := CheckRawTCP(config)
				if len(errors) == 0 {
					up.WithLabelValues("raw_tcp", config.Name, net.JoinHostPort(config.Host, config.Port), filename, oncallOffer).Set(1)
				} else {
					up.WithLabelValues("raw_tcp", config.Name, net.JoinHostPort(config.Host, config.Port), filename, oncallOffer).Set(0)

					notifications.SendNotifications(config.Name, net.JoinHostPort(config.Host, config.Port), errors)
				}
			}
		}
	}()
	return ticker
}
