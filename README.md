# twitch-service

[![Go Report Card](https://goreportcard.com/badge/github.com/ThyLeader/twitch-service)](https://goreportcard.com/report/github.com/ThyLeader/twitch-service) [![Discord](https://discordapp.com/api/guilds/173184118492889089/widget.png)](https://discord.gg/tatsumaki) [![Discord](https://img.shields.io/badge/Discord-thy%238914-blue.svg)](https://discord.gg/tatsumaki)

## what is this thing

* This is an external service meant for keeping track of [Twitch](https://twitch.tv) channels and sending webhooks to [Discord](https://discordapp.com) when they go live

* The API endpoints are protected with [JSON Web Tokens](https://jwt.io) which allow multiple bots to use the same instance while never sharing their scope and providing useful logging

* **100% self contained**. Meaning no databases to run, automatic TLS from [letsencrypt](https://letsencrypt.org/), and no downloading dependencies. Simply download the provided binary for your OS (when I get around to doing them), fill in the example config and you're ready for production

## what is it useful for

This was designed for [Tatsumaki](https://tatsumaki.xyz), which is a Discord bot serving over 300,000 Discord servers. Meaning, we have over 200 different processes that each handle around ~1500 guilds. If each process handled this separately, it would be a mess. This microservice provides and easy to use API with support for sharding and support for multiple bots using JWT custom claims

Additionally, it's use of webhooks requires no authentication to send messages. With the webhook ID and token messages can be sent without authentication from the main bot (aka without the bot's token)

The only caveat in this solution is that the bot must have the ability to make webhooks for the channel receiving updates, and also must monitor for webhooks being deleted and notify the API of changes. If a webhook is deleted, the bot should check the API to see if any active twitch channels are in that channel and notify how you see fit. In my eyes this can be done a few ways:

1. Notifying the Discord channel that had the webhook was deleted and twitch channels have stopped tracking, then on the backend deleting the Discord channel from the db
1. Readding the webhook and notifying the Discord channel that they must remove all Twitch channel tracking before deleting the webhook, then on the backend updating each of webhook IDs and tokens to the new webhook

Personally I think option 1 should work the best, because if someone is deleting the webhook they either want the updates to stop or are stupid enough to delete something which they don't understand. Either way they deserve everything to get deleted.

## how do I make this thing work??!?

Well I'm glad you asked. If you want a step by step guide on getting started [check the wiki](https://github.com/ThyLeader/twitch-service/wiki).

**Note:** the URLs provided assume you are running this on your local machine.

### Getting a token

#### `POST` `http://127.0.0.1:1323/v1/token`

#### Overview
Get an auth token issued to you.

**Note:** Tokens expire after a certain amount of time (changeable in the config).

##### Request Body:

```json
{
  "name": "your bot name",
  "shard": 0,
  "secret": "secret password"
}
```

##### Response:

A JSON object containing the auth token under a `token` key.

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ"
}
```

#### All protected routes require a `Authorization` header:

```Header
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ
```

### Tracking a Twitch channel

#### `POST` `http://127.0.0.1:1323/v1/api/:channelid/:twitchname`

- `:channelid` is the Discord channel the user wants notifications to go in
- `:twitchname` is the name of the Twitch channel the user wants updates for

#### Overview
Start tracking a certain twitch channel.

**Note:** When tracking a new Twitch channel you should always check to see if there is already a webhook you created for that Discord channel, and if not create one to supply in the request body. Take a look at the section ["Getting info about a discord channel"]() bellow to know how to get the current webhook for a certain discord channel.

You should __**NEVER**__ have two different webhooks for a given Discord channel because this screws with how the API tracks things internally.

##### Request Body:

```json
{
  "id": "webhook id",
  "token": "webhook token"
}
```

##### Response:

...

### Stop tracking a Twitch channel

#### `DELETE` `http://127.0.0.1:1323/v1/api/webhooks/:channelid/:twitchname/:webhookid`

- `:channelid` is the Discord channel ID
- `:twitchname` is the Twitch name that's being deleted
- `:webhook` is the webhook ID that is used for the Discord channel

#### Overview
Stop tracking a certain twitch channel.

##### Request Body:

...

##### Response:

...

### Getting info about a discord channel

#### `GET` `http://127.0.0.1:1323/v1/api/webhooks/:channelid`

- `:channelid` is the Discord channel ID

#### Overview
Get info about a certain discord channel. It returns a list of all Twitch channels being tracked and the current webhook used in that channel.

##### Request Body:

...

##### Response:

A JSON object containing two keys:

- `names` is an array of Twitch channel names being tracked (or empty array if no channels are being tracked)
- `webhook` is the current webhook used for that channel

```json
{
  "names": [
    "tsm_dyrus",
    "c9sneaky",
    "shroud",
    "moonmoon_ow",
    "tsm_theoddone"
  ],
  "webhook": {
    "id": "webhook id",
    "token": "webhook token"
  }
}
```
