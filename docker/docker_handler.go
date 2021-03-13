package docker

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func HandleDocker(ctx context.Context, slogger *zap.SugaredLogger) error {
	dc, err := client.NewEnvClient()
	if err != nil {
		return errors.Wrap(err, "while creating docker client")
	}

	evts, errs := dc.Events(ctx, types.EventsOptions{})

	containers, err := dc.ContainerList(ctx, types.ContainerListOptions{
		All: true,
	})

	if err != nil {
		return errors.Wrap(err, "while listing containers")
	}

	h := newEventHandler(func(containerID string) (pod string, namespace string) {
		res, err := dc.ContainerInspect(context.Background(), containerID)
		if err != nil {
			return "", ""
		}

		pod = res.Config.Labels["io.kubernetes.pod.name"]
		namespace = res.Config.Labels["io.kubernetes.pod.namespace"]
		return pod, namespace
	})
	for _, c := range containers {
		ci, err := dc.ContainerInspect(ctx, c.ID)
		if err != nil {
			slogger.With("container_id", c.ID, "error", err).Warn("while getting container info")
			continue
		}
		cnt := h.addContainer(c.ID, strings.TrimPrefix(c.Names[0], "/"), c.Image)

		if ci.State.Status == "exited" {
			cnt.die(ci.State.ExitCode)
		}

	}

	for {
		select {
		case e := <-evts:
			err := h.handle(e)
			if err != nil {
				return errors.Wrapf(err, "while handling event %#v", e)
			}
		case err := <-errs:
			return errors.Wrap(err, "while reading events")
		}
	}

}
