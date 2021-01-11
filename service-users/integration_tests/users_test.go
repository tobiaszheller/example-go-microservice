// +build integration

package integration_tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/tobiaszheller/example-go-microservice/service-users/proto"
)

func TestUsers(t *testing.T) {
	cli, conn := mustSetupClient(t)
	defer conn.Close()
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	firstUser := &pb.User{
		FirstName: "Johnny",
		LastName:  "Cash",
		Country:   "US",
		Nickname:  "Johnny Cash",
		Email:     fmt.Sprintf("%s@test.com", uuid.New().String()),
	}

	var firstUserId string
	t.Run("must create new user", func(t *testing.T) {
		got, err := cli.CreateUser(ctx, &pb.CreateUserRequest{User: firstUser})
		assertNoErr(t, err)
		assertUserEqual(t, firstUser, got, cmpopts.IgnoreFields(pb.User{}, "Id", "UpdatedAt"))
		firstUserId = got.Id

		// Make sure you cannot create user with the same email twice.
		got, err = cli.CreateUser(ctx, &pb.CreateUserRequest{User: firstUser})
		assertErr(t, err, codes.AlreadyExists)
	})
	t.Run("must get user", func(t *testing.T) {
		got, err := cli.GetUser(ctx, &pb.GetUserRequest{Id: firstUserId})
		assertNoErr(t, err)
		assertUserEqual(t, firstUser, got, cmpopts.IgnoreFields(pb.User{}, "Id", "UpdatedAt"))

		// Make sure you cannot get user with invalid id.
		got, err = cli.GetUser(ctx, &pb.GetUserRequest{Id: "invalid-id"})
		assertErr(t, err, codes.NotFound)
	})
	t.Run("must update user", func(t *testing.T) {
		updateReq := &pb.User{
			Id:        firstUserId,
			FirstName: "John",
			LastName:  "Cache",
			Nickname:  "john_cache",
			Country:   "US",
			Email:     fmt.Sprintf("%s@test.com", uuid.New().String()),
		}
		got, err := cli.UpdateUser(ctx, &pb.UpdateUserRequest{User: updateReq})
		assertNoErr(t, err)
		assertUserEqual(t, updateReq, got, cmpopts.IgnoreFields(pb.User{}, "Id", "UpdatedAt"))

		// Make sure that also after get we receive updated user.
		got, err = cli.GetUser(ctx, &pb.GetUserRequest{Id: firstUserId})
		assertNoErr(t, err)
		assertUserEqual(t, updateReq, got, cmpopts.IgnoreFields(pb.User{}, "Id", "UpdatedAt"))

		// Make sure you cannot update user with invalid id.
		got, err = cli.UpdateUser(ctx, &pb.UpdateUserRequest{User: &pb.User{
			Id:    "invalid-id",
			Email: fmt.Sprintf("%s@test.com", uuid.New().String()),
		}})
		assertErr(t, err, codes.NotFound)
	})
}

func mustSetupClient(t *testing.T) (pb.UsersClient, *grpc.ClientConn) {
	// FIXME: pass host addr to test via env.
	conn, err := grpc.Dial(":18082", grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Did not connect: %v", err)
	}
	return pb.NewUsersClient(conn), conn
}

func assertNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func assertErr(t *testing.T, err error, exp codes.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("Expected err, got nil")
	}
	if diff := cmp.Diff(exp, status.Code(err)); diff != "" {
		t.Errorf("Error code mismatch, diff: %s", diff)
	}
}

func assertUserEqual(t *testing.T, got, exp *pb.User, opts ...cmp.Option) {
	t.Helper()
	opts = append(opts, cmpopts.IgnoreUnexported(pb.User{}))
	if diff := cmp.Diff(got, exp, opts...); diff != "" {
		t.Errorf("User mismatch: %s", diff)
	}
}
