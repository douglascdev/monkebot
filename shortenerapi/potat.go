package shortenerapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func init() {
	shorteners = append(shorteners, &potatShortener{})
}

type potatShortenerJSONResult struct {
	Data []struct {
		URL string `json:"url"`
	} `json:"data"`
	Duration float64 `json:"duration"`
	Status   int     `json:"status"`
	Errors   []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type potatShortener struct{}

func (s *potatShortener) Name() string {
	return "potat"
}

func (s *potatShortener) ShortenURL(url string, api *httpAPI) (shortened string, err error) {
	requestURL := "https://api.potat.app/redirect"
	body := bytes.NewBuffer([]byte(fmt.Sprintf(`{"url": "%s"}`, url)))
	var req *http.Request
	req, err = api.NewRequest("POST", requestURL, body)
	if err != nil {
		err = fmt.Errorf("failed to create request for shortening url '%s': %w", url, err)
		return
	}

	var resp *http.Response
	resp, err = api.Client.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to send request to shorten url '%s': %w", url, err)
		return
	}
	defer resp.Body.Close()

	result := &potatShortenerJSONResult{}
	if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
		err = fmt.Errorf("failed to decode response from shorten url '%s': %w", url, err)
		return
	}

	switch resp.StatusCode {
	case 200:
		shortened = result.Data[0].URL
	default:
		errorMsgs := make([]string, 0, len(result.Errors))
		for _, err := range result.Errors {
			errorMsgs = append(errorMsgs, err.Message)
		}
		err = fmt.Errorf("failed to shorten url '%s' with status code %s: %s", url, resp.Status, strings.Join(errorMsgs, " | "))
	}

	return
}
