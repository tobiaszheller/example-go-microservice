package rpc

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/tobiaszheller/example-go-microservice/service-users/proto"
	"github.com/tobiaszheller/example-go-microservice/service-users/store"
)

func toStoreUser(in *pb.User) *store.User {
	if in == nil {
		return nil
	}
	return &store.User{
		ID:        in.GetId(),
		FirstName: in.GetFirstName(),
		LastName:  in.GetLastName(),
		Nickname:  in.GetNickname(),
		Email:     in.GetEmail(),
		Country:   in.GetCountry(),
		UpdatedAt: in.GetUpdatedAt().AsTime(),
	}
}

func toPbUser(in *store.User) *pb.User {
	if in == nil {
		return nil
	}
	return &pb.User{
		Id:        in.ID,
		FirstName: in.FirstName,
		LastName:  in.LastName,
		Nickname:  in.Nickname,
		Email:     in.Email,
		Country:   in.Country,
		UpdatedAt: timestamppb.New(in.UpdatedAt),
	}
}
