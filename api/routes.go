package api

import (
	"net/http"
	"time"

	"github.com/ThyLeader/twitch-service/twitch"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

type jwtCustomClaims struct {
	Name  string `json:"name"`
	Shard int    `json:"shard"`
	jwt.StandardClaims
}

type tokenRequest struct {
	Name   string `json:"name" form:"name" query:"name"`
	Shard  int    `json:"shard" form:"shard" query:"shard"`
	Secret string `json:"secret" form:"secret" query:"secret"`
}

func helloWorld(c echo.Context) error {
	return c.String(http.StatusOK, "Hello World!")
}

func getToken(c echo.Context) error {
	r := new(tokenRequest)
	if err := c.Bind(r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	if r.Secret == apisecret {
		// Set custom claims
		claims := &jwtCustomClaims{
			r.Name,
			r.Shard,
			jwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
			},
		}

		// Create token with claims
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		// Generate encoded token and send it as response.
		t, err := token.SignedString([]byte(signingsecret))
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, echo.Map{
			"token": t,
		})
	}

	return echo.ErrUnauthorized
}

func checkAuth(c echo.Context) error {
	return c.String(http.StatusOK, "Authorized")
}

func getTwitchChannels(c echo.Context) error {
	names, err := twitch.DB.GetTwitchNamesByChannel(c.Param("channelid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, echo.Map{
		"names": names,
	})
}

func addWebhook(c echo.Context) error {
	channel, twitchName := c.Param("channelid"), c.Param("twitchname")
	r := new(twitch.Webhook)
	if err := c.Bind(r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	err := twitch.DB.AddChannel(twitchName, channel, r)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, echo.Map{
		"message": "success",
	})
}

func deleteWebhook(c echo.Context) error {
	cID, tName, wID := c.Param("channelid"), c.Param("twitchname"), c.Param("webhookid")

	err := twitch.DB.DeleteWebhook(tName, wID, cID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.String(http.StatusOK, "success")
}
