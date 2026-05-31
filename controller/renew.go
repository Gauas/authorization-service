package controller

import (
	"net/http"

	"github.com/gauas/authorization-service/packages/httpresp"
	"github.com/labstack/echo/v4"
)

func (c *Controller) RenewToken(ctx echo.Context) error {
	refreshToken := refreshTokenFromReq(ctx)
	if refreshToken == "" {
		return httpresp.NewError(http.StatusBadRequest, "refresh_token is required")
	}

	deviceID := ctx.Request().Header.Get("X-Device-ID")
	if deviceID == "" {
		return httpresp.NewError(http.StatusBadRequest, "X-Device-ID header is required")
	}

	data, err := c.service.RenewToken(ctx.Request().Context(), refreshToken, deviceID)
	if err != nil {
		return err
	}

	return httpresp.OK(ctx, data)
}
