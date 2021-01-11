.PHONY: generate integration_tests run

generate:
	protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/*.proto

build: 
	docker-compose build

run:
	docker-compose up

integration_tests:
	go test ./integration_tests -count 1 --tags=integration -v