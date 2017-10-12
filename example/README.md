### Start prometheus-mock

```
docker-compose up
```

### Start promviz

- with simple configuration

```
go run ../cmd/promviz/main.go --config.file simple.yaml --api.listen-address ":8000" --storage.path ~/Downloads/db
```

- with full configuration

```
go run ../cmd/promviz/main.go --config.file full.yaml --api.listen-address ":8000" --storage.path ~/Downloads/db
```
