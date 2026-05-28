package controller

import (
	"net/http"

	"github.com/gauas/authorization-service/packages/response"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const BEARER_PREFIX = "Bearer "

type createTokenRequest struct {
	UserID     uuid.UUID `json:"user_id"`
	Permission string    `json:"permission"`
}

func (c *Controller) CreateToken(ctx echo.Context) error {
	var req createTokenRequest
	if err := ctx.Bind(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "invalid request body")
	}
	if req.UserID == uuid.Nil {
		return response.NewError(http.StatusBadRequest, "user_id is required")
	}

	deviceID := ctx.Request().Header.Get("X-Device-ID")
	if deviceID == "" {
		return response.NewError(http.StatusBadRequest, "X-Device-ID header is required")
	}

	pair, err := c.service.CreateToken(ctx.Request().Context(), req.UserID, req.Permission, deviceID)
	if err != nil {
		return response.Wrap(err)
	}

	return response.OK(ctx, pair)
}

func (c *Controller) ValidateToken(ctx echo.Context) error {
	tokenStr := tokenFromReq(ctx)
	if tokenStr == "" {
		return response.NewError(http.StatusBadRequest, "token is required")
	}

	claims, err := c.service.ValidateToken(ctx.Request().Context(), tokenStr)
	if err != nil {
		return response.Wrap(err)
	}

	return response.OK(ctx, echo.Map{
		"user_id":    claims.UserID,
		"permission": claims.Permission,
		"device_id":  claims.DeviceID,
	})
}

func tokenFromReq(ctx echo.Context) string {
	token := ctx.QueryParam("token")
	if token != "" {
		return token
	}

	auth := ctx.Request().Header.Get("Authorization")
	if len(auth) <= len(BEARER_PREFIX) {
		return ""
	}
	if auth[:len(BEARER_PREFIX)] != BEARER_PREFIX {
		return ""
	}
	return auth[len(BEARER_PREFIX):]
}

func (c *Controller) RenewToken(ctx echo.Context) error {
	refreshToken := ctx.Request().Header.Get("X-Refresh-Token")
	if refreshToken == "" {
		return response.NewError(http.StatusBadRequest, "X-Refresh-Token header is required")
	}

	deviceID := ctx.Request().Header.Get("X-Device-ID")
	if deviceID == "" {
		return response.NewError(http.StatusBadRequest, "X-Device-ID header is required")
	}

	result, err := c.service.RenewToken(ctx.Request().Context(), refreshToken, deviceID)
	if err != nil {
		return response.Wrap(err)
	}

	return response.OK(ctx, result)
}

func (c *Controller) RevokeToken(ctx echo.Context) error {
	refreshToken := ctx.Request().Header.Get("X-Refresh-Token")
	if refreshToken == "" {
		return response.NewError(http.StatusBadRequest, "X-Refresh-Token header is required")
	}

	deviceID := ctx.Request().Header.Get("X-Device-ID")

	if err := c.service.RevokeToken(ctx.Request().Context(), refreshToken, deviceID); err != nil {
		return response.Wrap(err)
	}

	return response.NoContent(ctx, "token revoked")
}
