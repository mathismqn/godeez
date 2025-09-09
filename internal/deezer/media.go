package deezer

type Media struct {
	Errors []MediaError `json:"errors"`
	Data   []struct {
		Media []struct {
			Type    string   `json:"media_type"`
			Cipher  Cipher   `json:"cipher"`
			Format  string   `json:"format"`
			Sources []Source `json:"sources"`
		}
		Errors []MediaError `json:"errors"`
	} `json:"data"`
}

type MediaError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Cipher struct {
	Type string `json:"type"`
}

type Source struct {
	URL      string `json:"url"`
	Provider string `json:"provider"`
}

func (m *Media) GetURL() string {
	return m.Data[0].Media[0].Sources[0].URL
}

func (m *Media) GetFormat() string {
	return m.Data[0].Media[0].Format
}
