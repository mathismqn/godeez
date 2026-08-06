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

// URL and Format read the first source Deezer offered, which is the best one
// available for the requested quality. Both index without checking because
// FetchMedia has already rejected empty and error responses; do not call them
// on a Media obtained any other way.

func (m *Media) URL() string {
	return m.Data[0].Media[0].Sources[0].URL
}

func (m *Media) Format() string {
	return m.Data[0].Media[0].Format
}
