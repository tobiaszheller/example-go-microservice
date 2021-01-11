FROM golang:1.14

WORKDIR /go/src/service-users

# Copy `go.mod` for definitions and `go.sum` to invalidate the next layer
# in case of a change in the dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o service-users

CMD ["./service-users"]
