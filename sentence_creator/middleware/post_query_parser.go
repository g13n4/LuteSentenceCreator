package middleware

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

const ContextKey = "queryHelper"

func InjectQueryHelperMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		if c.Request().Method == http.MethodPost {
			qh, err := NewQueryHelper(c)
			if err == nil {
				c.Set(ContextKey, &qh)
			}
		}
		return next(c)
	}
}
