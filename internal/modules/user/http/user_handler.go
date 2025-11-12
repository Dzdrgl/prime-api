package http

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
	user "githup.com/ademdizdar/prime-api/internal/modules/user"
	"githup.com/ademdizdar/prime-api/internal/modules/user/dto"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// PROTOKOL: Handler arayüzü
type UserHandler interface {
	List(c *gin.Context)
	Get(c *gin.Context)
	Create(c *gin.Context)
}

type userHandler struct {
	a       *app.App
	log     *zap.Logger
	svc     user.Service
	storage app.ObjectStorage
}

func NewUserHandler(a *app.App) (UserHandler, error) {
	if a == nil || a.Logger() == nil {
		return nil, fmt.Errorf("app/logger required")
	}
	if a.DB() == nil {
		return nil, fmt.Errorf("gorm db required for user module")
	}

	repo := user.NewRepository(a.DB())
	return &userHandler{
		a:       a,
		log:     a.Logger(),
		svc:     user.NewService(repo),
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
func (h *userHandler) List(c *gin.Context) {
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

	resp := user.ToUserListResponse(entities)
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
func (h *userHandler) Get(c *gin.Context) {
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
		if err == user.ErrNotFound {
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

	resp := user.ToUserResponse(u)
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
func (h *userHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	// 1) JSON bind
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.WriteError(c, http.StatusBadRequest, common.ValidationFrom(err))
		return
	}

	// 2) kullanıcıyı önce resimsiz oluştur

	pw := strings.TrimSpace(req.Password)
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		h.log.Error("user.hash_password", zap.Error(err))
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	fullName := strings.TrimSpace(req.FullName)
	u, err := h.svc.Create(ctx, email, fullName, string(hash))
	if err != nil {
		if errors.Is(err, user.ErrEmailExists) {
			common.WriteError(c, http.StatusConflict, &common.APIError{Code: common.CodeConflict, Message: "email already exists"})
			return
		}
		h.log.Error("user.create", zap.Error(err))
		return
	}

	// 3) avatar varsa MinIO’ya yükle, kullanıcıyı güncelle
	var upload *dto.CreateUploadURLResponse
	if ct := c.GetHeader("X-RETURN-AVATAR-PUT"); ct != "" {
		ext := ".jpg"
		switch ct {
		case "image/png":
			ext = ".png"
		case "image/webp":
			ext = ".webp"
		case "image/jpeg":
			ext = ".jpg"
		default:
			common.WriteError(c, 400, &common.APIError{Code: common.CodeBadRequest, Message: "unsupported content type for X-RETURN-AVATAR-PUT"})
			return
		}

		suffix := uuid.NewString()[:8]
		key := fmt.Sprintf("users/%s/avatar-%s%s", u.ID.String(), suffix, ext)

		url, err := h.a.Storage().PresignedPut(ctx, key, ct, 10*time.Minute)
		if err == nil {
			if h.a.Redis() != nil {
				_ = h.a.Redis().Set(ctx, "upload:"+key, u.ID.String(), 5*time.Minute).Err()
			}
			upload = &dto.CreateUploadURLResponse{
				Key:          key,
				UploadURL:    url,
				ExpiresInSec: 600,
			}
		} else {
			h.log.Warn("presigned put create failed", zap.Error(err))
		}
	}
	// 4) response
	resp := dto.UserCreatedResponse{
		User:         user.ToUserResponse(u),
		AvatarUpload: upload,
	}
	common.WriteOK(c, http.StatusCreated, &resp, nil)
}

// +++ EK: izinli mime'lar
