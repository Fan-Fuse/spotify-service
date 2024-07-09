package db

import (
	"time"

	"gorm.io/gorm"
)

type Artist struct {
	gorm.Model
	Name        string
	SpotifyID   string
	Albums      []Album
	LastUpdated time.Time
}

func CreateArtist(name, spotifyID string) (*Artist, error) {
	artist := Artist{
		Name:        name,
		SpotifyID:   spotifyID,
		LastUpdated: time.Now(),
	}
	if err := DB.Create(&artist).Error; err != nil {
		return nil, err
	}
	return &artist, nil
}

func GetArtistByID(id uint) (*Artist, error) {
	var artist Artist
	if err := DB.Preload("Albums").First(&artist, id).Error; err != nil {
		return nil, err
	}
	return &artist, nil
}

func GetArtistBySpotifyID(spotifyID string) (*Artist, error) {
	var artist Artist
	if err := DB.Where("spotify_id = ?", spotifyID).First(&artist).Error; err != nil {
		return nil, err
	}
	return &artist, nil
}

func GetArtists(limit, offset int32) ([]Artist, error) {
	var artists []Artist
	if err := DB.Offset(int(offset)).Limit(int(limit)).Find(&artists).Error; err != nil {
		return nil, err
	}
	return artists, nil
}

func AddAlbumToArtist(artist *Artist, album *Album) error {
	artist.Albums = append(artist.Albums, *album)
	if err := DB.Save(&artist).Error; err != nil {
		return err
	}
	return nil
}
