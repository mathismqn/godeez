package deezer

import (
	"fmt"
)

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

func (m *Media) GetURL() (string, error) {
	if len(m.Data) == 0 || len(m.Data[0].Media) == 0 || len(m.Data[0].Media[0].Sources) == 0 {
		return "", fmt.Errorf("no media sources found")
	}

	url := m.Data[0].Media[0].Sources[0].URL
	for _, source := range m.Data[0].Media[0].Sources {
		if source.Provider == "ak" {
			url = source.URL
			break
		}
	}

	return url, nil
}

func (m *Media) GetFormat() (string, error) {
	if len(m.Data) == 0 || len(m.Data[0].Media) == 0 {
		return "", fmt.Errorf("no media format found")
	}

	return m.Data[0].Media[0].Format, nil
}
