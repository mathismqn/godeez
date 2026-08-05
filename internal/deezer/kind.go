package deezer

import "fmt"

type Kind string

const (
	KindAlbum    Kind = "album"
	KindPlaylist Kind = "playlist"
	KindArtist   Kind = "artist"
	KindTrack    Kind = "track"
)

func (k Kind) pageMethod() string {
	switch k {
	case KindAlbum:
		return "Album"
	case KindPlaylist:
		return "Playlist"
	case KindArtist:
		return "Artist"
	case KindTrack:
		return "Track"
	}

	return ""
}

func (k Kind) idKey() string {
	switch k {
	case KindAlbum:
		return "alb_id"
	case KindPlaylist:
		return "playlist_id"
	case KindArtist:
		return "art_id"
	case KindTrack:
		return "sng_id"
	}

	return ""
}

func (k Kind) newResource() (Resource, error) {
	switch k {
	case KindAlbum:
		return &Album{}, nil
	case KindPlaylist:
		return &Playlist{}, nil
	case KindArtist:
		return &Artist{}, nil
	case KindTrack:
		return &Single{}, nil
	}

	return nil, fmt.Errorf("unsupported resource type: %s", k)
}
