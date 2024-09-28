package shortenerapi

import (
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

var shorteners = []shortener{}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type httpAPI struct {
	Client     httpClient
	NewRequest func(method, url string, body io.Reader) (*http.Request, error)
}

// replaced with mock in tests
var defaultAPI = &httpAPI{
	Client:     http.DefaultClient,
	NewRequest: http.NewRequest,
}

type shortener interface {
	Name() string
	ShortenURL(url string, api *httpAPI) (shortened string, err error)
}

func ShortenURL(url string) (shortened string, err error) {
	for _, s := range shorteners {
		shortened, err = s.ShortenURL(url, defaultAPI)
		if err == nil {
			return
		}
		log.Warn().Err(err).Str("shortener", s.Name()).Str("url", url).Msg("failed to shorten url")
	}

	err = fmt.Errorf("failed to shorten url '%s', all shorteners failed", url)
	return
}
