# twitch-service

[![Go Report Card](https://goreportcard.com/badge/github.com/ThyLeader/twitch-service)](https://goreportcard.com/report/github.com/ThyLeader/twitch-service) [![Discord](https://img.shields.io/badge/Chat-7534%20online-brightgreen.svg)](https://discord.gg/tatsumaki) [![Discord](
https://img.shields.io/badge/Discord-thy%238914-blue.svg)](https://discord.gg/tatsumaki)

## what is this thing

* This is an external service meant for keeping track of [Twitch](https://twitch.tv) channels and sending webhooks to [Discord](https://discordapp.com) when they go live

* The API endpoints are protected with [JSON Web Tokens](https://jwt.io) which allow multiple bots to use the same instance while never sharing their scope while providing useful logging

* **100% self contained**. Meaning no databases to run, automatic TLS from [letsencrypt](https://letsencrypt.org/), and no downloading dependencies. Simply download the provided binary for your OS (when I get around to doing them), fill in the example config and you're ready for production

## oof this looks hard

nah, there's only like 6 endpoints and only 4 of them do anything useful :D

i'll write a block explaining all the endpoints and how to use the service when i release v1 binaries

goodbye for now
