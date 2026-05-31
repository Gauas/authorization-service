package controller

import (
	"net/http"

	"github.com/gauas/authorization-service/packages/httpresp"
	"github.com/labstack/echo/v4"
)

func (c *Controller) ValidateToken(ctx echo.Context) error {
	tokenStr := accessTokenFromReq(ctx)
	if tokenStr == "" {
		return httpresp.NewError(http.StatusBadRequest, "token is required")
	}

	claims, err := c.service.ValidateToken(ctx.Request().Context(), tokenStr)
	if err != nil {
		return err
	}

	return httpresp.OK(ctx, echo.Map{
		"user_id":    claims.UserID,
		"permission": claims.Permission,
		"device_id":  claims.DeviceID,
	})
}
