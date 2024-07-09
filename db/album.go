package db

import "gorm.io/gorm"

type Album struct {
	gorm.Model
	Name      string
	SpotifyID string
	ArtistID  uint
}
