package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/docker/docker/api/types/events"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type container struct {
	id             string
	name           string
	imageID        string
	pod            string
	namespace      string
	container_name string
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
		"container_name":      c.container_name,
	}
}

func (c *container) create() {
	containerRestarts.GetMetricWith(c.labels())
	containerOOMs.GetMetricWith(c.labels())
	containerLastExitCode.GetMetricWith(c.labels())
}

func (c *container) die(exitCode int) {
	containerLastExitCode.With(c.labels()).Set(float64(exitCode))
}

func (c *container) start() {
	containerRestarts.With(c.labels()).Inc()
}

func (c *container) oom() {
	containerOOMs.With(c.labels()).Inc()
}

func (c *container) destroy() {
	containerRestarts.Delete(c.labels())
	containerOOMs.Delete(c.labels())
	containerLastExitCode.Delete(c.labels())
}

type eventHandler struct {
	containers                  map[string]*container
	mu                          *sync.Mutex
	getContainerInfo func(string) (string, string, string)
}

func newEventHandler(getContainerInfo func(string) (string, string, string)) *eventHandler {
	return &eventHandler{
		containers:                  map[string]*container{},
		mu:                          &sync.Mutex{},
		getContainerInfo: getContainerInfo,
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

	pod, namespace, container_name := eh.getContainerInfo(id)

	c := &container{
		id:        id,
		name:      name,
		imageID:   imageID,
		pod:       pod,
		namespace: namespace,
		container_name: container_name,
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
