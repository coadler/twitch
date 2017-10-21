package main

import (
	"fmt"

	"github.com/ThyLeader/twitch-service/api"
	"github.com/ThyLeader/twitch-service/twitch"
	"github.com/spf13/viper"
)

var (
	clientid   string
	apisecret  string
	signsecret string
)

func init() {
	viper.AddConfigPath(".")
	viper.SetConfigType("json")
	viper.SetConfigName("config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Sprintf("error reading in config file: %s \n", err))
	}

	clientid = viper.GetString("clientid")
	apisecret = viper.GetString("apisecret")
	signsecret = viper.GetString("signsecret")
}

func main() {
	apiconfig := api.Config{
		SignSecret: signsecret,
		APISecret:  apisecret,
	}
	restapi := api.New(apiconfig)
	restapi.Start(":1323")

	twitchapi := twitch.NewAPI(clientid)
	db := twitch.NewDB(twitchapi)
	twitchapi.Open(db)
}
