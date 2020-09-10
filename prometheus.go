package main

import "github.com/prometheus/client_golang/prometheus"

var containerRestarts = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "container_restarts",
		Help: "Number of restarts of a docker container",
	},
	[]string{"container_id", "container_short_id", "name"},
)

func init() {
	prometheus.MustRegister(containerRestarts)
}
