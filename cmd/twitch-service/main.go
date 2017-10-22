package main

import (
	"fmt"
	"strconv"

	"github.com/ThyLeader/twitch-service/api"
	"github.com/ThyLeader/twitch-service/twitch"
	"github.com/spf13/viper"
)

var (
	clientid       string
	apisecret      string
	signsecret     string
	updateinterval int64
)

func init() {
	viper.AddConfigPath(".")
	viper.SetConfigType("json")
	viper.SetConfigName("config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Sprintf("error reading in config file: %s \n", err))
	}

	clientid = viper.GetString("client-id")
	apisecret = viper.GetString("api-secret")
	signsecret = viper.GetString("sign-secret")
	i := viper.GetString("update-interval")
	updateinterval, err = strconv.ParseInt(i, 10, 64)
	if err != nil {
		panic("update interval is not an int")
	}

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
	twitchapi.Open(db, updateinterval)
}
