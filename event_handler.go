package main

import (
	"fmt"
	"strconv"

	"github.com/docker/docker/api/types/events"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type container struct {
	id       string
	restarts int
	ooms     int
	name     string
}

func (c *container) die(exitCode int) {
	c.restarts++
	containerLastExitCode.With(prometheus.Labels{
		"container_id":       c.id,
		"container_short_id": c.id[:12],
		"k8s_container_id":   fmt.Sprintf("docker://%s", c.id),
		"name":               c.name,
	}).Set(float64(exitCode))
}

func (c *container) start() {
	containerRestarts.With(prometheus.Labels{
		"container_id":       c.id,
		"container_short_id": c.id[:12],
		"k8s_container_id":   fmt.Sprintf("docker://%s", c.id),
		"name":               c.name,
	}).Set(float64(c.restarts))
}

func (c *container) oom() {
	c.ooms++
	containerOOMs.With(prometheus.Labels{
		"container_id":       c.id,
		"container_short_id": c.id[:12],
		"k8s_container_id":   fmt.Sprintf("docker://%s", c.id),
		"name":               c.name,
	}).Set(float64(c.ooms))
}

func (c *container) destroy() {
	containerRestarts.Delete(prometheus.Labels{
		"container_id":       c.id,
		"container_short_id": c.id[:12],
		"k8s_container_id":   fmt.Sprintf("docker://%s", c.id),
		"name":               c.name,
	})

	containerOOMs.Delete(prometheus.Labels{
		"container_id":       c.id,
		"container_short_id": c.id[:12],
		"k8s_container_id":   fmt.Sprintf("docker://%s", c.id),
		"name":               c.name,
	})

	containerLastExitCode.Delete(prometheus.Labels{
		"container_id":       c.id,
		"container_short_id": c.id[:12],
		"k8s_container_id":   fmt.Sprintf("docker://%s", c.id),
		"name":               c.name,
	})

}

type eventHandler struct {
	containers map[string]*container
}

func newEventHandler() *eventHandler {
	return &eventHandler{
		containers: map[string]*container{},
	}
}

func (eh *eventHandler) hasContainer(id string) bool {
	_, exists := eh.containers[id]
	return exists
}

func (eh *eventHandler) addContainer(id string, restarts, exitCode int, name string) {
	if eh.hasContainer(id) {
		return
	}
	c := &container{
		id:       id,
		restarts: restarts,
		name:     name,
	}
	eh.containers[id] = c

	shortID := id[:12]
	containerRestarts.With(prometheus.Labels{
		"container_id":       id,
		"container_short_id": shortID,
		"k8s_container_id":   fmt.Sprintf("docker://%s", id),
		"name":               name,
	}).Set(float64(restarts))

	containerLastExitCode.With(prometheus.Labels{
		"container_id":       id,
		"container_short_id": shortID,
		"k8s_container_id":   fmt.Sprintf("docker://%s", id),
		"name":               name,
	}).Set(float64(exitCode))

	containerOOMs.With(prometheus.Labels{
		"container_id":       id,
		"container_short_id": shortID,
		"k8s_container_id":   fmt.Sprintf("docker://%s", id),
		"name":               name,
	}).Set(0)

}

func (eh *eventHandler) handle(e events.Message) error {

	if e.Type != "container" {
		return nil
	}

	c := eh.containers[e.Actor.ID]
	switch e.Action {
	case "create":
		eh.addContainer(e.Actor.ID, 0, -1, e.Actor.Attributes["name"])
	case "destroy":
		if c != nil {
			c.destroy()
			delete(eh.containers, e.Actor.ID)
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
