package deezer

type Resource interface {
	GetTitle() string
	GetType() string
	GetTracks() []*Track
	SetTracks(tracks []*Track)
	GetOutputDir(outputDir string) string
	Unmarshal(data []byte) error
}
