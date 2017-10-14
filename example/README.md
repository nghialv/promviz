- Simple configuration example

```
docker-compose -f simple-compose.yaml up --build
```

- Full configuration example

```
docker-compose -f full-compose.yaml up --build
```

Now, you can reach each service at

- promviz: [http://localhost:8000/graph](http://localhost:8000/graph)
- prometheus: [http://localhost:9090/graph](http://localhost:9090/graph)
- mock-metric: [http://localhost:30001/metrics](http://localhost:30001/metrics)