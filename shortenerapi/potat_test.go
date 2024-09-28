package shortenerapi

import (
	"io"
	"net/http"
	"testing"
)

type testClient struct {
	Response *http.Response
}

func (c *testClient) Do(req *http.Request) (*http.Response, error) {
	return nil, nil
}

type mockAPI struct {
	Client *testClient
}

func (m *mockAPI) NewRequest(method, url string, body io.Reader) (*http.Request, error) {
	return nil, nil
}

func TestPotatShortenURL(t *testing.T) {
	// TODO
}
