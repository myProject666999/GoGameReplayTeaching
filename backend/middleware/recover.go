package middleware

import (
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/labstack/echo/v4"
)

func Recover() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					buf := make([]byte, 4096)
					n := runtime.Stack(buf, false)
					log.Printf("[PANIC] %v\n%s", r, buf[:n])
					c.JSON(http.StatusInternalServerError, map[string]string{
						"error": fmt.Sprintf("internal server error: %v", r),
					})
				}
			}()
			return next(c)
		}
	}
}

func RequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			log.Printf("[REQ] %s %s", req.Method, req.URL.Path)
			err := next(c)
			if err != nil {
				log.Printf("[ERR] %s %s -> %v", req.Method, req.URL.Path, err)
			}
			return err
		}
	}
}
