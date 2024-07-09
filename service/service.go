package service

import (
	"context"
	"strconv"
	"time"

	"github.com/Fan-Fuse/spotify-service/clients"
	"github.com/Fan-Fuse/spotify-service/db"
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

func (s *server) UpdateArtists(ctx context.Context, req *proto.UpdateArtistsRequest) (*empty.Empty, error) {
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

	// Loop through the artists and update them in the database
	for _, artist := range artists.Artists {
		dbArtist, err := db.GetArtistBySpotifyID(artist.ID.String())
		if err != nil {
			dbArtist, err = db.CreateArtist(artist.Name, artist.ID.String())
			if err != nil {
				return nil, status.Error(codes.Internal, "failed to create artist")
			}
		} else {
			dbArtist.LastUpdated = time.Now()
			if err := db.DB.Save(dbArtist).Error; err != nil {
				return nil, status.Error(codes.Internal, "failed to update artist")
			}
		}

		zap.S().Infof("Updated artist %s", dbArtist.Name)
	}

	return &empty.Empty{}, nil
}

func (s *server) GetArtists(ctx context.Context, req *proto.GetArtistsRequest) (*proto.GetArtistsResponse, error) {
	// Getartists with the limit and offset
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.Limit > 50 {
		req.Limit = 50
	}
	artists, err := db.GetArtists(req.Limit, req.Offset)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get artists")
	}

	// Convert the artists to protobuf
	var pbArtists []*proto.Artist
	for _, artist := range artists {
		id := strconv.Itoa(int(artist.ID))
		pbArtists = append(pbArtists, &proto.Artist{
			Id:          id,
			Name:        artist.Name,
			SpotifyId:   artist.SpotifyID,
			LastUpdated: artist.LastUpdated.Format(time.RFC3339),
		})
	}

	return &proto.GetArtistsResponse{Artists: pbArtists}, nil
}
