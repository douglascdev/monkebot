package twitchapi

import (
	"encoding/json"
	"fmt"
	"monkebot/config"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type HelixUser struct {
	ID              string    `json:"id"`
	Login           string    `json:"login"`
	DisplayName     string    `json:"display_name"`
	Type            string    `json:"type"`
	BroadcasterType string    `json:"broadcaster_type"`
	Description     string    `json:"description"`
	ProfileImageURL string    `json:"profile_image_url"`
	OfflineImageURL string    `json:"offline_image_url"`
	ViewCount       int       `json:"view_count"`      // Deprecated
	Email           string    `json:"email,omitempty"` // Optional field, omitted if empty
	CreatedAt       time.Time `json:"created_at"`
}

type helixUserResponse struct {
	Data []HelixUser `json:"data"`
}

func GetUserByName(config *config.Config, names ...string) (*[]HelixUser, error) {
	if len(names) == 0 {
		return nil, nil
	}
	if len(names) > 100 {
		return nil, fmt.Errorf("exceeded maximum number of names (100)")
	}
	var nameParams []string
	for _, name := range names {
		nameParams = append(nameParams, "login="+name)
	}
	requestURL := "https://api.twitch.tv/helix/users?" + strings.Join(nameParams, "&")
	log.Debug().Str("request", requestURL).Strs("names", names).Msg("generated helix user request")

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+config.TwitchToken)
	req.Header.Add("Client-Id", config.ClientID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get user. Status: %s", resp.Status)
	}
	defer resp.Body.Close()

	var response helixUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response.Data, nil
}

func GetUserByID(config config.Config, ids ...string) (*[]HelixUser, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	if len(ids) > 100 {
		return nil, fmt.Errorf("exceeded maximum number of names (100)")
	}
	var idsParams []string
	for _, id := range ids {
		idsParams = append(idsParams, "id="+id)
	}
	requestURL := "https://api.twitch.tv/helix/users?" + strings.Join(idsParams, "&")
	log.Debug().Str("request", requestURL).Strs("names", ids).Msg("generated helix user request")
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+config.TwitchToken)
	req.Header.Add("Client-Id", config.ClientID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get user. Status: %s", resp.Status)
	}
	defer resp.Body.Close()

	var response helixUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response.Data, nil
}
