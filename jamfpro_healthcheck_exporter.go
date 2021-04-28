package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// healthcheckEndpoint defines the expected response json structure found at jamf healthCheck.html.
type healthcheckEndpoint struct {
	HealthCode  int    `json:"healthCode"`
	HTTPCode    int    `json:"httpCode"`
	Description string `json:"description"`
}

var (
	address       = flag.String("listen.address", "0.0.0.0", "assign an IP address for this service to listen on")
	baseURL       string
	debug         = flag.Bool("debug", false, "enable the service debug output")
	disCol        = true
	enableColLogs = flag.Bool("enable-color-logs", false, "when developing in debug mode, prettier to set this for visual colors")
	help          = flag.Bool("help", false, "help will display this helpful dialog output")
	jamfProto     = flag.String("jamf.proto", "https", "set the protocol used to interact with the application")
	jamfURL       = flag.String("jamf.url", "", "REQUIRED: provide the application url to be monitored (ie. jamf.domain.com)")
	port          = flag.String("listen.port", "9613", "set the port that this service will listen on")
	usageMessage  = "The JamfPro Health Check Exporter is used to reach out and collect the info from\n" +
		"the healthCheck.html page, then turn that into a collectable metric.\n" +
		"\nMetrics Example:\n" +
		"jamfpro_healthcheck_healthcode{description='',description_full='',healthcode='',httpcode='',jamf_url=''} 0\n" +
		"jamfpro_healthcheck_scrape_url_up{httpcode='',jamf_url=''} 0\n" +
		"\nReference:\n" +
		"https://docs.jamf.com/10.18.0/jamf-pro/administrator-guide/Jamf_Pro_Health_Check_Page.html\n" +
		"\nUsage: jamfpro_healthcheck_exporter [Arguments...]\n" +
		"\nArguments:\n"
)

// usage is used to display this binaries usage description and then exit the program.
var usage = func() {
	fmt.Println(usageMessage)
	flag.PrintDefaults()
	os.Exit(0)
}

// healthcheckCollector is the structure of our prometheus collector containing it descriptors.
type healthcheckCollector struct {
	healthcodeMetric *prometheus.Desc
	scrapeUpMetric   *prometheus.Desc
}

// newHealthcheckCollector is the constructor for our collector used to initialize the metrics.
func newHealthcheckCollector() *healthcheckCollector {
	return &healthcheckCollector{
		healthcodeMetric: prometheus.NewDesc(
			"jamfpro_healthcheck_healthcode",
			"metric returns the healthcode status of the monitored jamf pro healthCheck.html endpoint",
			[]string{
				"healthcode",
				"httpcode",
				"description",
				"description_full",
				"jamf_url",
			},
			nil,
		),
		scrapeUpMetric: prometheus.NewDesc(
			"jamfpro_healthcheck_scrape_url_up",
			"metric shows the status of the connection to the jamf healthCheck.html endpoint",
			[]string{
				"httpcode",
				"jamf_url",
			},
			nil,
		),
	}
}

// Describe is required by prometheus to add out metrics to the default prometheus desc channel
func (collector *healthcheckCollector) Describe(ch chan<- *prometheus.Desc) {

	// Update this section with each metric you create for a given collector
	ch <- collector.healthcodeMetric
	ch <- collector.scrapeUpMetric
}

// Collect implements required collect function for all prometheus collectors
func (collector *healthcheckCollector) Collect(ch chan<- prometheus.Metric) {

	log.Debug("get url", baseURL)
	resp, err := http.Get(baseURL)
	if err != nil {
		log.Warn("http.Get base URL returned an error:", err)
		ch <- prometheus.MustNewConstMetric(collector.scrapeUpMetric, prometheus.GaugeValue, 0, "", *jamfURL)
		return
	}
	defer resp.Body.Close()

	log.Debug("set scrape up metric up with statuscode:", strconv.Itoa(resp.StatusCode))
	ch <- prometheus.MustNewConstMetric(collector.scrapeUpMetric, prometheus.GaugeValue, 1, strconv.Itoa(resp.StatusCode), *jamfURL)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("ioutil.ReadAll returned an error:", err)
	}

	// remove the trailing \n and any whitespace before checking if we got an empty body
	if strings.TrimSuffix(strings.Replace(string(body), " ", "", -1), "\n") == "[]" {
		log.Debug(*jamfURL, " health code: 0")
		ch <- prometheus.MustNewConstMetric(collector.healthcodeMetric, prometheus.GaugeValue, 0, "", "", "", descLong(0), *jamfURL)
		return
	}

	m := healthcheck(body)
	log.Debug("the returned body map:", m)

	// range over the map to create each metric with it's labels.
	for _, metric := range m {
		log.Debug("create healthcode metric for", metric.Description)
		ch <- prometheus.MustNewConstMetric(
			collector.healthcodeMetric,
			prometheus.GaugeValue,
			float64(metric.HealthCode),
			strconv.Itoa(metric.HealthCode),
			strconv.Itoa(metric.HTTPCode),
			metric.Description,
			descLong(metric.HealthCode),
			*jamfURL,
		)
	}

}

// healthcheck takes a http body btye slice and unmarshals it into the healthCheck.html structure.
func healthcheck(body []byte) []healthcheckEndpoint {

	log.Debug("create the json map to unmarshal the json body into")
	var m []healthcheckEndpoint

	log.Debug("unmarshal (turn unicode back into a string) request body into map structure")
	err := json.Unmarshal(body, &m)
	if err != nil {
		log.Error("error Unmarshalling: ", err)
		log.Info("Problem unmarshalling the following string: ", string(body))
	}

	return m
}

// descLong takes in the health code int and returns the long description that matches.
func descLong(healthcode int) string {
	switch healthcode {
	case 1:
		return "An error occurred while testing the database connection."
	case 2:
		return "The Jamf Pro Setup Assistant was detected."
	case 3:
		return "A configuration error occurred while attempting to connect to the database."
	case 4:
		return "The Jamf Pro web app is initializing."
	case 5:
		return "An instance of the Jamf Pro web app in a clustered environment failed to start."
	case 6:
		return "A fatal error occurred and prevented the Jamf Pro web app from starting."
	default:
		return "The Jamf Pro web app is running without error."
	}
}

func main() {
	flag.Parse()

	// check if help has been passed
	if *help {
		usage()
	}

	// check for required arguments
	if *jamfURL == "" {
		fmt.Printf("-jamf.url must be provided\n\n")
		usage()
	}

	// adjust the logrus logger. Disable colors by default (adjustable with enable-color-logs argument). Enable full time-stamps by default
	if *enableColLogs {
		disCol = false
	}
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		DisableColors: disCol,
	})

	// check for debug argument, adjust if set
	if *debug {
		log.SetLevel(log.DebugLevel)
		log.Debug("Log Level: debug")
	}

	baseURL = *jamfProto + "://" + *jamfURL + "/healthCheck.html"
	log.Debug("request url: ", baseURL)

	// Create a new instance of the healthcheckCollector and then
	// register it with the prometheus client.
	exporter := newHealthcheckCollector()
	prometheus.MustRegister(exporter)

	srv := http.Server{
		Addr: *address + ":" + *port,
	}

	// This will run metrics endpoint by the prometheus http handler.
	http.Handle("/metrics", promhttp.Handler())

	// make a channel to wait for os signals
	ch := make(chan os.Signal, 1)
	// define what signals we are going to wait for
	signal.Notify(ch, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	// start the http server in a go routine
	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			log.Fatal("ListenAndServe Error:", err)
		}
	}()

	log.Info("serving jamfpro_healthcheck_exporter on ", *address+":"+*port)

	// block waiting for channel. serivce will stay running waiting for a defined signal, once the signal comes, it will continue.
	s := <-ch
	log.Info("Got SIGNAL:", s)

	log.Debug("close channel")
	close(ch)

	log.Info("shutdown http server")
	err := srv.Shutdown(context.Background())
	if err != nil {
		// Error from closing listeners, or context timeout
		log.Fatal("Shutdown error:", err)
	}

	log.Info("Supposed Graceful Shutdown")

}
