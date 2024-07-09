package service

import (
	"context"
	"os"

	artistProto "github.com/Fan-Fuse/artist-service/proto"
	"github.com/Fan-Fuse/spotify-service/clients"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

func HandleSpotifyArtist(ctx context.Context, spotifyID string, client *spotify.Client) error {
	// Check if we already have a client (this happens when we call this function from the user handling)
	if client == nil {

		// Create an oauth token
		config := &clientcredentials.Config{
			ClientID:     os.Getenv("SPOTIFY_ID"),
			ClientSecret: os.Getenv("SPOTIFY_SECRET"),
			TokenURL:     spotifyauth.TokenURL,
		}
		token, err := config.Token(ctx)
		if err != nil {
			return err
		}

		httpClient := spotifyauth.New().Client(ctx, token)
		client = spotify.New(httpClient)
	}
	// Next, get the artist
	artist, err := client.GetArtist(ctx, spotify.ID(spotifyID))
	if err != nil {
		return err
	}

	// Build the images
	var images []*artistProto.Image
	for _, image := range artist.Images {
		images = append(images, &artistProto.Image{
			Height: int32(image.Height),
			Width:  int32(image.Width),
			Url:    image.URL,
		})
	}

	// Retrieve all the albums for the artist
	albumTypes := []spotify.AlbumType{spotify.AlbumTypeAlbum, spotify.AlbumTypeSingle, spotify.AlbumTypeCompilation}
	albums, err := client.GetArtistAlbums(ctx, spotify.ID(spotifyID), albumTypes, spotify.Limit(50))
	if err != nil {
		return err
	}

	// Build the albums
	var responseAlbums []*artistProto.Album
	for _, album := range albums.Albums {
		responseAlbums = append(responseAlbums, &artistProto.Album{
			Id:          album.ID.String(),
			Name:        album.Name,
			ReleaseDate: &timestamppb.Timestamp{Seconds: album.ReleaseDateTime().Unix()},
			Externals: &artistProto.Externals{
				Spotify: album.ID.String(),
			},
		})
	}

	// Handle pagination
	for albums.Next != "" {
		zap.S().Info("Getting next page of albums", zap.String("next", albums.Next))
		err = client.NextPage(ctx, albums)
		if err != nil {
			return err
		}

		for _, album := range albums.Albums {
			responseAlbums = append(responseAlbums, &artistProto.Album{
				Id:          album.ID.String(),
				Name:        album.Name,
				ReleaseDate: &timestamppb.Timestamp{Seconds: album.ReleaseDateTime().Unix()},
				Externals: &artistProto.Externals{
					Spotify: album.ID.String(),
				},
			})
		}
	}

	// Create the artist
	id, err := clients.CreateArtist(&artistProto.Artist{
		Name:      artist.Name,
		Images:    images,
		Albums:    responseAlbums,
		Externals: &artistProto.Externals{Spotify: artist.ID.String()},
	})
	if err != nil {
		return err
	}

	zap.S().Info("Created artist", zap.String("id", id.Id))

	return nil
}

func HandleSpotifyUser(ctx context.Context, userId string) error {
	// First, get the user we want to get the artist for
	user, err := clients.GetUser(userId)
	if err != nil {
		zap.S().Error("Failed to get user", zap.Error(err))
		return err
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
		zap.S().Error("Failed to get followed artists", zap.Error(err))
		return err
	}

	// Build an array of artist IDs
	var responseArtists []string
	for _, artist := range artists.Artists {
		responseArtists = append(responseArtists, artist.ID.String())
	}

	// TODO: Handle pagination

	// run a "HandleSpotifyArtist" for each artist
	for _, artist := range responseArtists {
		err = HandleSpotifyArtist(ctx, artist, client)
		if err != nil {
			zap.S().Error("Failed to handle artist", zap.Error(err))
			return err
		}
	}

	return nil
}
