# Vistio [![CircleCI Build Status](https://circleci.com/gh/nmnellis/vistio.svg?style=shield)](https://circleci.com/gh/nmnellis/vistio) [![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat)](http://makeapullrequest.com) [![MIT Licensed](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/nmnellis/vistio/blob/master/LICENSE)


Vistio is an application that helps you visualize the traffic of your cluster from Prometheus data.

It has 2 components:

- Vistio: retrieves data from Prometheus servers, aggregates them and provides an API to get the graph data.

- [Promviz-front](https://github.com/mjhd-devlion/promviz-front): based on Netflix's [vizceral](https://github.com/Netflix/vizceral) to render traffic graph.

#### Features:
- Generates and renders traffic graph in realtime
- Able to replay from any time in the past
- Able to generate notices on node and connection from prom query
- Provides a sidecar application for k8s that watches config changes and reload Vistio server in runtime
- Fits with [Istio](https://istio.io)'s metrics

![](https://github.com/nmnellis/vistio/blob/master/documentation/sample_filter.png)

## Architecture

![](https://github.com/nmnellis/vistio/blob/master/documentation/architecture.png)

## Install

#### Helm chart

If you are using [Helm](https://helm.sh), the simplest way to install is using the charts in `helm` directory with

```
helm install --name vistio ./helm/vistio
helm install --name promviz-front ./helm/promviz-front
```

#### Docker images

Docker images of both `vistio` and `promviz-front` are available on Docker Hub.

- [nmnellis/promviz](https://hub.docker.com/r/nmnellis/vistio)
- [mjhddevlion/promviz-front](https://hub.docker.com/r/mjhddevlion/promviz-front)

## Example

I have already prepared 2 examples and put them in the `example` directory.

You can try it by going to that directory and run

```
docker-compose -f simple-compose.yaml up --build
```

or

```
docker-compose -f full-compose.yaml up --build
```

Then checkout each service at:
- promviz-front: [http://localhost:8080/graph](http://localhost:8080/)
- vistio: [http://localhost:9091/graph](http://localhost:9091/graph)
- prometheus: [http://localhost:9090/graph](http://localhost:9090/graph)
- mock-metric: [http://localhost:30001/metrics](http://localhost:30001/metrics)

## Configuration

See [configuration.md](https://github.com/nmnellis/vistio/blob/master/documentation/configuration.md) in documentation directory.

## Contributing

Please feel free to create an issue or pull request.

## LICENSE

Vistio is released under the MIT license. See LICENSE file for details.
