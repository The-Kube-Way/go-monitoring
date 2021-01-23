package http

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// Conf HTTP probe config
type Conf struct {
	CheckInterval        time.Duration     `yaml:"check_interval"`
	URL                  string            `yaml:"url"`
	Method               string            `yaml:"method"`
	ContentType          string            `yaml:"content-type"`
	Body                 string            `yaml:"body"`
	ExpectedStatusCode   int               `yaml:"expected_status_code"`
	StatusCodeErrorAbove int               `yaml:"status_code_error_above"`
	VerifyCertificate    bool              `yaml:"verify_certificate"`
	Headers              map[string]string `yaml:"headers"`
}

// CheckHTTP HTTP probe
func CheckHTTP(config Conf) []string {

	contextLogger := log.WithFields(log.Fields{
		"probe": "http",
		"id":    config.URL})

	var errors []string

	contextLogger.Trace("Entering in CheckHTTP")

	method := "GET"
	if config.Method != "" {
		method = config.Method
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !config.VerifyCertificate},
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest(method, config.URL, strings.NewReader(config.Body))
	if err != nil {
		contextLogger.Error("Fail create request: " + err.Error())
	}
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}
	req.Close = true
	req.Header.Set("User-Agent", "go-monitoring/v1")

	resp, err := client.Do(req)
	if err != nil {
		contextLogger.Error("Fail to do request: " + err.Error())
		errors = append(errors, "Request failed")
		contextLogger.Warning(errors[len(errors)-1])
		return errors
	}

	defer resp.Body.Close()
	contextLogger.Debug(fmt.Sprintf("Status code: %d", resp.StatusCode))

	StatusCodeErrorAbove := 400
	if config.StatusCodeErrorAbove != 0 {
		StatusCodeErrorAbove = config.StatusCodeErrorAbove
	}

	if config.ExpectedStatusCode != 0 {
		if resp.StatusCode != config.ExpectedStatusCode {
			errors = append(
				errors,
				fmt.Sprintf(
					"Status code not as expected (%d != %d)",
					resp.StatusCode,
					config.ExpectedStatusCode))
			contextLogger.Warning(errors[len(errors)-1])
		}
	} else {
		if resp.StatusCode >= StatusCodeErrorAbove {
			errors = append(
				errors,
				fmt.Sprintf(
					"Status code %d >= %d",
					resp.StatusCode,
					StatusCodeErrorAbove))
			contextLogger.Warning(errors[len(errors)-1])
		}
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
				errors := CheckHTTP(config)
				if len(errors) == 0 {
					up.WithLabelValues("http", config.URL).Set(1)
				} else {
					up.WithLabelValues("http", config.URL).Set(0)
				}

			}
		}
	}()
	return ticker
}
