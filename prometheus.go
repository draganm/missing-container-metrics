package main

import "github.com/prometheus/client_golang/prometheus"

var containerRestarts = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "container_restarts",
		Help: "Number of restarts of a docker container",
	},
	[]string{"container_id", "container_short_id", "name"},
)

var containerOOMs = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "container_ooms",
		Help: "Number of OOM kills of a docker container",
	},
	[]string{"container_id", "container_short_id", "name"},
)

func init() {
	prometheus.MustRegister(containerRestarts)
	prometheus.MustRegister(containerOOMs)
}
