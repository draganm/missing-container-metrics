package containerd

import (
	"context"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/defaults"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/typeurl"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	// Register grpc event types

	_ "github.com/containerd/containerd/api/events"
)

func HandleContainerd(ctx context.Context, slogger *zap.SugaredLogger) error {
	cl, err := containerd.New(defaults.DefaultAddress)
	if err != nil {
		return errors.Wrap(err, "while creating containerd client")
	}

	nctx := namespaces.WithNamespace(ctx, "k8s.io")

	containerService := cl.ContainerService()

	eh := newEventHandler(slogger, func(containerID string) (containerInfo, error) {
		container, err := containerService.Get(nctx, containerID)
		if err != nil {
			return containerInfo{}, err
		}
		return containerInfo{
			imageID:   container.Image,
			pod:       container.Labels["io.kubernetes.pod.name"],
			namespace: container.Labels["io.kubernetes.pod.namespace"],
			name:      container.Labels["io.kubernetes.container.name"],
		}, nil
	})

	containers, err := containerService.List(nctx)
	if err != nil {
		return errors.Wrap(err, "while listing containers")
	}

	for _, c := range containers {
		eh.getOrCreateContainer(c.ID)
	}

	evts, errs := cl.EventService().Subscribe(nctx)
	for {
		select {
		case err = <-errs:
			return err
		case evt := <-evts:
			if evt.Event != nil {
				v, err := typeurl.UnmarshalAny(evt.Event)
				if err != nil {
					slogger.With("error", err).Error("while unmarshalling containerd event")
				}

				eh.handle(v)
			}
		}
	}

}
