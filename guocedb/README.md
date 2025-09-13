# guocedb

`guocedb` is a modern, MySQL-compatible, distributed relational database built in Go. It is designed for cloud-native environments, offering scalability, reliability, and high performance.

This project is currently under active development and serves as a blueprint implementation based on a detailed architectural design.

## Features

- **MySQL Protocol Compatibility**: Use your existing MySQL clients and tools.
- **Pluggable Storage Engines**: Starts with a high-performance BadgerDB backend.
- **Modern Architecture**: Built with a clean, layered architecture for maintainability and extensibility.
- **Observability**: Exposes Prometheus metrics and health check endpoints.
- **Extensible**: Designed to be extended with new storage engines, features, and distributed capabilities.

## Getting Started

### Prerequisites

- Go 1.21+
- `protoc` compiler and Go gRPC plugins (see `scripts/generate.sh` for details)

### Developer Onboarding: First-Time Setup

Before you can build or run the project, you must generate the Go code from the Protobuf definitions. This is a one-time setup step.

First, make the generation script executable:
```bash
chmod +x scripts/generate.sh
```

Then run the script:
```bash
./scripts/generate.sh
```
This will create the necessary `*.pb.go` files in the `api/` directory.

### Building the Project

You can build the server and CLI binaries using the provided build script. First, make the script executable:

```bash
chmod +x scripts/build.sh
```

Then run the build:
```bash
./scripts/build.sh
```

This will create `guocedb-server` and `guocedb-cli` in the `bin/` directory.

### Running the Server

1.  Copy the example configuration file:
    ```bash
    cp configs/config.yaml.example configs/config.yaml
    ```
2.  Start the server:
    ```bash
    ./bin/guocedb-server --config ./configs/config.yaml
    ```

The server will now be listening on the ports defined in the configuration file (default: MySQL on 3306).

### Running Tests

To run all unit and integration tests, first make the test script executable:

```bash
chmod +x scripts/test.sh
```

Then run the tests:
```bash
./scripts/test.sh
```

## Architecture

For a detailed overview of the project's architecture, please see the [Architecture Design Document](./docs/architecture.md).

## License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for details.
