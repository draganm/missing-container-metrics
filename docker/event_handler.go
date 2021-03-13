package docker

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/docker/docker/api/types/events"
	"github.com/draganm/missing-container-metrics/metrics"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
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
		"docker_container_id": c.id,
		"container_short_id":  c.id[:12],
		"container_id":        fmt.Sprintf("docker://%s", c.id),
		"name":                c.name,
		"image_id":            fmt.Sprintf("docker-pullable://%s", c.imageID),
		"pod":                 c.pod,
		"namespace":           c.namespace,
	}
}

func (c *container) create() {
	metrics.ContainerRestarts.GetMetricWith(c.labels())
	metrics.ContainerOOMs.GetMetricWith(c.labels())
	metrics.ContainerLastExitCode.GetMetricWith(c.labels())
}

func (c *container) die(exitCode int) {
	metrics.ContainerLastExitCode.With(c.labels()).Set(float64(exitCode))
}

func (c *container) start() {
	metrics.ContainerRestarts.With(c.labels()).Inc()
}

func (c *container) oom() {
	metrics.ContainerOOMs.With(c.labels()).Inc()
}

func (c *container) destroy() {
	metrics.ContainerRestarts.Delete(c.labels())
	metrics.ContainerOOMs.Delete(c.labels())
	metrics.ContainerLastExitCode.Delete(c.labels())
}

type eventHandler struct {
	containers                  map[string]*container
	mu                          *sync.Mutex
	getContainerPodAndNamespace func(string) (string, string)
}

func newEventHandler(getContainerPodAndNamespace func(string) (string, string)) *eventHandler {
	return &eventHandler{
		containers:                  map[string]*container{},
		mu:                          &sync.Mutex{},
		getContainerPodAndNamespace: getContainerPodAndNamespace,
	}
}

func (eh *eventHandler) hasContainer(id string) (*container, bool) {
	c, ex := eh.containers[id]
	return c, ex
}

func (eh *eventHandler) addContainer(id, name, imageID string) *container {

	cnt, ex := eh.hasContainer(id)
	if ex {
		return cnt
	}

	pod, namespace := eh.getContainerPodAndNamespace(id)

	c := &container{
		id:        id,
		name:      name,
		imageID:   imageID,
		pod:       pod,
		namespace: namespace,
	}

	c.create()
	eh.containers[id] = c

	return c

}

func (eh *eventHandler) handle(e events.Message) error {
	eh.mu.Lock()
	defer eh.mu.Unlock()

	if e.Type != "container" {
		return nil
	}

	c := eh.addContainer(e.Actor.ID, e.Actor.Attributes["name"], e.Actor.Attributes["image"])
	switch e.Action {
	case "create":
		// just ignore
	case "destroy":
		if c != nil {
			go func() {
				// wait 5 minutes to receive pending
				// events and for scraping by Prometheus
				time.Sleep(5 * time.Minute)
				eh.mu.Lock()
				defer eh.mu.Unlock()
				c.destroy()
				delete(eh.containers, e.Actor.ID)
			}()
		}
	case "die":
		if c != nil {
			exitCodeString := e.Actor.Attributes["exitCode"]
			ec, err := strconv.Atoi(exitCodeString)
			if err != nil {
				return errors.Wrapf(err, "while parsing exit code %q", exitCodeString)
			}
			c.die(ec)
		}
	case "start":
		if c != nil {
			c.start()
		}
	// case "exec_create":
	case "oom":
		if c != nil {
			c.oom()
		}
	}
	return nil
}
