package clients

import (
	"context"

	"github.com/Fan-Fuse/artist-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var artistClient proto.ArtistServiceClient

// InitUserClient creates a new UserServiceClient.
func InitArtistClient(addr string) {
	cc, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	artistClient = proto.NewArtistServiceClient(cc)
}

// CreateArtist creates a new artist.
func CreateArtist(artist *proto.Artist) (*proto.Id, error) {
	return artistClient.CreateArtist(context.Background(), artist)
}
