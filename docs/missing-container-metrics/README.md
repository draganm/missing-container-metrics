# Missing Container Metrics - metrics cadvisor won't give you

This chart will install a daemon set that will expose container metrics such as `container_ooms` that are not available otherwise.

## Introduction
For motivation and implementation details please refer to [blog post](https://www.netice9.com/blog/guide-to-oomkill-alerting-in-kubernetes-clusters/) introducing the `missing-container-metrics`

## TL;DR;

```bash
$ helm repo add missing-container-metrics https://draganm.github.io/missing-container-metrics
$ helm install missing-container-metrics missing-container-metrics/missing-container-metrics
```

## Adding Helm repo to your Helm client
```bash
$ helm repo add missing-container-metrics https://draganm.github.io/missing-container-metrics
```

## Installing the Chart
```bash
$ kubectl create namespace missing-container-metrics
$ helm install my-release-name missing-container-metrics/missing-container-metrics -n missing-container-metrics
```

## Configuration

| Parameter                                             | Description                                                       | Default                                                           |
|-------------------------------------------------------|-------------------------------------------------------------------|-------------------------------------------------------------------|
| image.repository                                      | missing-container-metrics image name                              | `dmilhdef/missing-container-metrics`                              |
| image.pullPolicy                                      | pull policy for the image                                         | `IfNotPresent`                                                    |
| image.tag                                             | tag of the missing-container-metrics image                        | `v0.21.0`                                                         |
| imagePullSecrets                                      | pull secrets for the image                                        | `[]`                                                              |
| nameOverride                                          | Override the generated chart name. Defaults to .Chart.Name.       |                                                                   |
| fullnameOverride                                      | Override the generated release name. Defaults to .Release.Name.   |                                                                   |
| podAnnotations                                        | Annotations for the started pods                                  | `{"prometheus.io/scrape": "true", "prometheus.io/port": "3001"}`  |
| podSecurityContext                                    | Set the security context for the pods                             |                                                                   |
| securityContext                                       | Set the security context for the container in the pods            |                                                                   |
| resources                                             | CPU/memory resource requests/limits                               | `{}`                                                              |
| useDocker                                             | If true, container info is obtained from Docker                   | `false`                                                           |
| useContainerd                                         | If true, container info is obtained from Containerd               | `true`                                                            |
