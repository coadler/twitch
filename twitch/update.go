package twitch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	humanize "github.com/dustin/go-humanize"
)

var db *Database
var client = http.Client{}
var updateInterval time.Duration

// Open ...
func (t *Twitch) Open(twitchdb *Database, interval int64) {
	fmt.Println("open")
	db = twitchdb
	updateInterval = time.Duration(interval) * time.Second
	ticker := time.NewTicker(updateInterval)

	t.checkForUpdates()
	for {
		select {
		case <-ticker.C:
			t.checkForUpdates()
		}
	}
}

// CheckForUpdates checks
func (t *Twitch) checkForUpdates() {
	channels, err := db.GetAllTwitchChannels()
	if err != nil {
		fmt.Println("Error getting channels", err.Error())
		return
	}
	liveCopy := copyMap(t.live)

	for len(channels) > 100 {
		res, err := t.RequestChannels(channels[:100])
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		for _, e := range res.Data {
			if _, ok := t.live[e.ID]; !ok {
				t.live[e.ID] = ""
				go sendChannelLive(e)
			} else {
				delete(liveCopy, e.ID)
			}
		}

		channels = channels[100:]
	}

	res, err := t.RequestChannels(channels)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for _, e := range res.Data {
		if _, ok := t.live[e.ID]; !ok {
			t.live[e.ID] = ""
			go sendChannelLive(e)
		} else {
			delete(liveCopy, e.ID)
		}
	}

	for i := range liveCopy {
		delete(t.live, i)
	}
}

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func randStringBytes(n int) string {
	b := make([]byte, n)
	// src.Int63() generates 63 random bits, enough for letterIdxMax characters
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func copyMap(src map[string]string) map[string]string {
	dst := map[string]string{}

	for k, v := range src {
		dst[k] = v
	}

	return dst
}

func sendChannelLive(channel *ChannelData) {
	user, err := db.GetUserByID(channel.UserID)
	if err != nil {
		fmt.Println("error getting user by id:", err.Error())
		return
	}

	game, err := db.GetGameByID(channel.GameID)
	if err != nil {
		fmt.Println("error getting game by id:", err.Error())
		return
	}

	webhooks, err := db.GetWebhooksByTwitchName(user.Login)
	if err != nil {
		fmt.Println("error getting webhooks:", err.Error())
		return
	}

	fmt.Println("total webhooks:", len(webhooks))

	for _, e := range webhooks {
		go executeWebook(e, user, channel, game)
	}
}

func executeWebook(webhook *Webhook, user *UserData, channel *ChannelData, game *GameData) {
	data := discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{
			&discordgo.MessageEmbed{
				URL:         "https://twitch.tv/" + user.Login,
				Title:       user.Login + " just went live",
				Description: channel.Title,
				Author: &discordgo.MessageEmbedAuthor{
					URL:     "https://twitch.tv",
					Name:    "Twitch",
					IconURL: "https://cdn.discordapp.com/attachments/272212345340690443/374388819643858955/twitch11.png",
				},
				Image: &discordgo.MessageEmbedImage{
					URL:    strings.Replace(strings.Replace(channel.ThumbnailURL, "{width}", "1280", -1), "{height}", "720", -1) + "?please-do-not-cache-this=" + randStringBytes(15),
					Width:  1280,
					Height: 720,
				},
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: user.ProfileImageURL,
				},
				Timestamp: channel.StartedAt.Format(time.RFC3339),
				Fields: []*discordgo.MessageEmbedField{
					&discordgo.MessageEmbedField{
						Name:   "Viewers",
						Value:  strconv.Itoa(channel.ViewerCount),
						Inline: true,
					},
					&discordgo.MessageEmbedField{
						Name:   "Game",
						Value:  game.Name,
						Inline: true,
					},
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text: "Live " + humanize.Time(channel.StartedAt),
				},
			},
		},
		Username:  "Twitch",
		AvatarURL: "https://cdn.discordapp.com/attachments/196118375485669376/419336810431250432/glitch_474x356.png",
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
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println("error doing webhook request:", err.Error())
		return
	}

	switch res.StatusCode {
	case http.StatusNoContent:
		fmt.Println("webhook req success")
	case http.StatusNotFound:
		fmt.Println("webhook 404'd. fixing...")
		go db.webhook404(webhook)

	default:
		fmt.Println("webhook req didnt respond OK, responded", res.Status)
	}
}
