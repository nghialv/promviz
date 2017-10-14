
# Promviz [![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat)](http://makeapullrequest.com)

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

Docker images of both `Promviz` and `Promviz-front` are avaiable on [Docker Hub](https://hub.docker.com/r/nghialv2607/promviz).

You can launch `Promviz` and `Promviz-front` with

```
$ docker run --name promviz -d -p 127.0.0.1:9091:9091 nghialv2607/promviz
```

```
$ docker run --name promviz-front -d -p 127.0.0.1:8080:8080 mjhd-devlion/promviz-front
```

`Promviz-front` now can reachable at [http:localhost:8080](http:localhost:8080).

## Example

I have already prepared 2 examples and put it in the `example` directory.

You can try it by going to that directory and run

```
docker-compose up --build
```

Then check your graph at [http:localhost:8080](http:localhost:8080).

## Configuration

## Contributing

Please feel free to create an issue or pull request.

## TODO

- [x] Auto generate notice link to prometheus query
- [x] metrics: storage, retrieval
- [x] cmd/k8s-config-reloader
- [ ] configurable the color of normal class
- [ ] templating notices
- [ ] helm chart
- [ ] cmd/config-validator
- [ ] suport severity name: info, warning, error
- [ ] tests
