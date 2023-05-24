Go Audit Server for FinanceApi
==========================

**Stack**:

- GO 1.20.4
- MongoDB 6.0.5
- gRPC
- RabbitMQ 3.11.16

This is a simple audit service consuming proto from gRPC and RabbitMQ

Please check [Proto](proto) for API details

## Before run:

1. Setup you configs for RabbitMQ and MongoDB. You can check example.env

## To run:

```shell
go run cmd/app/main.go
```
