
# Promviz [![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat)](http://makeapullrequest.com)

![](https://github.com/nghialv/promviz/blob/master/documentation/sample.png)

## Architecture

![](https://github.com/nghialv/promviz/blob/master/documentation/architecture.png)

## Install

#### Helm chart

If you are using [Helm](https://helm.sh), the simplest way to install is using the charts in `helm` directory with

```
helm install --name promviz ./helm/promviz
helm install --name promviz-front ./helm/promviz-front
```

#### Docker images

Docker images of both `promviz` and `promviz-front` are avaiable on Docker Hub.

- [nghialv2607/promviz](https://hub.docker.com/r/nghialv2607/promviz)
- [mjhd-devlion/promviz-front](https://hub.docker.com/r/mjhd-devlion/promviz-front)

## Example

I have already prepared 2 examples and put it in the `example` directory.

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
- promviz: [http://localhost:9091/graph](http://localhost:9091/graph)
- prometheus: [http://localhost:9090/graph](http://localhost:9090/graph)
- mock-metric: [http://localhost:30001/metrics](http://localhost:30001/metrics)

## Configuration

See [configuration.md](https://github.com/nghialv/promviz/blob/master/documentation/configuration.md) in documentation directory.

## Contributing

Please feel free to create an issue or pull request.

## LICENSE

Promviz is released under the MIT license. See LICENSE file for details.