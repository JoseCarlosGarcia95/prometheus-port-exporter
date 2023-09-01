package main

import (
	"log"
	"net/http"
	"os"

	"github.com/JoseCarlosGarcia95/prometheus-port-exporter/collector"
	"github.com/JoseCarlosGarcia95/prometheus-port-exporter/models"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	_, err := models.ReadInstances(os.Args[1])

	if err != nil {
		panic(err)
	}

	collector.StartCollector(os.Args[1])

	log.Printf("Starting metrics-server at port 20000")
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":20000", nil)
}
