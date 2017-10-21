Promviz is configured via command-line flags and a configuration file.

### Command-line flags

- `--config.file` Promviz configuration file path. Default is `/etc/promviz/promviz.yaml`.
- `--log.level` The level of logging. Default is `info`.
- `--api.port` Port to listen on for API requests. Default is `9091`.
- `--retrieval.scrape-interval` How frequently to scrape metrics from prometheus servers. Default is `10s`.
- `--retrieval.scrape-timeout` How long until a scrape request times out. Default is `8s`.
- `--cache.size` The maximum number of snapshots can be cached. Default is `100`.
- `--storage.path` Base path of local storage for graph data. Default is `/promviz`.
- `--storage.retention` How long to retain graph data in the storage. Default is `168h`.

### Configuration file

This file contains configuration information for the traffic graph. Promviz reads this file to know where to send prometheus query and how to generate graph data from that query results.

Some valid example files are placed in `example` directory ([simple.yaml](https://github.com/nghialv/promviz/blob/master/example/simple.yaml), [full.yaml](https://github.com/nghialv/promviz/blob/master/example/simple.yaml)).

#### Graph data

Basically, a graph contains a list of nodes and connections. And we have 2 graph levels:
 - global level: contains `cluster` nodes and connections between those nodes
 - cluster level: contains `service` nodes and connections between those nodes.

Promviz will send the specified queries to Prometheus servers and generates nodes and connections.

Let's see how it works.

Suppose we have the following metrics on a prometheus server.

```
code:grpc_server_requests_total:rate2m{client="demo-s1",service="demo-s3",code="OK"} 10
code:grpc_server_requests_total:rate2m{client="demo-s1",service="demo-s3",code="Internal"} 1
code:grpc_server_requests_total:rate2m{client="demo-s1",service="demo-s3",code="NotFound"} 1
code:grpc_server_requests_total:rate2m{client="demo-s1",service="demo-s3",code="InvalidArgument"} 2
code:grpc_server_requests_total:rate2m{client="demo-s2",service="demo-s3",code="OK"} 20
code:grpc_server_requests_total:rate2m{client="demo-s2",service="demo-s3",code="Internal"} 5
```

- example 1: Use the label value as the node name.

```
query: code:grpc_server_requests_total:rate2m
source:
    label: client
target:
    label: service
status:
    label: code
    warningRegex: ^InvalidArgument|FailedPrecondition$
    dangerRegex: Internal
```

3 nodes and 2 connections will be generated

```
nodes:
- "demo-s1"
- "demo-s2"
- "demo-s3"

connections:
- "demo-s1" -> "demo-s3": normal RPS = 11, warning RPS = 2, danger RPS = 1
- "demo-s2" -> "demo-s3": normal RPS = 20, warning RPS = 0, danger RPS = 5
```

- example 2: Want to remove prefix "demo" from node name.

```
query: code:grpc_server_requests_total:rate2m
source:
    label: client
    regex: ^demo-(.+)$
    replacement: $1
target:
    label: service
    regex: ^demo-(.+)$
    replacement: $1
status:
    label: code
    warningRegex: ^InvalidArgument|FailedPrecondition$
    dangerRegex: Internal
```

3 nodes and 2 connections will be generated

```
nodes:
- "s1"
- "s2"
- "s3"

connections:
- "s1" -> "s3": normal RPS = 11, warning RPS = 2, danger RPS = 1
- "s2" -> "s3": normal RPS = 20, warning RPS = 0, danger RPS = 5
```

- example 3: Want to add a connection notice with `severity=error` if the rate of `danger` RPS is greater than or equal to 0.1

```
query: code:grpc_server_requests_total:rate2m
source:
    label: client
target:
    label: service
status:
    label: code
    warningRegex: ^InvalidArgument|FailedPrecondition$
    dangerRegex: Internal
notices:
  - title: HighErrorRate
    statusType: danger
    severityThreshold:
       error: 0.1
```

3 nodes and 2 connections, 1 connection notice will be generated

```
nodes:
- "demo-s1"
- "demo-s2"
- "demo-s3"

connections:
- "demo-s1" -> "demo-s3": normal RPS = 11, warning RPS = 2, danger RPS = 1
- "demo-s2" -> "demo-s3": normal RPS = 20, warning RPS = 0, danger RPS = 5
   - notice: "HighErrorRate", severity = "error"
```

#### Full Template

```
# The name of graph.
graphName: <string>

# This block is used to generate global level of graph.
globalLevel:
  # The maximum volume seen recently to relatively measure particle density.
  maxVolume: <integer>

  # Used to generate cluster nodes and the connections between those nodes.
  clusterConnections:
    - prometheusURL: <string>
      # Query will be sent to prometheus. The result of this query should be a vector.
      query: <string>

      # How to generate source node name from result of query.
      source:
        label: <string>
        regex: <string>
        replacement: <string>
        # Set class name to the generated node.
        class: <string>

      # How to generate target node name from result of query.
      target:
        label: <string>
        regex: <string>
        replacement: <string>
        class: <string>

      # Used to calculate warning RPS and danger RPS of this connection.
      status:
        label: <string>
        warningRegex: <string>
        dangerRegex: <string>

# This block is used to generate cluster level of graph.
clusterLevel:
  - cluster: <string>
    # The maximum volume seen recently to relatively measure particle density.
    maxVolume: <integer>

    # Used to generate service nodes and the connections between those nodes.
    serviceConnections:
      - prometheusURL: <string>
        query: <string>

        # How to generate source node name from result of query.
        source:
          label: <string>
          regex: <string>
          replacement: <string>
          class: <string>

        # How to generate target node name from result of query.
        target:
          label: <string>
          regex: <string>
          replacement: <string>
          class: <string>

        # Used to calculate warning RPS and danger RPS on this connection.
        status:
          label: <string>
          warningRegex: <string>
          dangerRegex: <string>

        # Used to generate <warning|danger> connection notices of this connections.
        notices:
          - title: <string>
            statusType: <string>
            severityThreshold:
              warning: <float>
              error: <float>

    # Used to generate service node notices.
    serviceNotices:
      - title: <string>
        query: <string>
        prometheusURL: <string>
        severityThreshold:
          warning: <float>
          error: <float>
        service:
          label: <string>
          regex: <string>
          replacement: <string>

# <Optional> Customize color for each class.
classes:
  - name: <string>
    color: <string>
```
