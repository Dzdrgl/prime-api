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

	ug := v1.Group("/users")
	{
		ug.GET("", h.List)
		ug.GET("/:id", h.Get)
		ug.POST("", h.Create)

		ug.POST("/avatar/upload-url", h.CreateAvatarUploadURL) // JWT protected group’ta
		ug.POST("/avatar/confirm", h.ConfirmAvatar)
		ug.DELETE("/avatar", h.DeleteAvatar)
	}
	return nil
}
