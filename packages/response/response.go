package response

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

type Response struct {
	Status int         `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
}

func OK(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, Response{Status: http.StatusOK, Data: data})
}

func NoContent(c echo.Context, msg string) error {
	return c.JSON(http.StatusOK, Response{Status: http.StatusOK, Data: echo.Map{"message": msg}})
}

func Wrap(err error) error {
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	if errors.Is(err, redis.Nil) {
		return ErrorNotFound
	}
	return NewError(http.StatusInternalServerError, "internal server error")
}
