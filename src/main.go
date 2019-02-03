package main

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
)

// Should use env vars for the API and APP keys but hard coded for now.
type config struct {
	DDAPIKey string `env:"DD_API_KEY"`
	DDAPPKey string `env:"DD_APP_KEY"`
}

// flag value holders
var configmapns, configmap, namespace, action string

func init() {
	// setup some default logging params
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	// setup flags
	flag.StringVar(&configmapns, "configmapns", "kubeless", "namespace the contains the configmap for downtime tracking")
	flag.StringVar(&configmap, "configmap", "scaling-downtime", "name of the configmap to store and pull downtime ids")
	flag.StringVar(&namespace, "namespace", "sandbox", "namespace the scaledown operation is working on")
	flag.StringVar(&action, "action", "", "action being performed, scaledown = schedule downtime, scaleup = cancel downtime")

	flag.Parse()
}

// setupLoggerContext: define a consistent logging context to be reused
func setupLoggerContext() *log.Entry {
	var err error
	ctxLogger := log.WithFields(log.Fields{
		"namespace":   namespace,
		"configmap":   configmap,
		"configmapns": configmapns,
		"error":       err,
	})
	return ctxLogger
}

func main() {
	contextLogger := setupLoggerContext()
	if action == "scaledown" {
		dtid := scheduleDowntime(namespace)
		err := setDowntimeMapID(configmap, configmapns, configmap, dtid)
		if err != nil {
			contextLogger.Error("Error setting downtime ID from configmap")
		}
	} else if action == "scaleup" {
		dtid, err := getDowntimeMapID(configmap, configmapns, namespace)
		if err != nil {
			contextLogger.Error("Error getting downtime ID from configmap")
		}
		err = cancelDowntime(dtid)
		if err != nil {
			contextLogger.Fatal("Error cancelling downtime")
		}
	} else {
		contextLogger.Fatal("No valid action provide")
	}
}
