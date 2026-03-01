package deezer

type Media struct {
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

func (m *Media) GetURL() string {
	return m.Data[0].Media[0].Sources[0].URL
}

func (m *Media) GetFormat() string {
	return m.Data[0].Media[0].Format
}
