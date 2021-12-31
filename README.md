# Missing Container Metrics - metrics cadvisor won't give you

Fork of [draganm/missing-container-metrics](https://github.com/draganm/missing-container-metrics).

**STATUS: stable, maintained**

cadvisor is great, but missing a few important metrics, that every serious devops person wants to know about.
This is a secondary process to export all the missing [Prometheus](https://prometheus.io) metrics:

* OOM-kill
* number of container restarts
* last exit code

This was motivated by hunting down a OOM kills in a large Kubernetes cluster.
It's possible for containers to keep running, even after a OOM-kill, if a
sub-process got affect for example. Without this metric, it becomes much more
difficult to find the root cause of the issue.

True story; after this was deployed, a recurring OOM-kill in Fluentd was
quickly discovered on one of the nodes. It turns out that the resource limits
were set too low to process logs on that node. Logs were
not being forwarded because the Fluentd worker process kept being OOM-kill and
then restarted by the main process. A fix was then deployed 10 minutes later.

## Supported Container Runtimes
* Docker
* Containerd

Kubernetes 1.20 has deprecated Docker container runtime, so we have added support for Containerd since the version `0.21.0` of `missing-container-metrics`.
Both options should cover most of common use cases (EKS, GKE, K3S, Digital Ocean Kubernetes, ...).

## Deployment

### Kubernetes

The easiest way of installing `missing-container-metrics` in your kubernetes cluster is using the [Helm Chart](/charts/missing-container-metrics).

### Docker

```sh
$ docker run -d -p 3001:3001 -v /var/run/docker.sock:/var/run/docker.sock ghcr.io/cablespaghetti/missing-container-metrics:0.22.0
```

## Usage

Exposes metrics about Docker/Containerd containers.
Every metric contains following labels:
## Exposed Metrics

Each of those metrics, are published with the labels from the next section.

### `container_restarts` (counter)

Number of restarts of the container. 

### `container_ooms` (counter)

Number of OOM kills for the container. This covers OOM kill of any process in the container cgroup.

### `container_last_exit_code` (gauge)

Last exit code of the container.

## Labels

### `docker_container_id`

Full id of the container.

### `container_short_id`

First 6 bytes of the Docker container id.

### `container_id`

Container id represented in the same format as in metrics of kubernetes pods - prefixed with `docker://` and `containerd://` depending on the container runtime. This enables easy joins in Prometheus to `kube_pod_container_info` metric.

### `name`

Name of the container.

### `image_id`

Image id represented in the same format as in metrics of k8s pod. This enables easy joins in Prometheus to `kube_pod_container_info` metric.

### `pod`

If `io.kubernetes.pod.name` label is set on the container, it's value
will be set as the `pod` label in the metric

### `namespace`

If `io.kubernetes.pod.namespace` label is set on the container, it's value
will be set as the `namespace` label of the metric.

Together with `pod`, this label is useful in the context of Kubernetes deployments, to determine namespace/pod to which the container is part of.
One can see it as a shortcut to joining with the `kube_pod_container_info` metric to determine those values.


## Contributing

Contributions are welcome, send your issues and PRs to this repo.

## License

[MIT](LICENSE) - Copyright Dragan Milic and Sam Weston

