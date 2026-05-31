package route

import (
	"net/http"

	"github.com/gauas/authorization-service/controller"
	"github.com/labstack/echo/v4"
)

type Router struct {
	server       *echo.Echo
	controller   *controller.Controller
	internalAuth echo.MiddlewareFunc
	tokenSource  echo.MiddlewareFunc
}

func New(server *echo.Echo, ctrl *controller.Controller, internalAuth, tokenSource echo.MiddlewareFunc) *Router {
	return &Router{server: server, controller: ctrl, internalAuth: internalAuth, tokenSource: tokenSource}
}

func (r *Router) RegisterRoutes() {
	api := r.server.Group("/v1/authorization")

	api.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, echo.Map{"status": "ok"})
	})

	token := api.Group("/token", r.internalAuth, r.tokenSource)
	token.POST("", r.controller.CreateToken)
	token.GET("/validate", r.controller.ValidateToken)
	token.GET("/renew", r.controller.RenewToken)
	token.DELETE("", r.controller.RevokeToken)
}
