package user

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"githup.com/ademdizdar/prime-api/internal/app"
	"githup.com/ademdizdar/prime-api/internal/common"
	"githup.com/ademdizdar/prime-api/internal/modules/user/dto"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// PROTOKOL: Handler arayüzü
type Handler interface {
	List(c *gin.Context)
	Get(c *gin.Context)
	Create(c *gin.Context)
	CreateAvatarUploadURL(c *gin.Context)
	ConfirmAvatar(c *gin.Context)
	DeleteAvatar(c *gin.Context)
}

type handler struct {
	a       *app.App
	log     *zap.Logger
	svc     Service
	storage app.ObjectStorage
}

func NewHandler(a *app.App) (Handler, error) {
	if a == nil || a.Logger() == nil {
		return nil, fmt.Errorf("app/logger required")
	}
	if a.DB() == nil {
		return nil, fmt.Errorf("gorm db required for user module")
	}

	repo := NewGormRepository(a.DB())
	return &handler{
		a:       a,
		log:     a.Logger(),
		svc:     NewService(repo),
		storage: a.Storage(),
	}, nil
}

// @Summary      List users
// @Tags         users
// @Produce      json
// @Param        page   query   int   false  "page (default 1)"
// @Param        limit  query   int   false  "limit (default 20)"
// @Success      200  {object} apidoc.UserListEnvelope
// @Failure      400  {object} apidoc.UserListEnvelope
// @Failure      500  {object} apidoc.UserListEnvelope
// @Router       /users [get]
// @Security     BearerAuth
func (h *handler) List(c *gin.Context) {
	var pq common.PageQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		common.WriteError(c, http.StatusBadRequest, &common.APIError{
			Code:    common.CodeBadRequest,
			Message: "invalid pagination",
		})
		return
	}
	if pq.Page <= 0 {
		pq.Page = 1
	}
	if pq.Limit <= 0 {
		pq.Limit = 20
	}
	offset := (pq.Page - 1) * pq.Limit

	entities, err := h.svc.List(c.Request.Context(), pq.Limit, offset)
	if err != nil {
		h.log.Error("user.list", zap.Error(err))
		common.WriteError(c, http.StatusInternalServerError, &common.APIError{
			Code:    common.CodeInternal,
			Message: "cannot list users",
		})
		return
	}

	resp := ToUserListResponse(entities)
	meta := common.PageMeta{Page: pq.Page, Limit: pq.Limit}
	common.WriteOK(c, http.StatusOK, &resp, meta)
}

// @Summary      Get user by ID
// @Tags         users
// @Produce      json
// @Param        id   path   string  true  "UUID"
// @Success      200  {object} apidoc.UserEnvelope
// @Failure      400  {object} apidoc.UserEnvelope
// @Failure      404  {object} apidoc.UserEnvelope
// @Failure      500  {object} apidoc.UserEnvelope
// @Router       /users/{id} [get]
// @Security     BearerAuth
func (h *handler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		common.WriteError(c, http.StatusBadRequest, &common.APIError{
			Code:    common.CodeBadRequest,
			Message: "invalid user id",
		})
		return
	}

	u, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == ErrNotFound {
			common.WriteError(c, http.StatusNotFound, &common.APIError{
				Code:    common.CodeNotFound,
				Message: "user not found",
			})
			return
		}
		h.log.Error("user.get", zap.Error(err))
		common.WriteError(c, http.StatusInternalServerError, &common.APIError{
			Code:    common.CodeInternal,
			Message: "cannot get user",
		})
		return
	}

	resp := ToUserResponse(u)
	common.WriteOK(c, http.StatusOK, &resp, nil)
}

// @Summary      Create user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body  body  dto.CreateUserRequest  true  "payload"
// @Success      201  {object} apidoc.UserEnvelope
// @Failure      400  {object} apidoc.UserEnvelope
// @Failure      409  {object} apidoc.UserEnvelope
// @Failure      500  {object} apidoc.UserEnvelope
// @Router       /users [post]
func (h *handler) Create(c *gin.Context) {
	// 1) Hem JSON hem multipart destekleyelim:
	ct := strings.ToLower(c.ContentType())
	if ct == "" {
		ct = c.ContentType()
	}
	ct = strings.ToLower(ct)

	var req dto.CreateUserRequest
	var bindErr error

	if strings.HasPrefix(ct, "multipart/") {
		bindErr = c.ShouldBind(&req) // form binding
	} else {
		bindErr = c.ShouldBindJSON(&req)
	}
	if bindErr != nil {
		common.WriteError(c, http.StatusBadRequest, common.ValidationFrom(bindErr))
		return
	}

	// 2) kullanıcıyı önce resimsiz oluştur
	pw := strings.TrimSpace(req.Password)
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		h.log.Error("user.hash_password", zap.Error(err))
		common.WriteError(c, http.StatusInternalServerError, &common.APIError{
			Code:    common.CodeInternal,
			Message: "cannot create user",
		})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	fullName := strings.TrimSpace(req.FullName)
	u, err := h.svc.Create(c.Request.Context(), email, fullName, string(hash)) // AvatarURL nil
	if err != nil {
		if errors.Is(err, ErrEmailExists) {
			common.WriteError(c, http.StatusConflict, &common.APIError{
				Code:    common.CodeConflict,
				Message: "email already exists",
			})
			return
		}
		h.log.Error("user.create", zap.Error(err))
		common.WriteError(c, http.StatusInternalServerError, &common.APIError{
			Code:    common.CodeInternal,
			Message: "cannot create user",
		})
		return
	}

	// 3) avatar varsa MinIO’ya yükle, kullanıcıyı güncelle
	if c.GetHeader("X-USE-PICSUM") == "1" {
		key := fmt.Sprintf("users/%s/avatar.jpg", u.ID.String())
		// max 5MB
		pub, signed, err := FetchAndUploadFromURL(c.Request.Context(), h.a.Storage(), "https://picsum.photos/200/300", key, 5*1024*1024)
		if err != nil {
			h.log.Warn("picsum fetch/upload failed", zap.Error(err))
		} else {
			var final *string
			if pub != "" {
				final = &pub
			} else if signed != "" {
				final = &signed
			}
			if final != nil {
				_ = h.svc.UpdateAvatar(c.Request.Context(), u.ID, final)
				u.AvatarURL = final
			}
		}
	} else if req.AvatarDataURL != nil && *req.AvatarDataURL != "" {
		mime, bytes, err := common.ParseImageDataURL(*req.AvatarDataURL, 5*1024*1024) // 5MB
		if err != nil {
			h.log.Warn("invalid avatar data url", zap.Error(err))
		} else {
			ext := ".jpg"
			switch mime {
			case "image/png":
				ext = ".png"
			case "image/webp":
				ext = ".webp"
			}
			key := fmt.Sprintf("users/%s/avatar%s", u.ID.String(), ext)
			r, size := app.BytesReader(bytes)
			publicURL, signedURL, err := h.a.Storage().Put(c.Request.Context(), key, r, size, mime)
			if err != nil {
				h.log.Warn("avatar upload failed", zap.Error(err))
			} else {
				var final *string
				if publicURL != "" {
					final = &publicURL
				} else if signedURL != "" {
					final = &signedURL
				}
				if final != nil {
					if err := h.svc.UpdateAvatar(c.Request.Context(), u.ID, final); err == nil {
						u.AvatarURL = final
					}
				}
			}
		}
	}

	// 4) response
	resp := ToUserResponse(u)
	common.WriteOK(c, http.StatusCreated, &resp, nil)
}

// +++ EK: izinli mime'lar
var allowedImageCT = map[string]string{
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
func (h *handler) CreateAvatarUploadURL(c *gin.Context) {
	if h.storage == nil {
		common.WriteError(c, 503, &common.APIError{Code: common.CodeInternal, Message: "storage unavailable"})
		return
	}

	var req dto.CreateUploadURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.WriteError(c, 400, common.ValidationFrom(err))
		return
	}
	ext, ok := allowedImageCT[req.ContentType]
	if !ok {
		common.WriteError(c, 400, &common.APIError{Code: common.CodeBadRequest, Message: "unsupported content type"})
		return
	}
	if req.Ext != "" {
		ext = req.Ext
	} // (isteğe bağlı override)

	uidStr, ok := c.Get("user_id") // JWT middleware set eder
	if !ok {
		common.WriteError(c, 401, &common.APIError{Code: common.CodeBadRequest, Message: "missing user id"})
		return
	}
	userID, _ := uidStr.(string)

	key := fmt.Sprintf("users/%s/avatar%s", userID, ext)
	uploadURL, err := h.storage.PresignedPut(c.Request.Context(), key, req.ContentType, 10*time.Minute)
	if err != nil {
		h.log.Error("presigned put", zap.Error(err))
		common.WriteError(c, 500, &common.APIError{Code: common.CodeInternal, Message: "cannot create upload url"})
		return
	}

	resp := dto.CreateUploadURLResponse{
		Key:          key,
		UploadURL:    uploadURL,
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
// @Security     BearerAuth
func (h *handler) ConfirmAvatar(c *gin.Context) {
	if h.storage == nil {
		common.WriteError(c, 503, &common.APIError{Code: common.CodeInternal, Message: "storage unavailable"})
		return
	}

	var req dto.ConfirmAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.WriteError(c, 400, common.ValidationFrom(err))
		return
	}

	exists, err := h.storage.Exists(c.Request.Context(), req.Key)
	if err != nil || !exists {
		common.WriteError(c, 400, &common.APIError{Code: common.CodeBadRequest, Message: "object not found"})
		return
	}

	uidStr, ok := c.Get("user_id")
	if !ok {
		common.WriteError(c, 401, &common.APIError{Code: common.CodeBadRequest, Message: "missing user id"})
		return
	}
	uid, err := uuid.Parse(uidStr.(string))
	if err != nil {
		common.WriteError(c, 400, &common.APIError{Code: common.CodeBadRequest, Message: "invalid user id"})
		return
	}

	// Public bucket ise kalıcı URL; private ise nil bırak (isteğe göre)
	final := h.storage.PublicURL(req.Key)
	if err := h.svc.UpdateAvatar(c.Request.Context(), uid, final); err != nil {
		common.WriteError(c, 500, &common.APIError{Code: common.CodeInternal, Message: "cannot update avatar"})
		return
	}

	common.WriteOK(c, 200, &dto.ConfirmAvatarResponse{AvatarURL: final}, nil)
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
func (h *handler) DeleteAvatar(c *gin.Context) {
	if h.storage == nil {
		common.WriteError(c, 503, &common.APIError{Code: common.CodeInternal, Message: "storage unavailable"})
		return
	}

	uidStr, ok := c.Get("user_id")
	if !ok {
		common.WriteError(c, 401, &common.APIError{Code: common.CodeBadRequest, Message: "missing user id"})
		return
	}
	uid, err := uuid.Parse(uidStr.(string))
	if err != nil {
		common.WriteError(c, 400, &common.APIError{Code: common.CodeBadRequest, Message: "invalid user id"})
		return
	}

	ctx := c.Request.Context()
	exts := []string{".jpg", ".png", ".webp"}
	for _, ext := range exts {
		key := fmt.Sprintf("users/%s/avatar%s", uid.String(), ext)
		if err := h.storage.Delete(ctx, key); err != nil {
			h.log.Warn("avatar.delete", zap.String("key", key), zap.Error(err))
		}
	}
	if err := h.svc.DeleteAvatar(ctx, uid); err != nil {
		common.WriteError(c, 500, &common.APIError{Code: common.CodeInternal, Message: "cannot delete avatar"})
		return
	}
	common.WriteOK[any](c, http.StatusOK, nil, nil)
}
