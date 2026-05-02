package route

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/gauas/authorization-service/controller"
)

type Router struct {
	server       *echo.Echo
	controller   *controller.Controller
	internalAuth echo.MiddlewareFunc
}

func New(server *echo.Echo, ctrl *controller.Controller, internalAuth echo.MiddlewareFunc) *Router {
	return &Router{server: server, controller: ctrl, internalAuth: internalAuth}
}

func (r *Router) RegisterRoutes() {
	api := r.server.Group("/v1/authorization")

	api.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, echo.Map{"status": "ok"})
	})

	token := api.Group("/token", r.internalAuth)
	token.POST("", r.controller.CreateToken)
	token.GET("/validate", r.controller.ValidateToken)
	token.GET("/renew", r.controller.RenewToken)
	token.DELETE("", r.controller.RevokeToken)
}
