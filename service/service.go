package service

import (
	"context"

	"github.com/Fan-Fuse/spotify-service/clients"
	"github.com/Fan-Fuse/spotify-service/proto"
	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

type server struct {
	proto.UnimplementedSpotifyServiceServer
}

// RegisterServer registers the server with the gRPC server
func RegisterServer(s *grpc.Server) {
	proto.RegisterSpotifyServiceServer(s, &server{})
}

func (s *server) UpdateArtist(ctx context.Context, req *proto.UpdateArtistsRequest) (*empty.Empty, error) {
	// First, get the user we want to get the artist for
	user, err := clients.GetUser(req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	// Create an oauth token from the user's access token
	token := &oauth2.Token{
		AccessToken: user.SpotifyUser.AccessToken,
		TokenType:   "Bearer",
	}

	// Next, get the artist from the user's library using their spotify ID
	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient)
	artists, err := client.CurrentUsersFollowedArtists(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get users artists")
	}

	zap.S().Infof("Got artist: %s", artists)
	// TODO: Save the artist in the database

	return &empty.Empty{}, nil
}
