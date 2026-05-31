package controller

import (
	"net/http"

	"github.com/gauas/authorization-service/dto/request"
	"github.com/gauas/authorization-service/packages/httpresp"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (c *Controller) CreateToken(ctx echo.Context) error {
	var req request.CreateTokenRequest
	if err := ctx.Bind(&req); err != nil {
		return httpresp.NewError(http.StatusBadRequest, "invalid request body")
	}
	if req.UserID == uuid.Nil {
		return httpresp.NewError(http.StatusBadRequest, "user_id is required")
	}

	deviceID := ctx.Request().Header.Get("X-Device-ID")
	if deviceID == "" {
		return httpresp.NewError(http.StatusBadRequest, "X-Device-ID header is required")
	}

	pair, err := c.service.CreateToken(ctx.Request().Context(), req.UserID, req.Permission, deviceID)
	if err != nil {
		return err
	}

	return httpresp.OK(ctx, pair)
}
