[![Banner](/placeholder.jpg)](https://github.com/xavimondev/supaplay)
<p align="center">
  <img src="https://img.shields.io/github/languages/top/salesof7/eventbus "Language"" alt=" Language" />
  <img src="https://img.shields.io/github/stars/salesof7/eventbus "Stars"" alt=" Stars" />
  <img src="https://img.shields.io/github/issues-pr/salesof7/eventbus "Pull Requests"" alt=" Pull Requests" />
  <img src="https://img.shields.io/github/issues/salesof7/eventbus "Issues"" alt=" Issues" />
  <img src="https://img.shields.io/github/contributors/salesof7/eventbus "Contributors"" alt=" Contributors" />
</p>

## Table of Contents

- [Stack](#stack)
- [Project Summary](#project-summary)
- [Setting Up](#setting-up)
- [Run Locally](#run-locally)
- [FAQ](#faq)

## Stack

- [confluent-kafka-go](https://github.com/confluentinc/confluent-kafka-go): A Kafka client library for Go, essential for server-client communication and data streaming.
- [amqp](https://github.com/streadway/amqp): A Go client for AMQP 0.9.1, crucial for messaging and data fetching.
- [opentelemetry-go](https://github.com/open-telemetry/opentelemetry-go): OpenTelemetry Go SDK for instrumentation and observability, covering authentication, state management, and testing.
- [stdouttrace](https://github.com/open-telemetry/opentelemetry-go/tree/main/exporters/stdout/stdouttrace): A stdout exporter for OpenTelemetry traces, useful for debugging and deployment.
- [otel-sdk](https://github.com/open-telemetry/opentelemetry-go/tree/main/sdk): OpenTelemetry SDK for Go, essential for state management, animations, and styling through observability.
- [otel-trace](https://github.com/open-telemetry/opentelemetry-go/tree/main/trace): OpenTelemetry API for tracing in Go, covering authentication, data fetching, and state management.

## Project Summary

- [internal](/internal): Contains internal packages and modules specific to the project.
- [internal/brokers](/internal/brokers): Manages communication with external systems and services.
- [internal/eventbus](/internal/eventbus): Handles event-driven architecture and messaging within the application.

## Setting Up

Insert your environment variables.

## Run Locally

1. Clone eventbus repository:  
```bash  
git clone https://github.com/salesof7/eventbus  
```
2. Install the dependencies with one of the package managers listed below:  
```bash  
go build -o myapp  
```
3. Start the development mode:  
```bash  
go run main.go  
```

## Contributors

[![Contributors](https://contrib.rocks/image?repo=salesof7/eventbus)](https://github.com/salesof7/eventbus/graphs/contributors)

## FAQ

#### 1.What is this project about?

This project aims to **briefly describe your project's purpose and goals.**

#### 2.How can I contribute to this project?

Yes, we welcome contributions! Please refer to our [Contribution Guidelines](CONTRIBUTING.md) for more information on how to contribute.

#### 3.What is this project about?

Your answer.
