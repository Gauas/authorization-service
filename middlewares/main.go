package middlewares

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gauas/authorization-service/config"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

type Middleware struct {
	secretKey string
}

func New(cfg config.Config) *Middleware {
	return &Middleware{secretKey: cfg.SecretKey}
}

func (m *Middleware) RegisterGlobal(server *echo.Echo) {
	server.Use(echoMiddleware.Recover())
	server.Use(echoMiddleware.Logger())
	server.Use(echoMiddleware.RequestID())
}

func (m *Middleware) Internal() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := c.Request().Header.Get("Secret-Key")
			if subtle.ConstantTimeCompare([]byte(key), []byte(m.secretKey)) != 1 {
				return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
			}
			return next(c)
		}
	}
}

func (m *Middleware) TokenSource() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			accessToken := tokenFromAuthorization(c.Request().Header.Get("Authorization"))
			if accessToken == "" {
				accessToken = c.QueryParam("token")
			}
			if accessToken == "" {
				if cookie, err := c.Cookie("access_token"); err == nil {
					accessToken = strings.TrimSpace(cookie.Value)
				}
			}
			if accessToken != "" {
				c.Set("access_token", accessToken)
			}

			refreshToken := strings.TrimSpace(c.Request().Header.Get("X-Refresh-Token"))
			if refreshToken == "" {
				if cookie, err := c.Cookie("refresh_token"); err == nil {
					refreshToken = strings.TrimSpace(cookie.Value)
				}
			}
			if refreshToken != "" {
				c.Set("refresh_token", refreshToken)
			}

			return next(c)
		}
	}
}

func tokenFromAuthorization(auth string) string {
	auth = strings.TrimSpace(auth)
	const bearerPrefix = "Bearer "
	if len(auth) <= len(bearerPrefix) || !strings.HasPrefix(auth, bearerPrefix) {
		return ""
	}
	return strings.TrimSpace(auth[len(bearerPrefix):])
}
