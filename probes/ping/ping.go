package ping

import (
	"time"

	"github.com/go-ping/ping"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// Conf Config for ping probe
type Conf struct {
	CheckInterval time.Duration `yaml:"check_interval"`
	Host          string        `yaml:"host"`
	Timeout       time.Duration `yaml:"timeout"`
	RetryCount    int           `yaml:"retry_count"`
	RetryAfter    time.Duration `yaml:"retry_after"`
}

// CheckPing Ping probe
func CheckPing(config Conf) []string {

	contextLogger := log.WithFields(log.Fields{
		"probe": "ping",
		"id":    config.Host})

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
func Schedule(config Conf, interval time.Duration, up *prometheus.GaugeVec) *time.Ticker {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				errors := CheckPing(config)
				if len(errors) == 0 {
					up.WithLabelValues("ping", config.Host).Set(1)
				} else {
					up.WithLabelValues("ping", config.Host).Set(0)
				}

			}
		}
	}()
	return ticker
}
