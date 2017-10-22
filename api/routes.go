package api

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

func helloWorld(c echo.Context) error {
	return c.String(http.StatusOK, "Hello World!")
}

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
