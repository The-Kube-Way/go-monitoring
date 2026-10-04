package ping

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	probing "github.com/prometheus-community/pro-bing"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// Conf Config for ping probe
type Conf struct {
	CheckInterval time.Duration `yaml:"check_interval"`
	Host          string        `yaml:"host"`
	Name          string        `yaml:"name"`
	Count         int           `yaml:"count"`
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
// Returns the errors (target considered down) and the warnings (target still up,
// but something needs attention, e.g. partial packet loss)
func CheckPing(config Conf, latency *prometheus.GaugeVec, filename string, customer string, environment string, oncallOffer string) ([]string, []string) {
	probeName := getProbeName(config)

	contextLogger := log.WithFields(log.Fields{
		"probe":       "ping",
		"name":        probeName,
		"id":          config.Host,
		"filename":    filename,
		"customer":    customer,
		"environment": environment})

	var errors []string
	var warnings []string

	contextLogger.Trace("Entering in checkPing")

	count := 3
	if config.Count > 0 {
		count = config.Count
	}
	timeout := 5 * time.Second
	if config.Timeout != 0 {
		timeout = config.Timeout
	}

	for i := 0; i < config.RetryCount+1; i++ {
		pinger, err := probing.NewPinger(config.Host)
		if err != nil {
			contextLogger.Fatal("Fail to setup pinger: " + err.Error())
		}
		pinger.Count = count
		pinger.Timeout = timeout
		err = pinger.Run()
		if err != nil {
			contextLogger.Fatal("Fail to run pinger: " + err.Error())
		}

		stats := pinger.Statistics()

		if stats.PacketLoss < 100 { // At least one packet received
			contextLogger.Debug(fmt.Sprintf("Ping avg RTT: %fs", stats.AvgRtt.Seconds()))
			latency.WithLabelValues("ping", probeName, config.Host, filename, customer, environment, oncallOffer).Set(stats.AvgRtt.Seconds())

			// Some packets were lost: the target is still up, but degraded
			if stats.PacketLoss > 0 {
				warnings = append(warnings, fmt.Sprintf("%g%% packet loss", stats.PacketLoss))
				contextLogger.Warning(warnings[len(warnings)-1])
			}

			// A previous attempt had 100% packet loss, but this one succeeded
			if i > 0 {
				warnings = append(warnings, fmt.Sprintf("Ping succeeded after %d retry(ies)", i))
				contextLogger.Warning(warnings[len(warnings)-1])
			}

			contextLogger.Debug("warnings: ", warnings)

			return errors, warnings
		}

		contextLogger.Warning("100% packet loss")

		time.Sleep(config.RetryAfter)

	}

	// Set latency to 0 to indicate the ping has failed
	latency.WithLabelValues("ping", probeName, config.Host, filename, customer, environment, oncallOffer).Set(0)

	errors = append(errors, "100% packet loss")

	contextLogger.Debug("errors: ", errors)

	return errors, warnings
}

// Schedule a probe
func Schedule(config Conf, interval time.Duration, up *prometheus.GaugeVec, warn *prometheus.GaugeVec, latency *prometheus.GaugeVec, filename string, customer string, environment string, oncallOffer string) *time.Ticker {
	probeName := getProbeName(config)
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				// Wait between 0 and the interval to spread the load
				waitTime := time.Duration(rand.Int63n(int64(interval)))
				time.Sleep(waitTime)

				errors, warnings := CheckPing(config, latency, filename, customer, environment, oncallOffer)
				if len(errors) == 0 {
					up.WithLabelValues("ping", probeName, config.Host, filename, customer, environment, oncallOffer).Set(1)
				} else {
					up.WithLabelValues("ping", probeName, config.Host, filename, customer, environment, oncallOffer).Set(0)
				}

				if len(warnings) == 0 {
					warn.WithLabelValues("ping", probeName, config.Host, filename, customer, environment, oncallOffer).Set(0)
				} else {
					warn.WithLabelValues("ping", probeName, config.Host, filename, customer, environment, oncallOffer).Set(1)
				}
			}
		}
	}()
	return ticker
}
