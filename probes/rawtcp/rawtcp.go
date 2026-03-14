package rawtcp

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
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

func getProbeName(config Conf) string {
	id := net.JoinHostPort(config.Host, config.Port)
	name := strings.TrimSpace(config.Name)
	if name == "" {
		return id
	}

	return name
}

// CheckRawTCP RawTCP probe
func CheckRawTCP(config Conf, filename string, customer string, environment string) []string {
	probeName := getProbeName(config)

	contextLogger := log.WithFields(log.Fields{
		"probe":        "raw_tcp",
		"name":         probeName,
		"id":           net.JoinHostPort(config.Host, config.Port),
		"filename":     filename,
		"customer":     customer,
		"environment":  environment})

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
func Schedule(config Conf, interval time.Duration, up *prometheus.GaugeVec, filename string, customer string, environment string, oncallOffer string) *time.Ticker {
	probeName := getProbeName(config)
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				// Wait between 0 and the interval to spread the load
				waitTime := time.Duration(rand.Int63n(int64(interval)))
				time.Sleep(waitTime)

				errors := CheckRawTCP(config, filename, customer, environment)
				if len(errors) == 0 {
					up.WithLabelValues("raw_tcp", probeName, net.JoinHostPort(config.Host, config.Port), filename, customer, environment, oncallOffer).Set(1)
				} else {
					up.WithLabelValues("raw_tcp", probeName, net.JoinHostPort(config.Host, config.Port), filename, customer, environment, oncallOffer).Set(0)
				}
			}
		}
	}()
	return ticker
}
