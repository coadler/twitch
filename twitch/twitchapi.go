package twitch

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// ChannelData ...
type ChannelData struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	GameID       string    `json:"game_id"`
	Type         string    `json:"type"`
	Title        string    `json:"title"`
	ViewerCount  int       `json:"viewer_count"`
	StartedAt    time.Time `json:"started_at"`
	Language     string    `json:"language"`
	ThumbnailURL string    `json:"thumbnail_url"`
}

// StreamsResponse ..
type StreamsResponse struct {
	Data []*ChannelData `json:"data"`
}

// UserData holds info about a specific twitch streamer
type UserData struct {
	ID              string `json:"id"`
	Login           string `json:"login"`
	DisplayName     string `json:"display_name"`
	Type            string `json:"type"`
	BroadcasterType string `json:"broadcaster_type"`
	Description     string `json:"description"`
	ProfileImageURL string `json:"profile_image_url"`
	OfflineImageURL string `json:"offline_image_url"`
	ViewCount       int    `json:"view_count"`
}

// UsersResponse is the data structure for the twitch users endpoint
type UsersResponse struct {
	Data []*UserData `json:"data"`
}

// Twitch is the struct that holds the basic info for the package
type Twitch struct {
	client   http.Client
	ClientID string
}

// NewAPI ...
func NewAPI(clientID string) *Twitch {
	return &Twitch{
		client:   http.Client{},
		ClientID: clientID,
	}
}

// RequestChannels requests a list of channels
func (t *Twitch) RequestChannels(channels []string) (*StreamsResponse, error) {
	req, err := http.NewRequest(
		"GET",
		"https://api.twitch.tv/helix/streams?user_login="+strings.Join(channels, "&user_login="),
		nil,
	)
	if err != nil {
		fmt.Println("Error creating request:", err.Error())
		return nil, err
	}
	req.Header.Add("Client-ID", t.ClientID)

	res, err := t.client.Do(req)
	if err != nil {
		fmt.Println("Error executing request:", err.Error())
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading body:", err.Error())
		return nil, err
	}

	var twitch *StreamsResponse
	err = json.Unmarshal(body, &twitch)
	if err != nil {
		fmt.Println("Error unmarshaling twitch response:", err.Error())
		return nil, err
	}

	return twitch, nil
}

// GetUserByID polls the twitch api for a user by their id
func (t *Twitch) GetUserByID(id string) (*UserData, error) {
	req, err := http.NewRequest(
		"GET",
		"https://api.twitch.tv/helix/users?id="+id,
		nil,
	)
	if err != nil {
		fmt.Println("Error creating request:", err.Error())
		return nil, err
	}
	req.Header.Add("Client-ID", t.ClientID)

	res, err := t.client.Do(req)
	if err != nil {
		fmt.Println("Error executing request:", err.Error())
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading body:", err.Error())
		return nil, err
	}

	var user *UsersResponse
	err = json.Unmarshal(body, &user)
	if err != nil {
		fmt.Println("Error unmarshaling twitch response:", err.Error())
		return nil, err
	}

	if len(user.Data) < 1 {
		return nil, errors.New("no data returned")
	}
	return user.Data[0], nil
}
