# example-go-microservice

This repo contains example microservice written in Go using gRPC and MySQL.
It aims to be example of production ready service (beside many TODOs),
which can guide newcommers how to implement microservice in Go.

## Service-users

Service users allows to manage users (create, get, update, list, delete).
It is gRPC service (REST API can be added via gRPC-gateway).

All endpoints can be find in `proto/users.proto`.
It also publish events on users change - defined in `proto/users.proto`
(currently via mock but can be easily swap with real pubsub).

Good introduction into how service works is API `proto/users.proto` and
`integration_tests`.

It was designed as simple CRUD application and domain layer was skipped.

### How to use

Go into directory `service-users`.

Test can be executed using command:
`go test ./...`

In order to build application execute:
`make build`

In order to run application along with db execute:
`make run`
This command use docker-compose to setup mysql container, executes migrations
and starts service.

In order to run integration-tests execute:
`make integration_tests` (make sure that app is running before).

Integration tests are also best way to check how application works.
`Examples` directory contains grpc examples create and update users.
