package main

import (
	"errors"
	"strconv"

	log "github.com/sirupsen/logrus"
	datadog "gopkg.in/zorkian/go-datadog-api.v2"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// scheduleDowntime submit an API request to DataDog to create a downtime event scoped to the namespace
func scheduleDowntime(namespace string) int {
	// establish API client with datadog, hardcoded apikey, appkey for now
	client := datadog.NewClient("xxxxxxxxxxxx", "xxxxxxxxxxxxxxxx")
	// make and append to a slice the tags for scope which is the namespace
	var scope []string
	scope = append(scope, "kube_namespace:"+namespace)
	// create a downtime object
	dt := datadog.Downtime{
		Message: datadog.String("triggered downtime"),
		Scope:   scope,
	}
	// create a dowtime from the previous object
	downtime, err := client.CreateDowntime(&dt)
	// error occurred tring to set the downtime
	if err != nil {
		log.WithFields(log.Fields{
			"namespace": namespace,
		}).Error("Failed to create downtime")
	}
	// return the id of the recently created downtime
	return downtime.GetId()
}

// getDowntimeMapID get the downtime id from the configmap for the provided namespace that is being scaled
func getDowntimeMapID(configmap, configmapns, namespace string) (int, error) {
	_, c, err := getDowntimeMap(configmap, configmapns)
	dts := c.Data[namespace]
	if dts == "" {
		log.WithFields(log.Fields{
			"namespace":   namespace,
			"configmap":   configmap,
			"configmapns": configmapns,
		}).Error("Failed to get namespace in defined configmap")
		return 0, errors.New("Failed to get namespace")
	}
	dtid, err := strconv.Atoi(dts)
	if err != nil {
		log.WithFields(log.Fields{
			"namespace":   namespace,
			"configmap":   configmap,
			"configmapns": configmapns,
			"id":          dtid,
		}).Error("Failed to convert value to int")
		return 0, errors.New("Failed to convert value to int")
	}
	return dtid, nil
}

// setDowntimeMapID set the namespace key in the configmap to the downtime ID provided
func setDowntimeMapID(configmap, configmapns, namespace string, downtimeid int) error {
	cs, c, err := getDowntimeMap(configmap, configmapns)
	c.Data[namespace] = strconv.Itoa(downtimeid)
	c, err = cs.CoreV1().ConfigMaps(configmapns).Update(c)
	if err != nil {
		log.WithFields(log.Fields{
			"namespace": namespace,
			"id":        downtimeid,
			"configmap": configmap,
		}).Error("Failed to update configmap with downtime id")
	}
	return nil
}

// getDowntimeMap retrieve the given configmap from the namespace
// return clientset and configmap for reuse
func getDowntimeMap(configmap, namespace string) (*kubernetes.Clientset, *v1.ConfigMap, error) {
	// Generate the configuration for connecting to cluster based on running in cluster
	config, err := rest.InClusterConfig()
	if err != nil {
		log.WithFields(log.Fields{
			"error":     err,
			"configmap": configmap,
			"namespace": namespace,
		}).Error("Unable to get InCluster config")
		return nil, nil, errors.New("Unable to get InCluster config")
	}
	// Establish a client using the generated configuration
	clientset, err := kubernetes.NewForConfig(config)
	// Get the configmap from the namespace
	c, err := clientset.CoreV1().ConfigMaps(namespace).Get(configmap, metav1.GetOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"error":     err,
			"configmap": configmap,
			"namespace": namespace,
		}).Error("Unable to get configmap")
	}
	return clientset, c, nil
}

// Only way to cancel a downtime is to delete it
func cancelDowntime(d int) error {
	client := datadog.NewClient("xxxxxxxxxxxxxx", "xxxxxxxxxxxxxxxx")
	// Make sure we have a downtime for the id
	_, err := client.GetDowntime(d)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"id":    d,
		}).Error("Unable to get the specified downtime id")
		return errors.New("Unable to get the specified downtime id")
	}
	// Remove the downtime
	client.DeleteDowntime(d)
	return nil
}
