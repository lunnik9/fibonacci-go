# Fibonacci Service

The Fibonacci Service is a gRPC-based application for calculating Fibonacci sequences with two modes:

1. **Simple Fibonacci Sequence**: Returns the first `n` numbers in one response.
2. **Chunked Fibonacci Sequence**: Streams Fibonacci numbers in chunks, optimized for very large sequences.

gRPC was chosen to enable efficient streaming for heavy workloads.

---

## Features
- **Modes**:
    - **Simple Sequence**: Calculates and returns the first `n` numbers.
    - **Chunked Sequence**: Streams results incrementally for large inputs.
- **gRPC APIs**: Efficient performance with real-time streaming.
- **Metrics**: Prometheus integration for monitoring calculation time and frequency.
- **Graceful Shutdown**: Supports soft, hard, and forced shutdown.
- **Dockerized**: Deploy easily with Docker Compose, Grafana, and Prometheus.


---

## Project Structure

```
fibonacci/
├── api/                # gRPC service definition (.proto files)
├── cmd/                # Main application entry point
├── config/             # Configuration logic
├── internal/           # Core application logic
│   ├── domain/         # Domain-specific models and logic
│   ├── genproto/       # Generated protobuf files
│   ├── metrics/        # Prometheus metrics definition
│   ├── mock/           # Mock files for unit testing
│   ├── server/         # gRPC server implementation
│   └── service/        # Business logic implementation
├── monitoring/         # Prometheus and Grafana configuraions and dashboards 
├── .env                # Environment variables for configuration
├── .gitignore          # Git ignored files
├── docker-compose.yml  # Docker Compose file for multi-container setup
├── Dockerfile          # Dockerfile for building the Fibonacci app
├── go.mod              # Go module dependencies
├── Makefile            # Automation scripts
└── README.md           # Project documentation
```

---

## Make

Some Makefile commands require some preinstall. Before developing, testing, ... make sure that rqruired packages installed on your system. You can do it w/ 

```bash
make install-tools
```

---

## Development

### Tests

Can be run w/  

```bash
make test
```

### Mocks

Project utilizes https://github.com/vektra/mockery (primarily because of its simple generic support). Mocks can be generated w/  

```bash
make mocks
```

### Proto

Generate w/  

```bash
make proto
```


both **Proto** and **Mocks** can be generated w/ 

```bash
make generate
```

---

## Running

App requires **docker-compose** to be preinstalled

### Running app w/ metrics

```bash
make up
```

### Running only app 

```bash
make run-app
```

### Shut down

**Graceful shutdown**

App supports graceful shutdown - you need to send **SIGINT** to a running process, for example w/

```bash
kill -SIGINT 1
```

However, the app supports hard shutdown. Difference is - during soft shutdown an app waits all processing endpoints to finish their calculation, whether hard shutdown terminates context and finishes all processing. To execute hard shutdown you need to send second **SIGINT** signal during soft shutdown or **SIGTERM** 

**Compose down**

To stop app w/ all metrics, you can use  

```bash
make down
```


---

## Usage

### gRPC APIs

Install `grpcurl` to test the gRPC APIs:

```bash
# On macOS
brew install grpcurl

# On Debian-based Linux
apt-get install grpcurl
```

#### Query Fibonacci Numbers (Simple Mode):
```bash
grpcurl -plaintext -d '{"n": 10}' localhost:50051 api.FibonacciService/Fibonacci
```

#### Stream Fibonacci Numbers in Chunks (Chunked Mode):
```bash
grpcurl -plaintext -d '{"n": 100, "chunk_size": 10}' localhost:50051 api.FibonacciService/FibonacciStream
```

---

## Monitoring

**Prometheus**: Metrics are available at http://localhost:9090/metrics.

**Grafana**:
- Use http://localhost:3000 to access Grafana.
  - creds **admin:admin**
- Dashboards from **monitoring/dashboards/fibonacci.json** will be preloaded.

<img title="a title" alt="Alt text" src="/dashboard.png">
