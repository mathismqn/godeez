package deezer

// mediaResponse is the get_url response. Nothing outside this package sees
// it: callers get a Media, which is the one entry worth playing, already
// picked out and paired with the id it belongs to.
type mediaResponse struct {
	Errors []mediaError `json:"errors"`
	Data   []struct {
		Media []struct {
			Format  string `json:"format"`
			Sources []struct {
				URL string `json:"url"`
			} `json:"sources"`
		}
		Errors []mediaError `json:"errors"`
	} `json:"data"`
}

type mediaError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Media is a track resolved to playable audio: where to fetch the bytes, the
// format Deezer served, and the id of the track those bytes belong to.
//
// That last id is not always the one the caller asked for, so it is carried
// here rather than left to the caller to keep alongside. Blowfish decryption
// with a mismatched key does not fail, it produces noise that is written to
// disk as if it were audio.
type Media struct {
	sourceID string
	url      string
	format   string
}

// newMedia keeps the first source of the first entry, which is the best one
// available for the requested quality. It indexes without checking because
// fetchMediaForToken has already rejected empty and error responses.
func newMedia(sourceID string, res *mediaResponse) *Media {
	return &Media{
		sourceID: sourceID,
		url:      res.Data[0].Media[0].Sources[0].URL,
		format:   res.Data[0].Media[0].Format,
	}
}

// Format is the format Deezer served, which may be lower than the one
// requested; see fetchMediaForToken for the quality chain.
func (m *Media) Format() string {
	return m.format
}

// Key is the Blowfish key for this media. It is the only way to obtain one
// outside this package, so a caller cannot pair audio with the key of a
// different track.
func (m *Media) Key() []byte {
	return blowfishKey(m.sourceID)
}
