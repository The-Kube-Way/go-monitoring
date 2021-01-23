package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"

	httpProbe "github.com/the-kube-way/go-monitoring/probes/http"
	pingProbe "github.com/the-kube-way/go-monitoring/probes/ping"
	rawtcpProbe "github.com/the-kube-way/go-monitoring/probes/rawtcp"
	"gopkg.in/yaml.v2"

	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	up = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: "go_monitoring_up"},
		[]string{"probe", "id"})
)

// Conf is global config
type Conf struct {
	Global struct {
		CheckInterval time.Duration `yaml:"check_interval"`
	}
	HTTP   []httpProbe.Conf   `yaml:"http"`
	Ping   []pingProbe.Conf   `yaml:"ping"`
	Rawtcp []rawtcpProbe.Conf `yaml:"raw_tcp"`
}

// ReadConf Read config in YAML
func ReadConf(filename string) (*Conf, error) {

	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	c := &Conf{}
	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return nil, fmt.Errorf("in file %q: %v", filename, err)
	}

	return c, nil
}

// PrettyPrint in JSON of Go objects
func PrettyPrint(obj interface{}) string {
	objJSON, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		log.Error("Fail to represent the object: " + err.Error())
		return "Fail to represent the object"
	}
	return string(objJSON)
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
	return
}

func loadConfig(configPath string) {
	conf, err := ReadConf(configPath)
	if err != nil {
		log.Fatal("Fail to load config" + err.Error())
	}
	log.Debug(PrettyPrint(conf))

	for _, config := range conf.HTTP {

		var CheckInterval time.Duration
		if config.CheckInterval > 0 {
			CheckInterval = config.CheckInterval
		} else {
			CheckInterval = conf.Global.CheckInterval
		}

		log.Info(fmt.Sprintf("Adding HTTP probe for %s every %s", config.URL, CheckInterval))
		log.Trace(fmt.Sprintf("Probe config: %s", PrettyPrint(config)))

		httpProbe.Schedule(
			config,
			CheckInterval,
			up)
	}

	for _, config := range conf.Ping {

		var CheckInterval time.Duration
		if config.CheckInterval > 0 {
			CheckInterval = config.CheckInterval
		} else {
			CheckInterval = conf.Global.CheckInterval
		}

		log.Info(fmt.Sprintf("Adding Ping probe for %s every %s", config.Host, CheckInterval))
		log.Trace(fmt.Sprintf("Probe config: %s", PrettyPrint(config)))

		pingProbe.Schedule(
			config,
			CheckInterval,
			up)
	}

	for _, config := range conf.Rawtcp {

		var CheckInterval time.Duration
		if config.CheckInterval > 0 {
			CheckInterval = config.CheckInterval
		} else {
			CheckInterval = conf.Global.CheckInterval
		}

		log.Info(fmt.Sprintf("Adding Rawtcp probe for %s:%s every %s", config.Host, config.Port, CheckInterval))
		log.Trace(fmt.Sprintf("Probe config: %s", PrettyPrint(config)))

		rawtcpProbe.Schedule(
			config,
			CheckInterval,
			up)
	}
}

func main() {

	var (
		listenAddr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
		configPath = flag.String("config", "/config", "The path to the config directory. All *.yaml will be considered as config files.")
		debugLog   = flag.Bool("debug", false, "Enable debug log level")
		traceLog   = flag.Bool("trace", false, "Enable trace log level")
	)
	flag.Parse()

	log.Println("Version: 0.1-alpha")
	if *traceLog {
		log.SetLevel(log.TraceLevel)
	} else if *debugLog {
		log.SetLevel(log.DebugLevel)
	}

	files, err := filepath.Glob(path.Join(*configPath, "*.yaml"))
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Found config files: ", files)

	if len(files) == 0 {
		log.Fatal("No config file")
	}

	for _, configFile := range files {
		loadConfig(configFile)
	}

	prometheus.MustRegister(up)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/healthz", handleHealthz)
	log.Info("Starting...")
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
