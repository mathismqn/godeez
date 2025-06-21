package deezer

type Resource interface {
	GetTitle() string
	GetType() string
	GetSongs() []*Song
	SetSongs(songs []*Song)
	GetOutputDir(outputDir string) string
	Unmarshal(data []byte) error
}
