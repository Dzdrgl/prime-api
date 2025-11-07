package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"githup.com/ademdizdar/prime-api/internal/app"
	"githup.com/ademdizdar/prime-api/internal/common"
	"githup.com/ademdizdar/prime-api/internal/modules/auth/dto"
	"githup.com/ademdizdar/prime-api/internal/modules/user"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Handler interface {
	Login(c *gin.Context)
	Refresh(c *gin.Context)
	Logout(c *gin.Context)
}

type handler struct {
	log        *zap.Logger
	users      user.Repository
	sessions   SessionRepository
	audit      AuditRepository
	jwt        *common.JWTMaker
	accessTTL  time.Duration
	refreshTTL time.Duration
	pepper     []byte
	bl         Blacklist // for access revoke (optional)
}

func NewHandler(a *app.App, jwt *common.JWTMaker, accessTTL, refreshTTL time.Duration, pepper []byte, bl Blacklist) (Handler, error) {
	return &handler{
		log:        a.Logger(),
		users:      user.NewGormRepository(a.DB()),
		sessions:   NewSessionRepository(a.DB()),
		audit:      NewAuditRepository(a.DB()),
		jwt:        jwt,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		pepper:     pepper,
		bl:         bl,
	}, nil
}

// @Summary      Login
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  dto.LoginRequest  true  "payload"
// @Success      200  {object} apidoc.LoginEnvelope
// @Failure      400  {object} apidoc.LoginEnvelope
// @Failure      401  {object} apidoc.LoginEnvelope
// @Router       /auth/login [post]
func (h *handler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.WriteError(c, http.StatusBadRequest, common.ValidationFrom(err))
		return
	}
	u, err := h.users.GetByEmail(c.Request.Context(), strings.ToLower(req.Email))
	if err != nil || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		common.WriteError(c, http.StatusUnauthorized, &common.APIError{Code: common.CodeBadRequest, Message: "invalid credentials"})
		return
	}

	// Tokens
	access, _, accessExp, err := h.jwt.GenerateAccess(u.ID.String(), u.Email, h.accessTTL)
	if err != nil {
		h.log.Error("jwt.generate", zap.Error(err))
		common.WriteError(c, 500, &common.APIError{Code: common.CodeInternal, Message: "cannot create token"})
		return
	}
	rawRefresh := uuid.NewString() + "." + uuid.NewString() // random opaque
	refreshHash := common.HashRefreshToken(rawRefresh, h.pepper)
	now := time.Now().UTC()
	refreshExp := now.Add(h.refreshTTL)

	// Session create
	s := &UserSession{
		ID:          uuid.New(),
		UserID:      u.ID,
		DeviceID:    req.DeviceID,
		RefreshHash: refreshHash,
		IssuedAt:    now,
		ExpiresAt:   refreshExp,
		LastUsedAt:  now,
		IP:          c.ClientIP(),
		UserAgent:   c.Request.UserAgent(),
		IsRevoked:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := h.sessions.Create(c.Request.Context(), s); err != nil {
		h.log.Error("session.create", zap.Error(err))
		common.WriteError(c, 500, &common.APIError{Code: common.CodeInternal, Message: "cannot create session"})
		return
	}

	// Audit
	_ = h.audit.Add(c.Request.Context(), &AuditEvent{
		UserID: &u.ID, SessionID: &s.ID, Event: "login", IP: c.ClientIP(), UserAgent: c.Request.UserAgent(), CreatedAt: now,
	})

	resp := dto.LoginResponse{
		AccessToken:  access,
		AccessExp:    accessExp.Unix(),
		RefreshToken: rawRefresh,
		RefreshExp:   refreshExp.Unix(),
		SessionID:    s.ID.String(),
	}
	common.WriteOK(c, http.StatusOK, &resp, nil)
}

// @Summary      Refresh token (rotate)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  dto.RefreshRequest  true  "payload"
// @Success      200  {object} apidoc.RefreshEnvelope
// @Failure      400  {object} apidoc.RefreshEnvelope
// @Failure      401  {object} apidoc.RefreshEnvelope
// @Router       /auth/refresh [post]
func (h *handler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.WriteError(c, 400, common.ValidationFrom(err))
		return
	}
	now := time.Now().UTC()
	hash := common.HashRefreshToken(req.RefreshToken, h.pepper)

	s, err := h.sessions.GetActiveByHash(c.Request.Context(), hash, req.DeviceID)
	if err != nil {
		// reuse detection: eğer hash aktif kayıtta yok ama geride "prev" olarak tespit edilirse tüm oturumları revoke et (örn. ayrı bir sorgu yazılabilir).
		// Basit yaklaşım: audit ve generic unauthorized.
		_ = h.audit.Add(c.Request.Context(), &AuditEvent{Event: "refresh_failed", IP: c.ClientIP(), UserAgent: c.Request.UserAgent(), CreatedAt: now})
		common.WriteError(c, 401, &common.APIError{Code: common.CodeBadRequest, Message: "invalid refresh"})
		return
	}

	// Rotate: eski hash -> prev, yeni hash üret
	newRaw := uuid.NewString() + "." + uuid.NewString()
	newHash := common.HashRefreshToken(newRaw, h.pepper)
	if err := h.sessions.UpdateRotate(c.Request.Context(), s.ID, newHash, now, hash); err != nil {
		h.log.Error("session.rotate", zap.Error(err))
		common.WriteError(c, 500, &common.APIError{Code: common.CodeInternal, Message: "cannot rotate session"})
		return
	}

	// Yeni access
	uuidStr := s.UserID.String()
	// Uyarı: email’i almak için kullanıcıyı çekiyoruz (alternatif: session’a cacheleyebilirsin)
	// (Prod'da email'i claims yerine sadece uid ile tutmak tercih edilebilir)
	access, _, accessExp, err := h.jwt.GenerateAccess(uuidStr, "", h.accessTTL)
	if err != nil {
		common.WriteError(c, 500, &common.APIError{Code: common.CodeInternal, Message: "cannot create token"})
		return
	}
	newExp := now.Add(h.refreshTTL)

	// Audit
	_ = h.audit.Add(c.Request.Context(), &AuditEvent{
		UserID: &s.UserID, SessionID: &s.ID, Event: "refresh", IP: c.ClientIP(), UserAgent: c.Request.UserAgent(), CreatedAt: now,
	})

	resp := dto.RefreshResponse{
		AccessToken:  access,
		AccessExp:    accessExp.Unix(),
		RefreshToken: newRaw,
		RefreshExp:   newExp.Unix(),
		SessionID:    s.ID.String(),
	}
	common.WriteOK(c, 200, &resp, nil)
}

func (h *handler) Logout(c *gin.Context) {
	var req dto.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.WriteError(c, 400, common.ValidationFrom(err))
		return
	}
	sid, err := uuid.Parse(req.SessionID)
	if err != nil {
		common.WriteError(c, 400, &common.APIError{Code: common.CodeBadRequest, Message: "invalid session id"})
		return
	}

	if err := h.sessions.RevokeByID(c.Request.Context(), sid); err != nil {
		common.WriteError(c, 500, &common.APIError{Code: common.CodeInternal, Message: "cannot revoke session"})
		return
	}

	// Access token'ı (varsa) blacklist’e ekle (opsiyonel)
	if h.bl != nil {
		if jtiVal, ok := c.Get("access_jti"); ok {
			if jti, _ := jtiVal.(string); jti != "" {
				h.bl.Add(jti, h.accessTTL)
			}
		}
	}

	now := time.Now().UTC()
	_ = h.audit.Add(c.Request.Context(), &AuditEvent{
		Event: "logout", SessionID: &sid, IP: c.ClientIP(), UserAgent: c.Request.UserAgent(), CreatedAt: now,
	})

	common.WriteOK[any](c, 200, nil, nil)
}
