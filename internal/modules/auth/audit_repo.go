package auth

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditRepository interface {
	Add(ctx context.Context, ev *AuditEvent) error
}

type gormAuditRepo struct{ db *gorm.DB }

func NewAuditRepository(db *gorm.DB) AuditRepository { return &gormAuditRepo{db: db} }

func (r *gormAuditRepo) Add(ctx context.Context, ev *AuditEvent) error {
	if ev.ID == uuid.Nil {
		ev.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(ev).Error
}
