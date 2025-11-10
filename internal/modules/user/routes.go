package user

import (
	"github.com/gin-gonic/gin"
	"githup.com/ademdizdar/prime-api/internal/app"
)

func RoutesPublic(v1 *gin.RouterGroup, a *app.App) error {
	h, err := NewHandler(a)
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
func RoutesProtected(v1 *gin.RouterGroup, a *app.App) error {
	h, err := NewHandler(a)
	if err != nil {
		return err
	}

	g := v1.Group("/users")
	{
		g.GET("", h.List)    // GET /api/v1/users
		g.GET("/:id", h.Get) // GET /api/v1/users/:id
		g.POST("/avatar/upload-url", h.CreateAvatarUploadURL)
		g.POST("/avatar/confirm", h.ConfirmAvatar)
		g.DELETE("/avatar", h.DeleteAvatar)
		// ileride: g.GET("/me", h.Me) vb.
	}
	return nil
}
