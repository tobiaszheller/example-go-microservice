package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"

	pb "github.com/tobiaszheller/example-go-microservice/service-users/proto"
)

func main() {
	addr := flag.String("addr", ":18082", "grpc server addr")
	conn, err := grpc.Dial(*addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewUsersClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.CreateUser(ctx, &pb.CreateUserRequest{
		User: &pb.User{FirstName: "user 1"},
	})
	if err != nil {
		log.Fatalf("Failed: to execute method: %v", err)
	}
	if err := json.NewEncoder(os.Stdout).Encode(r); err != nil {
		log.Fatalf("Failed: to encode resp: %v", err)
	}
}
