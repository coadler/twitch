package api

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// API ...
type API struct {
	router *echo.Echo
}

// Config ...
type Config struct {
	SignSecret string
	APISecret  string
}

var (
	signingsecret string
	apisecret     string
)

// New creates a new instance of the rest api
func New(config Config) *API {
	if config.SignSecret == "" || config.APISecret == "" {
		panic("api and signing secret not set")
	}
	signingsecret = config.SignSecret
	apisecret = config.APISecret

	api := &API{}
	api.router = echo.New()

	api.initAPI()

	return api
}

// Start starts the API
func (a *API) Start(port string) {
	a.router.Logger.Fatal(a.router.Start(port))
}

func (a *API) initAPI() {
	// Middleware
	a.router.Use(middleware.Logger())
	a.router.Use(middleware.Recover())

	// Routes
	a.router.GET("/", helloWorld)
	a.router.POST("/v1/token", getToken)

	v1 := a.router.Group("/v1/api")
	config := middleware.JWTConfig{
		Claims:     &jwtCustomClaims{},
		SigningKey: []byte(signingsecret),
	}
	v1.Use(middleware.JWTWithConfig(config))
	//v1.
}

// tentative routes

// 						UNAUTHENTICATED
// GET 	/ 										- hello world
// GET 	/v1/token 								- get jwt for auth

//						 AUTHENTICATED
// GET 	/v1/api 								- check jwt validity
// GET 	/v1/api/webhooks/:channelid				- returns a list of webhooks for a specific channel
// POST /v1/api/webhooks/:channelid/:twitchname	- make a new webhook
// DEL 	/v1/api/webhooks/:channelid/:twitchname	- delete a webhook
