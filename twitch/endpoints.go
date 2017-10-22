package twitch

import "strings"

var (
	webhookEndpoint  = func(id, token string) string { return "https://discordapp.com/api/v6/webhooks/" + id + "/" + token }
	channelsEndpoint = func(channels []string) string {
		return "https://api.twitch.tv/helix/streams?user_login=" + strings.Join(channels, "&user_login=")
	}
	userEndpoint = func(id string) string { return "https://api.twitch.tv/helix/users?id=" + id }
)
