package deezer

import "fmt"

// Kind is the type of Deezer resource being downloaded. It is the single
// source of truth for the four supported resources: the cmd package derives
// its download subcommands from these constants, and each kind maps to a
// gw-light page method, the request field naming its id, and a Resource
// implementation. Adding a kind means extending all three switches below.
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

// idKey returns the request field that carries the resource id. The names are
// Deezer's own internal abbreviations and do not follow from the kind, so they
// have to be spelled out. A track is a "song" on the wire.
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

// newResource returns an empty Resource for the kind. KindTrack maps to
// Single because a single track page has its own response shape rather than
// being an album with one entry.
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
