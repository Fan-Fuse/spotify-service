package service

import (
	"context"
	"os"

	"github.com/Fan-Fuse/spotify-service/clients"
	"github.com/Fan-Fuse/spotify-service/proto"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
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

func (s *server) GetArtist(ctx context.Context, req *proto.GetArtistRequest) (*proto.SpotifyArtist, error) {
	// Create an oauth token
	config := &clientcredentials.Config{
		ClientID:     os.Getenv("SPOTIFY_ID"),
		ClientSecret: os.Getenv("SPOTIFY_SECRET"),
		TokenURL:     spotifyauth.TokenURL,
	}
	token, err := config.Token(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get token")
	}

	// Next, get the artist from the user's library using their spotify ID
	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient)
	artist, err := client.GetArtist(ctx, spotify.ID(req.Id))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get artist")
	}

	// Build the images
	var images []*proto.SpotifyImage
	for _, image := range artist.Images {
		images = append(images, &proto.SpotifyImage{
			Height: int32(image.Height),
			Width:  int32(image.Width),
			Url:    image.URL,
		})
	}

	// Build the genres
	var genres []string
	for _, genre := range artist.Genres {
		genres = append(genres, genre)
	}

	return &proto.SpotifyArtist{
		Id:     artist.ID.String(),
		Name:   artist.Name,
		Images: images,
		Genres: genres,
	}, nil
}

func (s *server) GetArtistForUser(ctx context.Context, req *proto.GetArtistsForUserRequest) (*proto.GetArtistsForUserResponse, error) {
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

	// Build the response
	var responseArtists []string
	for _, artist := range artists.Artists {
		responseArtists = append(responseArtists, artist.ID.String())
	}

	return &proto.GetArtistsForUserResponse{
		ArtistIds: responseArtists,
	}, nil
}

func (s *server) GetReleasesForArtist(ctx context.Context, req *proto.GetReleasesRequest) (*proto.GetReleasesResponse, error) {
	// Create an oauth token
	config := &clientcredentials.Config{
		ClientID:     os.Getenv("SPOTIFY_ID"),
		ClientSecret: os.Getenv("SPOTIFY_SECRET"),
		TokenURL:     spotifyauth.TokenURL,
	}
	token, err := config.Token(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get token")
	}

	// Next, get all the albums for the artist
	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient)
	albumTypes := []spotify.AlbumType{spotify.AlbumTypeAlbum, spotify.AlbumTypeSingle, spotify.AlbumTypeCompilation}
	albums, err := client.GetArtistAlbums(ctx, spotify.ID(req.ArtistId), albumTypes, spotify.Limit(50))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get artist albums")
	}

	// Build the response
	var responseAlbums []*proto.SpotifyRelease
	for _, album := range albums.Albums {
		// Build the images
		var images []*proto.SpotifyImage
		for _, image := range album.Images {
			images = append(images, &proto.SpotifyImage{
				Height: int32(image.Height),
				Width:  int32(image.Width),
				Url:    image.URL,
			})
		}

		responseAlbums = append(responseAlbums, &proto.SpotifyRelease{
			Id:          album.ID.String(),
			Name:        album.Name,
			Images:      images,
			ReleaseDate: album.ReleaseDate,
		})
	}

	// Handle pagination
	for albums.Next != "" {
		zap.S().Info("Getting next page of albums", zap.String("next", albums.Next))
		err = client.NextPage(ctx, albums)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to get artist albums")
		}

		for _, album := range albums.Albums {
			// Build the images
			var images []*proto.SpotifyImage
			for _, image := range album.Images {
				images = append(images, &proto.SpotifyImage{
					Height: int32(image.Height),
					Width:  int32(image.Width),
					Url:    image.URL,
				})
			}

			responseAlbums = append(responseAlbums, &proto.SpotifyRelease{
				Id:          album.ID.String(),
				Name:        album.Name,
				Images:      images,
				ReleaseDate: album.ReleaseDate,
			})
		}
	}

	return &proto.GetReleasesResponse{
		Releases: responseAlbums,
	}, nil
}
