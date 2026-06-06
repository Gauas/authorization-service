package http

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gauas/authorization-service/config"
	"github.com/gauas/authorization-service/controller"
	"github.com/gauas/authorization-service/middlewares"
	response "github.com/gauas/authorization-service/packages/httpresp"
	"github.com/gauas/authorization-service/route"
	"github.com/labstack/echo/v4"
)

type Kernel struct {
	controller *controller.Controller
	middleware *middlewares.Middleware
	config     config.Config
}

func Register(ctrl *controller.Controller, mw *middlewares.Middleware, cfg config.Config) *Kernel {
	return &Kernel{controller: ctrl, middleware: mw, config: cfg}
}

func (k *Kernel) Start(ctx context.Context) error {
	server := echo.New()
	server.HideBanner = true
	server.HTTPErrorHandler = func(err error, c echo.Context) {
		var e *response.Error
		if errors.As(err, &e) {
			_ = c.JSON(e.Code, response.Response{Status: e.Code, Error: e.Message})
			return
		}

		var httpErr *echo.HTTPError
		if errors.As(err, &httpErr) {
			code := httpErr.Code
			_ = c.JSON(code, response.Response{Status: code, Error: fmt.Sprintf("%v", httpErr.Message)})
			return
		}

		_ = c.JSON(http.StatusInternalServerError, response.Response{Status: http.StatusInternalServerError, Error: "internal server error"})
	}

	k.middleware.RegisterGlobal(server)

	routerInstance := route.New(server, k.controller, k.middleware.Internal(), k.middleware.TokenSource())
	routerInstance.RegisterRoutes()

	addr := fmt.Sprintf(":%s", k.config.Port)

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("authorization-service http shutdown error: %v", err)
		}
	}()

	log.Printf("authorization-service http listening on %s", addr)
	if err := server.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server: %w", err)
	}

	return nil
}
