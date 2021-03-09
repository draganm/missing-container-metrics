package containerd

import (
	"fmt"
	"sync"
	"time"

	"github.com/containerd/containerd/api/events"

	"github.com/draganm/missing-container-metrics/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type container struct {
	id        string
	name      string
	imageID   string
	pod       string
	namespace string
}

func (c *container) labels() prometheus.Labels {
	return prometheus.Labels{
		"docker_container_id": "not-a-docker-container",
		"container_short_id":  c.id[:12],
		"container_id":        fmt.Sprintf("containerd://%s", c.id),
		"name":                c.name,
		"image_id":            c.imageID,
		"pod":                 c.pod,
		"namespace":           c.namespace,
	}
}

func (c *container) create() {
	if c.id == "" {
		return
	}
	metrics.ContainerRestarts.GetMetricWith(c.labels())
	metrics.ContainerOOMs.GetMetricWith(c.labels())
	metrics.ContainerLastExitCode.GetMetricWith(c.labels())
}

func (c *container) die(exitCode int) {
	if c.id == "" {
		return
	}
	metrics.ContainerLastExitCode.With(c.labels()).Set(float64(exitCode))
}

func (c *container) start() {
	if c.id == "" {
		return
	}
	metrics.ContainerRestarts.With(c.labels()).Inc()
}

func (c *container) oom() {
	if c.id == "" {
		return
	}
	metrics.ContainerOOMs.With(c.labels()).Inc()
}

func (c *container) destroy() {
	metrics.ContainerRestarts.Delete(c.labels())
	metrics.ContainerOOMs.Delete(c.labels())
	metrics.ContainerLastExitCode.Delete(c.labels())
}

type eventHandler struct {
	slogger          *zap.SugaredLogger
	containers       map[string]*container
	mu               *sync.Mutex
	getContainerInfo func(string) (containerInfo, error)
}

type containerInfo struct {
	name      string
	imageID   string
	pod       string
	namespace string
}

func newEventHandler(slogger *zap.SugaredLogger, getContainerInfo func(string) (containerInfo, error)) *eventHandler {
	return &eventHandler{
		slogger:          slogger,
		containers:       map[string]*container{},
		mu:               &sync.Mutex{},
		getContainerInfo: getContainerInfo,
	}
}

func (eh *eventHandler) hasContainer(id string) (*container, bool) {
	c, ex := eh.containers[id]
	return c, ex
}

func (eh *eventHandler) getOrCreateContainer(id string) *container {

	cnt, ex := eh.hasContainer(id)
	if ex {
		return cnt
	}

	ci, err := eh.getContainerInfo(id)
	if err != nil {
		eh.slogger.With("error", err, "container_id", id).Warn("while getting container info")
		return &container{}
	}

	c := &container{
		name:      ci.name,
		id:        id,
		imageID:   ci.imageID,
		pod:       ci.pod,
		namespace: ci.namespace,
	}

	c.create()
	eh.containers[id] = c

	return c

}

func (eh *eventHandler) handle(e interface{}) error {
	eh.mu.Lock()
	defer eh.mu.Unlock()

	switch ev := e.(type) {
	case *events.ContainerCreate:
		id := ev.ID
		eh.getOrCreateContainer(id)
	case *events.TaskOOM:
		id := ev.ContainerID
		c := eh.getOrCreateContainer(id)
		c.oom()
	case *events.ContainerDelete:
		id := ev.ID
		c := eh.getOrCreateContainer(id)
		if c != nil {
			go func() {
				// wait 5 minutes to receive pending
				// events and for scraping by Prometheus
				time.Sleep(5 * time.Minute)
				eh.mu.Lock()
				defer eh.mu.Unlock()
				c.destroy()
				delete(eh.containers, id)
			}()

		}
	case *events.TaskExit:
		id := ev.ContainerID
		c := eh.getOrCreateContainer(id)
		c.die(int(ev.ExitStatus))
	}

	return nil
}
