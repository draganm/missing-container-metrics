package main

import "github.com/prometheus/client_golang/prometheus"

var containerRestarts = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "container_restarts",
		Help: "Number of restarts of a docker container",
	},
	[]string{"container_id", "container_short_id", "k8s_container_id", "name"},
)

var containerOOMs = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "container_ooms",
		Help: "Number of OOM kills of a docker container",
	},
	[]string{"container_id", "container_short_id", "k8s_container_id", "name"},
)

var containerLastExitCode = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "container_last_exit_code",
		Help: "Last exit code of the container",
	},
	[]string{"container_id", "container_short_id", "k8s_container_id", "name"},
)

func init() {
	prometheus.MustRegister(containerRestarts)
	prometheus.MustRegister(containerOOMs)
	prometheus.MustRegister(containerLastExitCode)
}
