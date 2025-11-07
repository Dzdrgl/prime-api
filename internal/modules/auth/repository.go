package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionRevoked  = errors.New("session revoked")
)

type SessionRepository interface {
	Create(ctx context.Context, s *UserSession) error
	GetActiveByHash(ctx context.Context, refreshHash, deviceID string) (*UserSession, error)
	UpdateRotate(ctx context.Context, id uuid.UUID, newHash string, now time.Time, prev string) error
	RevokeByID(ctx context.Context, id uuid.UUID) error
	RevokeAllByUser(ctx context.Context, userID uuid.UUID) error
	//GetByPrevHash(ctx context.Context, prevHash, deviceID string) (*UserSession, error)
}

type gormSessionRepo struct{ db *gorm.DB }

func NewSessionRepository(db *gorm.DB) SessionRepository { return &gormSessionRepo{db: db} }

func (r *gormSessionRepo) Create(ctx context.Context, s *UserSession) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *gormSessionRepo) GetActiveByHash(ctx context.Context, refreshHash, deviceID string) (*UserSession, error) {
	var s UserSession
	err := r.db.WithContext(ctx).
		Where("refresh_hash = ? AND device_id = ? AND is_revoked = false AND expires_at > NOW()", refreshHash, deviceID).
		First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *gormSessionRepo) UpdateRotate(ctx context.Context, id uuid.UUID, newHash string, now time.Time, prev string) error {
	return r.db.WithContext(ctx).Model(&UserSession{}).
		Where("id = ? AND is_revoked = false", id).
		Updates(map[string]any{
			"prev_refresh_hash": prev,
			"refresh_hash":      newHash,
			"last_used_at":      now,
		}).Error
}

func (r *gormSessionRepo) RevokeByID(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&UserSession{}).
		Where("id = ?", id).
		Update("is_revoked", true).Error
}

func (r *gormSessionRepo) RevokeAllByUser(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&UserSession{}).
		Where("user_id = ?", userID).
		Update("is_revoked", true).Error
}
