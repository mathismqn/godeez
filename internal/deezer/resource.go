package deezer

type Resource interface {
	GetTitle() string
	GetType() string
	GetSongs() []*Song
	GetOutputDir(outputDir string) string
	Unmarshal(data []byte) error
}
