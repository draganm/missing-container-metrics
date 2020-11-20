# Missing Container Metrics - metrics cadvisor won't give you

[![Docker Pulls](https://img.shields.io/docker/pulls/dmilhdef/missing-container-metrics.svg?maxAge=604800)][hub]
[![Docker Automated Build](https://img.shields.io/docker/automated/dmilhdef/missing-container-metrics)][hub]


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
were set too low, and this particular node was logging a lot more. Logs were
not being forwarded because the Fluentd worker process kept being OOM-kill and
then restarted by the main process. A fix was then deployed 10 minute later.

## Deployment

### Docker

```sh
$ docker run -d -p 3001:3001 -v /var/run/docker.sock:/var/run/docker.sock dmilhdef/missing-container-metrics:v0.14.0
```

### Kubernetes

[> daemon-set.yaml](daemon-set.yaml)
```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: missing-container-metrics
  namespace: kube-system
  labels:
    k8s-app: missing-container-metrics
spec:
  selector:
    matchLabels:
      name: missing-container-metrics
  template:
    metadata:
      labels:
        name: missing-container-metrics
      annotations:
        prometheus.io/scrape: 'true'
        prometheus.io/port: '3001'
    spec:
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      containers:
      - name: missing-container-metrics
        image: dmilhdef/missing-container-metrics:v0.14.0
        resources:
          limits:
            memory: 20Mi
          requests:
            memory: 20Mi
        volumeMounts:
        - name: dockersock
          mountPath: /var/run/docker.sock
      terminationGracePeriodSeconds: 30
      volumes:
      - name: dockersock
        hostPath:
          path: /var/run/docker.sock
```

## Usage

Exposes metrics about Docker containers from Docker events.
Every metric contains following labels:
## Exposed Metrics

Each of those metrics, are published with the labels from the next section.

### `container_restarts` (counter)

Number of restarts of the container. 

### `container_ooms` (conunter)

Number of OOM kills for the container. This covers OOM kill of any process in
the container cgroup.

### `container_last_exit_code` (gauge)

Last exit code of the container.

## Labels

### `docker_container_id`

Full id of the Docker container.

### `container_short_id`

First 6 bytes of the Docker container id.

### `container_id`

Container id represented in the same format as in metrics of k8s pods - prefixed with `docker://`. This enables easy joins in Prometheus to kube_pod_container_info metric.

### `name`

Name of the container.

### `image_id`

Image id represented in the same format as in metrics of k8s pods - prefixed with `docker-pullable://`. This enables easy joins in Prometheus to kube_pod_container_info metric.

## Contributing

Contributions are welcome, send your issues and PRs to this repo.

## License

[MIT](LICENSE) - Copyright Dragan Milic and contributors


[hub]: https://hub.docker.com/r/dmilhdef/missing-container-metrics/
