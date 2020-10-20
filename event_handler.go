package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/docker/docker/api/types/events"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type container struct {
	id   string
	name string
}

func (c *container) labels() prometheus.Labels {
	return prometheus.Labels{
		"docker_container_id": c.id,
		"container_short_id":  c.id[:12],
		"container_id":        fmt.Sprintf("docker://%s", c.id),
		"name":                c.name,
	}
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
	go func() {
		// keep metrics alive for a minute so that they can be
		// picked up by Prometheus
		time.Sleep(time.Minute)
		containerRestarts.Delete(c.labels())
		containerOOMs.Delete(c.labels())
		containerLastExitCode.Delete(c.labels())
	}()
}

type eventHandler struct {
	containers map[string]*container
}

func newEventHandler() *eventHandler {
	return &eventHandler{
		containers: map[string]*container{},
	}
}

func (eh *eventHandler) hasContainer(id string) (*container, bool) {
	c, ex := eh.containers[id]
	return c, ex
}

func (eh *eventHandler) addContainer(id string, name string) *container {

	cnt, ex := eh.hasContainer(id)
	if ex {
		return cnt
	}

	c := &container{
		id:   id,
		name: name,
	}
	eh.containers[id] = c

	return c

}

func (eh *eventHandler) handle(e events.Message) error {

	if e.Type != "container" {
		return nil
	}

	c := eh.addContainer(e.Actor.ID, e.Actor.Attributes["name"])
	switch e.Action {
	case "create":
		// just ignore
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
