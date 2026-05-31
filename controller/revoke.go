package controller

import (
	"net/http"

	"github.com/gauas/authorization-service/packages/httpresp"
	"github.com/labstack/echo/v4"
)

func (c *Controller) RevokeToken(ctx echo.Context) error {
	refreshToken := refreshTokenFromReq(ctx)
	if refreshToken == "" {
		return httpresp.NewError(http.StatusBadRequest, "refresh_token is required")
	}

	deviceID := ctx.Request().Header.Get("X-Device-ID")

	if err := c.service.RevokeToken(ctx.Request().Context(), refreshToken, deviceID); err != nil {
		return err
	}

	return httpresp.NoContent(ctx, "token revoked")
}
