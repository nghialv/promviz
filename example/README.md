- Start prometheus

```
docker-compose -f simple-compose.yaml
```

- Start promviz

```
go run ../cmd/promviz/main.go --config.file simple-promviz.yaml --api.listen-address ":8000"
```