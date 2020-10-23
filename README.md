# Missing Container Metrics

Exposes metrics about Docker containers from Docker events.
Every metric contains following labels:

**docker_container_id**
: Full id of the Docker container.

**container_short_id**
: First 6 bytes of the Docker container id.

**container_id**
: Container id represented in the same format as in metrics of k8s pods - prefixed with `docker://`. This enables easy joins in Prometheus to kube_pod_container_info metric.

**name**
: Name of the container.

**image_id**
: Image id represented in the same format as in metrics of k8s pods - prefixed with `docker-pullable://`. This enables easy joins in Prometheus to kube_pod_container_info metric.

## Exposed Metrics

### container_restarts (counter)
Number of restarts of the container. 

### container_ooms (conunter)
Number of OOM kills for the container. This covers OOM kill of any process in the container cgroup.

### container_last_exit_code (gauge)
Last exit code of the container.
