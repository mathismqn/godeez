package deezer

type Resource interface {
	Title() string
	Tracks() []*Track
	SetTracks(tracks []*Track)
	OutputDir(outputDir string) string
	decode(data []byte) error
}
