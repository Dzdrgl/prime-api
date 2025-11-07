package user

import (
	"github.com/gin-gonic/gin"
	"githup.com/ademdizdar/prime-api/internal/app"
)

func Routes(v1 *gin.RouterGroup, a *app.App) error {
	h, err := NewHandler(a)
	if err != nil {
		return err
	}

	g := v1.Group("/users")
	{
		g.GET("", h.List)
		g.GET("/:id", h.Get)
		g.POST("", h.Create)
	}
	return nil
}
