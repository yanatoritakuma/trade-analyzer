package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/yanatoritakuma/trade-analyzer/back/domain"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/invitation"
	"github.com/yanatoritakuma/trade-analyzer/back/model"
)

// InvitationCodeRepositoryImpl は invitation.InvitationCodeRepository のGORM実装。
type InvitationCodeRepositoryImpl struct {
	db *gorm.DB
}

func NewInvitationCodeRepositoryImpl(db *gorm.DB) invitation.InvitationCodeRepository {
	return &InvitationCodeRepositoryImpl{db: db}
}

func toInvitationEntity(m *model.InvitationCode) *invitation.InvitationCode {
	return &invitation.InvitationCode{
		ID:        m.ID,
		Code:      m.Code,
		CreatedBy: m.CreatedBy,
		UsedBy:    m.UsedBy,
		ExpiresAt: m.ExpiresAt,
		UsedAt:    m.UsedAt,
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt,
	}
}

func (r *InvitationCodeRepositoryImpl) FindByCode(code string) (*invitation.InvitationCode, error) {
	var m model.InvitationCode
	if err := r.db.Where("code = ?", code).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrInvalidCode
		}
		return nil, err
	}
	return toInvitationEntity(&m), nil
}

func (r *InvitationCodeRepositoryImpl) FindByID(id uint) (*invitation.InvitationCode, error) {
	var m model.InvitationCode
	if err := r.db.First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toInvitationEntity(&m), nil
}

func (r *InvitationCodeRepositoryImpl) FindAll() ([]*invitation.InvitationCode, error) {
	var ms []model.InvitationCode
	if err := r.db.Order("created_at DESC").Find(&ms).Error; err != nil {
		return nil, err
	}
	out := make([]*invitation.InvitationCode, 0, len(ms))
	for i := range ms {
		out = append(out, toInvitationEntity(&ms[i]))
	}
	return out, nil
}

func (r *InvitationCodeRepositoryImpl) CountValid() (int64, error) {
	var count int64
	err := r.db.Model(&model.InvitationCode{}).
		Where("is_active = ? AND used_by IS NULL AND expires_at >= ?", true, time.Now()).
		Count(&count).Error
	return count, err
}

func (r *InvitationCodeRepositoryImpl) Save(i *invitation.InvitationCode) error {
	m := model.InvitationCode{
		Code:      i.Code,
		CreatedBy: i.CreatedBy,
		UsedBy:    i.UsedBy,
		ExpiresAt: i.ExpiresAt,
		UsedAt:    i.UsedAt,
		IsActive:  i.IsActive,
	}
	if err := r.db.Create(&m).Error; err != nil {
		return err
	}
	i.ID = m.ID
	i.CreatedAt = m.CreatedAt
	return nil
}

func (r *InvitationCodeRepositoryImpl) Update(i *invitation.InvitationCode) error {
	return r.db.Model(&model.InvitationCode{}).Where("id = ?", i.ID).Updates(map[string]interface{}{
		"used_by":   i.UsedBy,
		"used_at":   i.UsedAt,
		"is_active": i.IsActive,
	}).Error
}
