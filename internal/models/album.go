package models

type Album struct {
	Songs struct {
		Data []Song `json:"DATA"`
	} `json:"SONGS"`
}
