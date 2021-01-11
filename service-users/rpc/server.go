package rpc

import (
	"context"
	"errors"
	"strings"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/tobiaszheller/example-go-microservice/service-users/proto"
	"github.com/tobiaszheller/example-go-microservice/service-users/store"
)

type server struct {
	pb.UnimplementedUsersServer
	storer          storer
	eventsPublisher eventsPublisher
}

func New(storer storer, eventsPublisher eventsPublisher) *server {
	return &server{
		storer:          storer,
		eventsPublisher: eventsPublisher,
	}
}

type storer interface {
	CreateUser(context.Context, *store.User) (*store.User, error)
	UpdateUser(context.Context, *store.User) (*store.User, error)
	GetUser(context.Context, string) (*store.User, error)
}

type eventsPublisher interface {
	Publish(context.Context, proto.Message) error
}

func (s *server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	if err := validateCreateUserRequest(req); err != nil {
		return nil, err
	}
	user, err := s.storer.CreateUser(ctx, toStoreUser(req.GetUser()))
	if err != nil {
		if errors.Is(err, store.ErrUserAlreadyExists) {
			return nil, grpc.Errorf(codes.AlreadyExists, "failed to create user: %v", err)
		}
		return nil, grpc.Errorf(codes.Internal, "failed to create user: %v", err)
	}
	out := toPbUser(user)
	if err := s.eventsPublisher.Publish(ctx, &pb.UserCreated{User: out}); err != nil {
		return nil, grpc.Errorf(codes.Internal, "failed to publish event: %v", err)
	}
	return out, nil
}

func validateCreateUserRequest(req *pb.CreateUserRequest) error {
	// TODO: replace with better validation builder.
	eb := strings.Builder{}
	if req.GetUser().GetId() != "" {
		eb.WriteString("'user.id' cannot be provided,")
	}
	if req.GetUser().GetUpdatedAt() != nil {
		eb.WriteString("'user.updated_at' cannot be provided,")
	}
	// TODO: check for valid email signiture.
	if req.GetUser().GetEmail() == "" {
		eb.WriteString("'user.email' must be provided,")
	}
	// TODO: validate country for ISO 3166-1 alpha-2.
	if eb.String() != "" {
		return grpc.Errorf(codes.InvalidArgument, "invalid request: %s", eb.String())
	}
	return nil
}

func (s *server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	if err := validateUpdateUserRequest(req); err != nil {
		return nil, err
	}
	user, err := s.storer.UpdateUser(ctx, toStoreUser(req.GetUser()))
	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			return nil, grpc.Errorf(codes.NotFound, "failed to update user: %v", err)
		}
		return nil, grpc.Errorf(codes.Internal, "failed to update user: %v", err)
	}
	out := toPbUser(user)
	if err := s.eventsPublisher.Publish(ctx, &pb.UserUpdated{User: out}); err != nil {
		return nil, grpc.Errorf(codes.Internal, "failed to publish event: %v", err)
	}
	return out, nil
}

func validateUpdateUserRequest(req *pb.UpdateUserRequest) error {
	// TODO: replace with better validation builder.
	eb := strings.Builder{}
	if req.GetUser().GetId() == "" {
		eb.WriteString("'user.id' must be provided,")
	}
	if req.GetUser().GetUpdatedAt() != nil {
		eb.WriteString("'user.updated_at' cannot be provided,")
	}
	// TODO: check for valid email signiture.
	if req.GetUser().GetEmail() == "" {
		eb.WriteString("'user.email' must be provided,")
	}
	// TODO: validate country for ISO 3166-1 alpha-2.
	if eb.String() != "" {
		return grpc.Errorf(codes.InvalidArgument, "invalid request: %s", eb.String())
	}
	return nil
}

func (s *server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	if err := validateGetUserRequest(req); err != nil {
		return nil, err
	}
	user, err := s.storer.GetUser(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			return nil, grpc.Errorf(codes.NotFound, "failed to get user: %v", err)
		}
		return nil, grpc.Errorf(codes.Internal, "failed to get user: %v", err)
	}
	return toPbUser(user), nil
}

func validateGetUserRequest(req *pb.GetUserRequest) error {
	// TODO: replace with better validation builder.
	eb := strings.Builder{}
	if req.GetId() == "" {
		eb.WriteString("'id' must be provided,")
	}
	if eb.String() != "" {
		return grpc.Errorf(codes.InvalidArgument, "invalid request: %s", eb.String())
	}
	return nil
}
