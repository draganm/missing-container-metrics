package main

import (
	"strconv"

	"github.com/docker/docker/api/types/events"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type container struct {
	id           string
	restarts     int
	lastExitCode int
	name         string
}

func (c *container) die(exitCode int) {
	c.lastExitCode = exitCode
	c.restarts++
}

func (c *container) start() {
	containerRestarts.With(prometheus.Labels{
		"container_id":       c.id,
		"container_short_id": c.id[:12],
		"name":               c.name,
	}).Set(float64(c.restarts))
}

func (c *container) destroy() {
	containerRestarts.Delete(prometheus.Labels{
		"container_id":       c.id,
		"container_short_id": c.id[:12],
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

// 2020-09-11T00:09:37.387259418+02:00 container oom c1583ef58f45cc0d82172cae63ddbd44699ad330175d12afd283467f13a843f3 (image=ubuntu, name=elated_keldysh)
// 2020-09-11T00:09:44.678570439+02:00 container oom c1583ef58f45cc0d82172cae63ddbd44699ad330175d12afd283467f13a843f3 (image=ubuntu, name=elated_keldysh)
// 2020-09-11T00:09:44.808568793+02:00 container die c1583ef58f45cc0d82172cae63ddbd44699ad330175d12afd283467f13a843f3 (exitCode=1, image=ubuntu, name=elated_keldysh)

func (eh *eventHandler) hasContainer(id string) bool {
	_, exists := eh.containers[id]
	return exists
}

func (eh *eventHandler) addContainer(id string, restarts int, name string) {
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
		"name":               name,
	}).Set(float64(restarts))
}

func (eh *eventHandler) handle(e events.Message) error {

	if e.Type != "container" {
		return nil
	}

	c := eh.containers[e.Actor.ID]
	switch e.Action {
	case "create":
		eh.addContainer(e.Actor.ID, 0, e.Actor.Attributes["name"])
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
	}
	return nil
}
