package deezer

// Resource is a downloadable Deezer page: an Album, Playlist, Artist or
// Single. Implementations differ only in how they decode the gw-light
// response and where they place their output, so the accessors are
// intentionally thin and are not documented individually.
//
// The unexported decode method seals the interface. Only the four types in
// this package can satisfy it, which lets Kind.newResource stay an exhaustive
// switch and guarantees FetchResource never receives an implementation whose
// wire format it does not know.
type Resource interface {
	Title() string
	Tracks() []*Track
	SetTracks(tracks []*Track)
	OutputDir(outputDir string) string
	decode(data []byte) error
}
