package ping

import (
	"math/rand"
	"strings"
	"time"

	"github.com/go-ping/ping"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// Conf Config for ping probe
type Conf struct {
	CheckInterval time.Duration `yaml:"check_interval"`
	Host          string        `yaml:"host"`
	Name          string        `yaml:"name"`
	Timeout       time.Duration `yaml:"timeout"`
	RetryCount    int           `yaml:"retry_count"`
	RetryAfter    time.Duration `yaml:"retry_after"`
}

func getProbeName(config Conf) string {
	id := config.Host
	name := strings.TrimSpace(config.Name)
	if name == "" {
		return id
	}
	return name
}

// CheckPing Ping probe
func CheckPing(config Conf, filename string, customer string, environment string) []string {
	probeName := getProbeName(config)

	contextLogger := log.WithFields(log.Fields{
		"probe":        "ping",
		"name":         probeName,
		"id":           config.Host,
		"filename":     filename,
		"customer":     customer,
		"environment": environment})

	var errors []string

	contextLogger.Trace("Entering in checkPing")

	timeout := 5 * time.Second
	if config.Timeout != 0 {
		timeout = config.Timeout
	}

	for i := 0; i < config.RetryCount+1; i++ {
		pinger, err := ping.NewPinger(config.Host)
		if err != nil {
			contextLogger.Fatal("Fail to setup pinger: " + err.Error())
		}
		pinger.Count = 3
		pinger.Timeout = timeout
		err = pinger.Run()
		if err != nil {
			contextLogger.Fatal("Fail to run pinger: " + err.Error())
		}

		stats := pinger.Statistics()

		if stats.PacketLoss < 100 { // At least one packet received
			return errors
		}

		errors = append(errors, "100% packet loss")
		contextLogger.Warning(errors[len(errors)-1])

		time.Sleep(config.RetryAfter)

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

				errors := CheckPing(config, filename, customer, environment)
				if len(errors) == 0 {
					up.WithLabelValues("ping", probeName, config.Host, filename, customer, environment, oncallOffer).Set(1)
				} else {
					up.WithLabelValues("ping", probeName, config.Host, filename, customer, environment, oncallOffer).Set(0)
				}
			}
		}
	}()
	return ticker
}
