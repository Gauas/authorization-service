package controller

import "github.com/labstack/echo/v4"

func accessTokenFromReq(ctx echo.Context) string {
	if val, ok := ctx.Get("access_token").(string); ok {
		return val
	}
	return ""
}

func refreshTokenFromReq(ctx echo.Context) string {
	if val, ok := ctx.Get("refresh_token").(string); ok {
		return val
	}
	return ""
}
