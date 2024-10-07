package models

type Media []struct {
	Type    string   `json:"media_type"`
	Cipher  Cipher   `json:"cipher"`
	Format  string   `json:"format"`
	Sources []Source `json:"sources"`
}

type Cipher struct {
	Type string `json:"type"`
}

type Source struct {
	URL      string `json:"url"`
	Provider string `json:"provider"`
}

type MediaResponse struct {
	Errors []MediaError `json:"errors"`
	Data   []struct {
		Media Media `json:"media"`
	}
}

type MediaError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
