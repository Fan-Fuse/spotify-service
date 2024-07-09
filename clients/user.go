package clients

import (
	"context"

	"github.com/Fan-Fuse/user-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var userClient proto.UserServiceClient

// InitUserClient creates a new UserServiceClient.
func InitUserClient(addr string) {
	cc, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	userClient = proto.NewUserServiceClient(cc)
}

// GetUser gets a user by ID.
func GetUser(id string) (*proto.GetUserResponse, error) {
	return userClient.GetUser(context.Background(), &proto.GetUserRequest{Id: id})
}
