package metrics

import "github.com/prometheus/client_golang/prometheus"

var labelNames = []string{
	"container_id",
	"container_short_id",
	"docker_container_id",
	"name",
	"image_id",
	"pod",
	"namespace",
}

var ContainerRestarts = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "container_restarts",
		Help: "Number of restarts of a docker container",
	},
	labelNames,
)

var ContainerOOMs = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "container_ooms",
		Help: "Number of OOM kills of a docker container",
	},
	labelNames,
)

var ContainerLastExitCode = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "container_last_exit_code",
		Help: "Last exit code of the container",
	},
	labelNames,
)

func init() {
	prometheus.MustRegister(ContainerRestarts)
	prometheus.MustRegister(ContainerOOMs)
	prometheus.MustRegister(ContainerLastExitCode)
}
