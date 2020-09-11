package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func main() {
	a := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "bind-address",
				Value: ":3001",
				EnvVars: []string{
					"BIND_ADDRESS",
				},
			},
		},
		Action: func(c *cli.Context) error {
			logger, err := zap.NewProduction()
			if err != nil {
				return err
			}

			slogger := logger.Sugar()

			dc, err := client.NewEnvClient()
			if err != nil {
				return errors.Wrap(err, "while creating docker client")
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			evts, errs := dc.Events(ctx, types.EventsOptions{})

			containers, err := dc.ContainerList(ctx, types.ContainerListOptions{
				All: true,
			})

			if err != nil {
				return errors.Wrap(err, "while listing containers")
			}

			h := newEventHandler()
			for _, c := range containers {
				ci, err := dc.ContainerInspect(ctx, c.ID)
				if err != nil {
					slogger.With("container_id", c.ID, "error", err).Warn("while getting container info")
				}
				exitCode := -1
				if ci.State.Status == "exited" {
					exitCode = ci.State.ExitCode
				}

				h.addContainer(c.ID, ci.RestartCount, exitCode, strings.TrimPrefix(c.Names[0], "/"))
			}

			http.Handle("/metrics", promhttp.Handler())
			a := c.String("bind-address")

			go func() {

				slogger.Infof("Listening on %s", a)
				err := http.ListenAndServe(a, nil)
				if err != nil {
					slogger.With("error", err).Errorf("while listening on %s", a)
				}
				cancel()
			}()
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

		},
	}

	a.RunAndExitOnError()

}
