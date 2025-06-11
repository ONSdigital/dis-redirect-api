# dis-redirect-api

A Go API for management of redirects.

## Getting started

* Run `make debug` to run application on <http://localhost:29900>
* Run `make help` to see full list of make targets

### Dependencies

* No further dependencies other than those defined in `go.mod`

#### Validating Specification

To validate the swagger specification you can do this via:

```sh
make validate-specification
```

To run this, you will need to run Node > v20 and have [redocly CLI](https://github.com/Redocly/redocly-cli) installed:

```sh
npm install -g redocly-cli
```

### Configuration

| Environment variable         | Default            | Description                                                                                                          |
| ---------------------------- |--------------------|----------------------------------------------------------------------------------------------------------------------|
| BIND_ADDR                    | :29900             | The host and port to bind to                                                                                         |
| GRACEFUL_SHUTDOWN_TIMEOUT    | 5s                 | The graceful shutdown timeout in seconds (`time.Duration` format)                                                    |
| HEALTHCHECK_INTERVAL         | 30s                | Time between self-healthchecks (`time.Duration` format)                                                              |
| HEALTHCHECK_CRITICAL_TIMEOUT | 90s                | Time to wait until an unhealthy dependent propagates its state to make this app unhealthy (`time.Duration` format)   |
| OTEL_EXPORTER_OTLP_ENDPOINT  | localhost:4317     | Endpoint for OpenTelemetry service                                                                                   |
| OTEL_SERVICE_NAME            | dis-redirect-api   | Label of service for OpenTelemetry service                                                                           |
| OTEL_BATCH_TIMEOUT           | 5s                 | Timeout for OpenTelemetry                                                                                            |
| OTEL_ENABLED                 | false              | Feature flag to enable OpenTelemetry                                                                                 |
 | REDIS_ADDRESS                | localhost:6379     | Endpoint for Redis service                                                                                           |

### SDKs

This API has two SDKs available:

* [Go](./sdk/go/README.md)
* [Java](./sdk/java/README.md)

## Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

## License

Copyright © 2025, Office for National Statistics (<https://www.ons.gov.uk>)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
