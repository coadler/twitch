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

Well I'm glad you asked. Here is a step by step guide on getting started

Note: the URLs provided assume you are running this on your local machine

1. First you need to get an auth token issued to you. This can be done a few ways
    1. JSON encoded body
    1. Query strings
    1. Form values

    I suggest JSON. But, if you decide to use the others they all use the same naming scheme.

    Upon every bot startup, you should send a `POST` request to the path `http://127.0.0.1:1323/v1/token`

    ```JSON
    {
        "name": "your bot name",
        "shard": 0,
        "secret": "secret password"
    }
    ```

    The API will then return

    ```JSON
    {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ"
    }
    ```

    To access all the protected routes you use the `Authorization` header

    ```Header
    Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ
    ```

    Please note that these tokens do expire after a certain amount of time, this is changeable in the config. You can do something like this to refresh the token ever X hours

    ```Go
    var Token string
    type twitchResponse struct {
        Token string `json:"token"`
    }
    func refreshToken() {
        reqData, _ := json.Marshal(map[string]interface{}{
            "name": "Tatsumaki",
            "shard": 0,
            "secret": "secretPassword",
        })

        resp, _ := http.Post("http://127.0.0.1:1323/v1/token", "application/json", bytes.NewBuffer(reqData))
        defer resp.Body.Close()

        var respData twitchResponse
        json.Unmarshal(resp.Body, &respData)

        Token = respData.Token

        time.Sleep(72 * time.Hour)
        refreshToken()
    }
    ```

    This example ignores all sense of error handling but you get the basics. After calling the function once, it calls the token endpoint and unmarshals the JSON into the data type `tokenRequest`. Then it sets the global variable `Token` to what was returned by the request. Finally it waits a given amount of time before calling itself to refresh the token again

    Wow that was a lot, but it's time to move on

1. Now we want to start tracking some Twitch channels.

    To do that we simply send a `POST` request to `http://127.0.0.1:1323/v1/api/:channelid/:twitchname`

    1. `:channelid` is the Discord channel the user wants notifications to go in
    1. `:twitchname` is the name of the Twitch channel the user wants updates for
    1. Request body:

    ```json
    {
        "id": "webhook id",
        "token": "webhook token"
    }
    ```

    When a user wants to add a Twitch channel you should always check to see if there is already a webhook you created for that Discord channel, and if not create one to supply for the request.

    You should __**NEVER**__ have two different webhooks for a given Discord channel because this screws with how the API tracks things internally.

1. What if a user wants to delete a specific Twitch channel from being tracked?

    Here what I recommend doing is using a menu. A [friend and fellow Tatsumaki developer](https://pyraxo.moe/) wrote a blog post about it [here](https://blog.pyraxo.moe/2017/01/bot-menus/).

    ![menu](https://cdn.discordapp.com/attachments/309741345264631818/377027893580005376/unknown.png)
    ![response](https://cdn.discordapp.com/attachments/309741345264631818/377029933270040577/unknown.png)

    Never heard of them? Let me explain

    When a user types the command `t!twitch remove` it brings up an expanded context menu, allowing the user to pick between the current Twitch channels being tracked. All the user has to do is type the number corresponding with the Twitch channel and that one is removed. Easy as that!

    Tatsumaki uses these a lot, and they're a great way to enhance user experience when used in moderation. Now lets go through what API calls you need to make for this to work.

    First, when the user calls the remove command, we need to make a call to get all the Twitch channels currently active in the Discord channel in question

    Send a`POST` request to `http://127.0.0.1:1323/v1/api/webhooks/:channelid` where `:channelid` is the Discord channel ID

    You'll receive back:

    ```json
    {
        "names": [
            "tsm_dyrus",
            "c9sneaky",
            "shroud",
            "moonmoon_ow",
            "tsm_theoddone"
        ]
    }
    ```

    You can then use this to populate the menu or handle an empty array. Once the user has picked the entry to delete, we'd send a `DELETE` request to `http://127.0.0.1:1323/v1/api/webhooks/:channelid/:twitchname/:webhookid`

    1. `:channelid` is the Discord channel ID
    1. `:twitchname` is the Twitch name that's being deleted
    1. `:webhook` is the webhook ID that is used for the Discord channel

    Upon a successful response we can notify the user the command succeeded!

That's it for now. I still have to add in deleting all the Twitch channels being tracked on a specific Discord channel and I think it'll be ready for release