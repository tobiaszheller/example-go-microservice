package rpc

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/tobiaszheller/example-go-microservice/service-users/proto"
	"github.com/tobiaszheller/example-go-microservice/service-users/store"
)

type check func(*pb.User, *mockPublisher, error, *testing.T)

var (
	checks   = func(cs ...check) []check { return cs }
	hasError = func(exp string) check {
		return func(_ *pb.User, _ *mockPublisher, err error, t *testing.T) {
			t.Helper()
			if err == nil {
				t.Fatalf("Expected err but got nil")
			}
			if diff := cmp.Diff(exp, err.Error()); diff != "" {
				t.Errorf("Error mismatch, diff: %s", diff)
			}
		}
	}
	hasNoError = func() check {
		return func(_ *pb.User, _ *mockPublisher, err error, t *testing.T) {
			t.Helper()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		}
	}
	hasUser = func(exp *pb.User, opts ...cmp.Option) check {
		return func(resp *pb.User, _ *mockPublisher, _ error, t *testing.T) {
			t.Helper()
			opts = append(opts, cmpopts.IgnoreUnexported(pb.User{}, timestamppb.Timestamp{}))
			if diff := cmp.Diff(exp, resp, opts...); diff != "" {
				t.Errorf("User mismatch, diff: %s", diff)
			}
		}
	}
	hasPublishedNEvents = func(exp int) check {
		return func(_ *pb.User, mp *mockPublisher, _ error, t *testing.T) {
			t.Helper()
			if diff := cmp.Diff(exp, len(mp.events)); diff != "" {
				t.Fatalf("Number of published events mismatch, diff: %s", diff)
			}
		}
	}
	hasLastEvent = func(exp proto.Message, opts ...cmp.Option) check {
		return func(_ *pb.User, mp *mockPublisher, _ error, t *testing.T) {
			t.Helper()
			opts = append(opts, cmpopts.IgnoreUnexported(pb.User{}, timestamppb.Timestamp{}))
			if diff := cmp.Diff(exp, mp.events[len(mp.events)-1], opts...); diff != "" {
				t.Errorf("Published events mismatch, diff: %s", diff)
			}
		}
	}
)

func TestCreateUser(t *testing.T) {
	testCases := []struct {
		desc             string
		createUserRespFn func() (*store.User, error)
		req              *pb.CreateUserRequest
		checks           []check
	}{
		{
			desc: "invalid req",
			req: &pb.CreateUserRequest{User: &pb.User{
				Id:        "id",
				UpdatedAt: timestamppb.New(time.Now()),
			}},
			checks: checks(
				hasError("rpc error: code = InvalidArgument desc = invalid request: 'user.id' cannot be provided,'user.updated_at' cannot be provided,'user.email' must be provided,"),
			),
		},
		{
			desc: "valid req, already exists user with given email",
			req: &pb.CreateUserRequest{User: &pb.User{
				Email: "test@test.com",
			}},
			createUserRespFn: func() (*store.User, error) {
				return nil, store.ErrUserAlreadyExists
			},
			checks: checks(
				hasError("rpc error: code = AlreadyExists desc = failed to create user: user already exists"),
			),
		},
		{
			desc: "valid req, undifiend store err",
			req: &pb.CreateUserRequest{User: &pb.User{
				Email: "test@test.com",
			}},
			createUserRespFn: func() (*store.User, error) {
				return nil, fmt.Errorf("some err")
			},
			checks: checks(
				hasError("rpc error: code = Internal desc = failed to create user: some err"),
			),
		},
		{
			desc: "valid req, user created",
			req: &pb.CreateUserRequest{User: &pb.User{
				Email: "test@test.com",
			}},
			createUserRespFn: func() (*store.User, error) {
				return &store.User{
					ID:        "id-1",
					UpdatedAt: time.Date(2020, 12, 10, 11, 0, 0, 0, time.UTC),
				}, nil
			},
			checks: checks(
				hasNoError(),
				hasUser(&pb.User{
					Id:        "id-1",
					UpdatedAt: timestamppb.New(time.Date(2020, 12, 10, 11, 0, 0, 0, time.UTC)),
				}),
				hasPublishedNEvents(1),
				hasLastEvent(&pb.UserCreated{User: &pb.User{
					Id:        "id-1",
					UpdatedAt: timestamppb.New(time.Date(2020, 12, 10, 11, 0, 0, 0, time.UTC)),
				}}, cmpopts.IgnoreUnexported(pb.UserCreated{})),
			),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			store := &mockStore{
				createUserRespFn: tC.createUserRespFn,
			}
			eventsPublisher := &mockPublisher{}
			svc := New(store, eventsPublisher)
			resp, err := svc.CreateUser(context.Background(), tC.req)
			for _, ch := range tC.checks {
				ch(resp, eventsPublisher, err, t)
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	testCases := []struct {
		desc             string
		updateUserRespFn func() (*store.User, error)
		req              *pb.UpdateUserRequest
		checks           []check
	}{
		{
			desc: "invalid req",
			req: &pb.UpdateUserRequest{User: &pb.User{
				UpdatedAt: timestamppb.New(time.Now()),
			}},
			checks: checks(
				hasError("rpc error: code = InvalidArgument desc = invalid request: 'user.id' must be provided,'user.updated_at' cannot be provided,'user.email' must be provided,"),
			),
		},
		{
			desc: "valid req, user not found",
			req: &pb.UpdateUserRequest{User: &pb.User{
				Id:    "id-1",
				Email: "test@test.com",
			}},
			updateUserRespFn: func() (*store.User, error) {
				return nil, store.ErrUserNotFound
			},
			checks: checks(
				hasError("rpc error: code = NotFound desc = failed to update user: user not found"),
			),
		},
		{
			desc: "valid req, other store err",
			req: &pb.UpdateUserRequest{User: &pb.User{
				Id:    "id-1",
				Email: "test@test.com",
			}},
			updateUserRespFn: func() (*store.User, error) {
				return nil, fmt.Errorf("some err")
			},
			checks: checks(
				hasError("rpc error: code = Internal desc = failed to update user: some err"),
			),
		},
		{
			desc: "valid req, user updated",
			req: &pb.UpdateUserRequest{User: &pb.User{
				Id:    "id-1",
				Email: "test@test.com",
			}},
			updateUserRespFn: func() (*store.User, error) {
				return &store.User{
					ID:        "id-1",
					UpdatedAt: time.Date(2020, 12, 10, 11, 0, 0, 0, time.UTC),
				}, nil
			},
			checks: checks(
				hasNoError(),
				hasUser(&pb.User{
					Id:        "id-1",
					UpdatedAt: timestamppb.New(time.Date(2020, 12, 10, 11, 0, 0, 0, time.UTC)),
				}),
				hasPublishedNEvents(1),
				hasLastEvent(&pb.UserUpdated{User: &pb.User{
					Id:        "id-1",
					UpdatedAt: timestamppb.New(time.Date(2020, 12, 10, 11, 0, 0, 0, time.UTC)),
				}}, cmpopts.IgnoreUnexported(pb.UserUpdated{})),
			),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			store := &mockStore{
				updateUserRespFn: tC.updateUserRespFn,
			}
			eventsPublisher := &mockPublisher{}
			svc := New(store, eventsPublisher)
			resp, err := svc.UpdateUser(context.Background(), tC.req)
			for _, ch := range tC.checks {
				ch(resp, eventsPublisher, err, t)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	testCases := []struct {
		desc          string
		getUserRespFn func() (*store.User, error)
		req           *pb.GetUserRequest
		checks        []check
	}{
		{
			desc: "invalid req",
			req:  &pb.GetUserRequest{},
			checks: checks(
				hasError("rpc error: code = InvalidArgument desc = invalid request: 'id' must be provided,"),
			),
		},
		{
			desc: "valid req, user not found",
			req: &pb.GetUserRequest{
				Id: "id-1",
			},
			getUserRespFn: func() (*store.User, error) {
				return nil, store.ErrUserNotFound
			},
			checks: checks(
				hasError("rpc error: code = NotFound desc = failed to get user: user not found"),
			),
		},
		{
			desc: "valid req, other store err",
			req: &pb.GetUserRequest{
				Id: "id-1",
			},
			getUserRespFn: func() (*store.User, error) {
				return nil, fmt.Errorf("some err")
			},
			checks: checks(
				hasError("rpc error: code = Internal desc = failed to get user: some err"),
			),
		},
		{
			desc: "valid req, user returned",
			req: &pb.GetUserRequest{
				Id: "id-1",
			},
			getUserRespFn: func() (*store.User, error) {
				return &store.User{
					ID:        "id-1",
					UpdatedAt: time.Date(2020, 12, 10, 11, 0, 0, 0, time.UTC),
				}, nil
			},
			checks: checks(
				hasNoError(),
				hasUser(&pb.User{
					Id:        "id-1",
					UpdatedAt: timestamppb.New(time.Date(2020, 12, 10, 11, 0, 0, 0, time.UTC)),
				}),
				hasPublishedNEvents(0),
			),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			store := &mockStore{
				getUserRespFn: tC.getUserRespFn,
			}
			eventsPublisher := &mockPublisher{}
			svc := New(store, eventsPublisher)
			resp, err := svc.GetUser(context.Background(), tC.req)
			for _, ch := range tC.checks {
				ch(resp, eventsPublisher, err, t)
			}
		})
	}
}

type mockStore struct {
	createUserRespFn func() (*store.User, error)
	updateUserRespFn func() (*store.User, error)
	getUserRespFn    func() (*store.User, error)
}

func (m *mockStore) CreateUser(context.Context, *store.User) (*store.User, error) {
	return m.createUserRespFn()
}

func (m *mockStore) UpdateUser(context.Context, *store.User) (*store.User, error) {
	return m.updateUserRespFn()
}

func (m *mockStore) GetUser(context.Context, string) (*store.User, error) {
	return m.getUserRespFn()
}

type mockPublisher struct {
	mu     sync.Mutex
	events []proto.Message
}

func (m *mockPublisher) Publish(_ context.Context, in proto.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, in)
	return nil
}
