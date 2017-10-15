- Simple configuration example

```
docker-compose -f simple-compose.yaml up --build
```

- Full configuration example

```
docker-compose -f full-compose.yaml up --build
```

Now, you can reach each service at

- promviz-front: [http://localhost:8080/graph](http://localhost:8080/)
- promviz: [http://localhost:9091/graph](http://localhost:9091/graph)
- prometheus: [http://localhost:9090/graph](http://localhost:9090/graph)
- mock-metric: [http://localhost:30001/metrics](http://localhost:30001/metrics)