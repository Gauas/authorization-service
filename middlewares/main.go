package middlewares

import (
	"crypto/subtle"
	"net/http"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/gauas/authorization-service/config"
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
