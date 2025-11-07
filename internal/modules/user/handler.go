package user

import (
	"fmt"
	"net/http"
	"strings"

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
}

type handler struct {
	log *zap.Logger
	svc Service
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
		log: a.Logger(),
		svc: NewService(repo),
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
// @Security     BearerAuth
func (h *handler) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Gin içindeki validator.v10 tetiklenir → ayrıntılı field hataları
		common.WriteError(c, http.StatusBadRequest, common.ValidationFrom(err))
		return
	}

	pw := strings.TrimSpace(req.Password)
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		h.log.Error("user.create.hash", zap.Error(err))
		common.WriteError(c, http.StatusInternalServerError, &common.APIError{
			Code:    common.CodeInternal,
			Message: "cannot hash password",
		})
		return
	}

	u, err := h.svc.Create(c.Request.Context(), req.Email, req.FullName, string(hash))
	if err != nil {
		if err == ErrEmailExists {
			common.WriteError(c, http.StatusConflict, &common.APIError{
				Code:    common.CodeConflict,
				Message: "email already registered",
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

	resp := ToUserResponse(u)
	common.WriteOK(c, http.StatusCreated, &resp, nil)
}
