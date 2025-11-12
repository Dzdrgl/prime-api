package http

import (
	"github.com/gin-gonic/gin"
	"githup.com/ademdizdar/prime-api/internal/app"
)

func UserRoutesPublic(v1 *gin.RouterGroup, a *app.App) error {
	h, err := NewUserHandler(a)
	if err != nil {
		return err
	}

	g := v1.Group("/users")
	{
		g.POST("", h.Create) // POST /api/v1/users  (register)  <-- TOKEN GEREKMEZ
	}
	return nil
}

// JWT zorunlu uçlar
func UserRoutesProtected(v1 *gin.RouterGroup, a *app.App) error {
	h, err := NewUserHandler(a)
	if err != nil {
		return err
	}

	g := v1.Group("/users")
	{
		g.GET("", h.List)    // GET /api/v1/users
		g.GET("/:id", h.Get) // GET /api/v1/users/:id
	}
	return nil
}

func AvatarRoutesProtected(v1 *gin.RouterGroup, a *app.App) error {
	h, err := NewAvatarHandler(a)
	if err != nil {
		return err
	}

	g := v1.Group("/users/avatar")
	{
		g.POST("/upload-url", h.CreateUploadURL)
		g.POST("/confirm", h.Confirm)
		g.DELETE("", h.Delete)
	}
	return nil
}
