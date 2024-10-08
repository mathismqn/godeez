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
