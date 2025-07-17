package http

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
	"io/ioutil"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/the-kube-way/go-monitoring/notifications"
	log "github.com/sirupsen/logrus"
)

// Conf HTTP probe config
type Conf struct {
	CheckInterval        time.Duration     `yaml:"check_interval"`
	URL                  string            `yaml:"url"`
	Name                 string            `yaml:"name"`
	Method               string            `yaml:"method"`
	Body                 string            `yaml:"body"`
	ExpectedStatusCode   int               `yaml:"expected_status_code"`
	StatusCodeErrorAbove int               `yaml:"status_code_error_above"`
	ExpectedResponseBody string            `yaml:"expected_response_body"`
	InsecureSkipVerify   bool              `yaml:"skip_verify"`
	Headers              map[string]string `yaml:"headers"`
}

// CheckHTTP HTTP probe
func CheckHTTP(config Conf, latency *prometheus.GaugeVec, filename string, oncallOffer string) []string {

	contextLogger := log.WithFields(log.Fields{
		"probe":    "http",
		"name":     config.Name,
		"id":       config.URL,
		"filename": filename})

	var errors []string

	contextLogger.Trace("Entering in CheckHTTP")

	method := "GET"
	if config.Method != "" {
		method = config.Method
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: config.InsecureSkipVerify},
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest(method, config.URL, strings.NewReader(config.Body))
	if err != nil {
		errors = append(errors, "Fail create request: " + err.Error())
		contextLogger.Warning(errors[len(errors)-1])
		return errors
	}
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}
	req.Close = true
	userAgent := os.Getenv("HTTP_USER_AGENT")
	if userAgent == "" {
		userAgent = "go-monitoring/v1"
	}
	req.Header.Set("User-Agent", userAgent)

	startTime := time.Now()
	resp, err := client.Do(req)
	requestLatency := time.Since(startTime)

	if err != nil {		
		errors = append(errors, "Request failed: " + err.Error())
		contextLogger.Warning(errors[len(errors)-1])
		return errors
	}

	// Save latency
	contextLogger.Debug(fmt.Sprintf("Request latency: %fs", requestLatency.Seconds()))
	latency.WithLabelValues("http", config.Name, config.URL, filename, oncallOffer).Set(requestLatency.Seconds())

	defer resp.Body.Close()
	contextLogger.Debug(fmt.Sprintf("Status code: %d", resp.StatusCode))

	// Check TLS certificate expiration if the connection was secure
	if resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0 {
		cert := resp.TLS.PeerCertificates[0] // Get the leaf certificate
		certExpiresIn := time.Until(cert.NotAfter)
		expirationWarningThreshold := 10 * 24 * time.Hour // 10 days

		contextLogger.Debug(fmt.Sprintf("TLS certificate expires on %s", cert.NotAfter.Format(time.RFC3339)))

		// If certificate has expired, request will fail before
		if certExpiresIn <= expirationWarningThreshold {
			errors = append(
				errors,
				fmt.Sprintf(
					"TLS certificate will expire in %s (on %s)",
					certExpiresIn.Round(time.Hour).String(),
					cert.NotAfter.Format(time.RFC3339)))
			contextLogger.Warning(errors[len(errors)-1])
		}
	}


	// Check status code
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

	// Check response body
	ExpectedResponseBody := ""
	if config.ExpectedResponseBody != "" {
		ExpectedResponseBody = config.ExpectedResponseBody
	}

	if ExpectedResponseBody != "" {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			errors = append(errors, "Fail to read response body: " + err.Error())
			contextLogger.Warning(errors[len(errors)-1])
			return errors
		}

		contextLogger.Debug(fmt.Sprintf("Response body: '%s'", string(body)))
		if string(body) != ExpectedResponseBody {
			errors = append(errors, "Response body not as expected")
			contextLogger.Warning(errors[len(errors)-1])
		}
	}

	contextLogger.Debug("errors: ", errors)

	return errors

}

// Schedule a probe
func Schedule(config Conf, interval time.Duration, up *prometheus.GaugeVec, latency *prometheus.GaugeVec, filename string, oncallOffer string) *time.Ticker {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				// Wait between 0 and the interval to spread the load
				waitTime := time.Duration(rand.Int63n(int64(interval)))
				time.Sleep(waitTime)

				errors := CheckHTTP(config, latency, filename, oncallOffer)

				if len(errors) == 0 {
					up.WithLabelValues("http", config.Name, config.URL, filename, oncallOffer).Set(1)
				} else {
					up.WithLabelValues("http", config.Name, config.URL, filename, oncallOffer).Set(0)

					notifications.SendNotifications(config.Name, config.URL, errors)
				}
			}
		}
	}()
	return ticker
}
