package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"githup.com/ademdizdar/prime-api/internal/app"
	"githup.com/ademdizdar/prime-api/internal/common"
	"githup.com/ademdizdar/prime-api/internal/modules/user"
	userdto "githup.com/ademdizdar/prime-api/internal/modules/user/dto"
)

type AvatarHandler interface {
	CreateUploadURL(c *gin.Context)
	Confirm(c *gin.Context)
	Delete(c *gin.Context)
}
type avatarHandler struct {
	a   *app.App
	log *zap.Logger
	svc user.Service
}

func NewAvatarHandler(a *app.App) (AvatarHandler, error) {
	if a == nil || a.Logger() == nil {
		return nil, fmt.Errorf("app/logger required")
	}
	if a.DB() == nil {
		return nil, fmt.Errorf("gorm db required for user avatar module")
	}

	repo := user.NewRepository(a.DB())
	return &avatarHandler{
		a:   a,
		log: a.Logger(),
		svc: user.NewService(repo),
	}, nil
}

var allowedCT = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// @Summary      Create presigned avatar upload URL
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body  body  dto.CreateUploadURLRequest  true  "payload"
// @Success      200  {object}  dto.CreateUploadURLResponse
// @Failure      400  {object}  common.Envelope
// @Failure      401  {object}  common.Envelope
// @Failure      503  {object}  common.Envelope
// @Router       /users/avatar/upload-url [post]
// @Security     BearerAuth
func (h *avatarHandler) CreateUploadURL(c *gin.Context) {
	var req userdto.CreateUploadURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.WriteError(c, 400, common.ValidationFrom(err))
		return
	}
	ext, ok := allowedCT[req.ContentType]
	if !ok {
		common.WriteError(c, 400, &common.APIError{Code: common.CodeBadRequest, Message: "unsupported content type"})
		return
	}

	uidRaw, ok := c.Get("user_id")
	if !ok {
		common.WriteError(c, http.StatusUnauthorized, &common.APIError{Code: common.CodeUnauthorized, Message: "missing user id"})
		return
	}
	userID, ok := uidRaw.(string)
	if !ok || userID == "" {
		common.WriteError(c, http.StatusUnauthorized, &common.APIError{Code: common.CodeUnauthorized, Message: "invalid user id"})
		return
	}
	ctx := c.Request.Context()

	// tahmin edilemez key
	suffix := uuid.NewString()[:8]
	key := fmt.Sprintf("users/%s/avatar-%s%s", userID, suffix, ext)

	url, err := h.a.Storage().PresignedPut(ctx, key, req.ContentType, 10*time.Minute)
	if err != nil {
		h.log.Error("presigned put", zap.Error(err))
		common.WriteError(c, 500, &common.APIError{Code: common.CodeInternal, Message: "cannot create upload url"})
		return
	}

	// tek-kullanımlık kayıt
	if h.a.Redis() != nil {
		if err := h.a.Redis().Set(ctx, "upload:"+key, userID, 10*time.Minute).Err(); err != nil {
			h.log.Warn("avatar.upload.redis_set", zap.Error(err))
		}
	}

	resp := userdto.CreateUploadURLResponse{
		Key:          key,
		UploadURL:    url,
		ExpiresInSec: 600,
	}
	common.WriteOK(c, 200, &resp, nil)
}

// @Summary      Confirm uploaded avatar object
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body  body  dto.ConfirmAvatarRequest  true  "payload"
// @Success      200  {object} dto.ConfirmAvatarResponse
// @Failure      400  {object} common.Envelope
// @Failure      401  {object} common.Envelope
// @Failure      500  {object} common.Envelope
// @Failure      503  {object} common.Envelope
// @Router       /users/avatar/confirm [post]
func (h *avatarHandler) Confirm(c *gin.Context) {
	var req userdto.ConfirmAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.WriteError(c, 400, common.ValidationFrom(err))
		return
	}

	uidRaw, ok := c.Get("user_id")
	if !ok {
		common.WriteError(c, http.StatusUnauthorized, &common.APIError{Code: common.CodeUnauthorized, Message: "missing user id"})
		return
	}
	uidStr, ok := uidRaw.(string)
	if !ok {
		common.WriteError(c, http.StatusUnauthorized, &common.APIError{Code: common.CodeUnauthorized, Message: "invalid user id"})
		return
	}
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		common.WriteError(c, http.StatusBadRequest, &common.APIError{Code: common.CodeBadRequest, Message: "invalid user id"})
		return
	}

	ctx := c.Request.Context()

	// sahiplik + tek kullanım
	if h.a.Redis() != nil {
		val, err := h.a.Redis().Get(ctx, "upload:"+req.Key).Result()
		if err != nil || val != uid.String() {
			common.WriteError(c, 403, &common.APIError{Code: common.CodeForbidden, Message: "invalid or expired upload key"})
			return
		}
	} else {
		common.WriteError(c, http.StatusServiceUnavailable, &common.APIError{Code: common.CodeInternal, Message: "temporary unavailable"})
		return
	}

	// obje var mı?
	exists, _ := h.a.Storage().Exists(ctx, req.Key)
	if !exists {
		common.WriteError(c, 400, &common.APIError{Code: common.CodeBadRequest, Message: "object not found"})
		return
	}

	// boyut & tip kontrolü (5MB)
	meta, err := h.a.Storage().Stat(ctx, req.Key)
	if err != nil {
		common.WriteError(c, 400, &common.APIError{Code: common.CodeBadRequest, Message: "cannot stat object"})
		return
	}
	if meta.Size > 5*1024*1024 {
		_ = h.a.Storage().Delete(ctx, req.Key)
		common.WriteError(c, 400, &common.APIError{Code: common.CodeBadRequest, Message: "file too large (max 5MB)"})
		return
	}
	switch meta.ContentType {
	case "image/jpeg", "image/png", "image/webp":
	default:
		_ = h.a.Storage().Delete(ctx, req.Key)
		common.WriteError(c, 400, &common.APIError{Code: common.CodeBadRequest, Message: "unsupported content type"})
		return
	}

	// DB’de avatar_key yaz
	if err := h.svc.UpdateAvatarKey(ctx, uid, &req.Key); err != nil {
		common.WriteError(c, 500, &common.APIError{Code: common.CodeInternal, Message: "cannot update avatar"})
		return
	}

	// tek-kullanımlık kaydı tüket
	if h.a.Redis() != nil {
		_ = h.a.Redis().Del(ctx, "upload:"+req.Key).Err()
	}

	// response (public bucket ise hesaplanmış URL)
	final := h.a.Storage().PublicURL(req.Key)
	common.WriteOK(c, 200, &userdto.ConfirmAvatarResponse{AvatarURL: final}, nil)
}

// @Summary      Delete current user avatar
// @Tags         users
// @Produce      json
// @Success      200  {object}  common.Envelope
// @Failure      400  {object}  common.Envelope
// @Failure      401  {object}  common.Envelope
// @Failure      500  {object}  common.Envelope
// @Failure      503  {object}  common.Envelope
// @Router       /users/avatar [delete]
// @Security     BearerAuth
func (h *avatarHandler) Delete(c *gin.Context) {
	uidRaw, ok := c.Get("user_id")
	if !ok {
		common.WriteError(c, http.StatusUnauthorized, &common.APIError{Code: common.CodeUnauthorized, Message: "missing user id"})
		return
	}
	uidStr, ok := uidRaw.(string)
	if !ok {
		common.WriteError(c, http.StatusUnauthorized, &common.APIError{Code: common.CodeUnauthorized, Message: "invalid user id"})
		return
	}
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		common.WriteError(c, http.StatusBadRequest, &common.APIError{Code: common.CodeBadRequest, Message: "invalid user id"})
		return
	}

	ctx := c.Request.Context()
	u, err := h.svc.GetByID(ctx, uid)
	if err != nil {
		if err == user.ErrNotFound {
			common.WriteError(c, http.StatusNotFound, &common.APIError{Code: common.CodeNotFound, Message: "user not found"})
			return
		}
		h.log.Error("avatar.delete.get_user", zap.Error(err))
		common.WriteError(c, http.StatusInternalServerError, &common.APIError{Code: common.CodeInternal, Message: "cannot load user"})
		return
	}

	if u.AvatarURL != nil && *u.AvatarURL != "" {
		if err := h.a.Storage().Delete(ctx, *u.AvatarURL); err != nil {
			h.log.Warn("avatar.delete.storage", zap.Error(err))
		}
	}

	if err := h.svc.DeleteAvatar(ctx, uid); err != nil {
		common.WriteError(c, 500, &common.APIError{Code: common.CodeInternal, Message: "cannot delete avatar"})
		return
	}
	common.WriteOK[any](c, http.StatusOK, nil, nil)
}
