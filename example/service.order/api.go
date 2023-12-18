package order

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func RegisterHandlers(group *echo.Group, service Service) {
	res := resource{service}
	g := group.Group("/order")
	g.POST("/product", res.Create)
}

type resource struct {
	svc Service
}

func (r resource) Create(c echo.Context) error {
	req := new(createRequest)
	if err := c.Bind(req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	err := r.svc.Create(ctx, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Response{
		Data: "ok",
	})
}
