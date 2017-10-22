package twitch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/iopred/discordgo"
)

var db *Database
var client = http.Client{}

// Open ...
func (t *Twitch) Open(twitchdb *Database, interval int64) {
	db = twitchdb
	ticker := time.NewTicker(time.Duration(interval) * time.Second)

	for {
		select {
		case <-ticker.C:
			t.CheckForUpdates()
		}
	}
}

// CheckForUpdates checks
func (t *Twitch) CheckForUpdates() {
	channels, err := db.GetTwitchChannels()
	if err != nil {
		fmt.Println("Error getting channels", err.Error())
	}

	if len(channels) < 100 {
		if len(channels) < 1 {
			fmt.Println("channel list empty")
			return
		}
		res, err := t.RequestChannels(channels)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		for _, e := range res.Data {
			if e.Type == "live" {
				if time.Now().Sub(e.StartedAt) < (2 * time.Minute) {
					go sendChannelLive(e)
				}
			}
		}
	}
}

func sendChannelLive(channel *ChannelData) {
	user, err := db.GetUserByID(channel.UserID)
	if err != nil {
		fmt.Println("error getting user by id:", err.Error())
		return
	}

	webhooks, err := db.GetWebhooks(user.Login)
	if err != nil {
		fmt.Println("error getting webhooks:", err.Error())
		return
	}

	for _, e := range webhooks {
		go executeWebook(e, user, channel)
	}
}

var webhookEndpoint = func(id, token string) string { return "https://discordapp.com/api/v6/webhooks/" + id + "/" + token }

func executeWebook(webhook *Webhook, user *UserData, channel *ChannelData) {
	data := discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{
			&discordgo.MessageEmbed{},
		},
	}

	raw, err := json.Marshal(data)
	if err != nil {
		fmt.Println("unable to marshal webhook embed:", err.Error())
		return
	}
	req, err := http.NewRequest("POST", webhookEndpoint(webhook.ID, webhook.Token), bytes.NewBuffer(raw))
	if err != nil {
		fmt.Println("unable to make webhook request:", err.Error())
		return
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println("error doing webhook request:", err.Error())
		return
	}

	switch res.StatusCode {
	case http.StatusOK:
		fmt.Println("webhook req success")
	default:
		fmt.Println("webhook req didnt respond OK, responded", res.Status)
	}
}
