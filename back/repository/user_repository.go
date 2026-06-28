package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/yanatoritakuma/trade-analyzer/back/domain"
	"github.com/yanatoritakuma/trade-analyzer/back/domain/user"
	"github.com/yanatoritakuma/trade-analyzer/back/model"
)

// UserRepositoryImpl は user.UserRepository のGORM実装。
type UserRepositoryImpl struct {
	db *gorm.DB
}

func NewUserRepositoryImpl(db *gorm.DB) user.UserRepository {
	return &UserRepositoryImpl{db: db}
}

func toUserEntity(m *model.User) *user.User {
	return &user.User{
		ID:           m.ID,
		Email:        m.Email,
		Name:         m.Name,
		PasswordHash: m.PasswordHash,
		Role:         user.Role(m.Role),
		IsActive:     m.IsActive,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func (r *UserRepositoryImpl) FindByID(id uint) (*user.User, error) {
	var m model.User
	if err := r.db.First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toUserEntity(&m), nil
}

func (r *UserRepositoryImpl) FindByEmail(email string) (*user.User, error) {
	var m model.User
	if err := r.db.Where("email = ?", email).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toUserEntity(&m), nil
}

func (r *UserRepositoryImpl) FindFirstAdmin() (*user.User, error) {
	var m model.User
	if err := r.db.Where("role = ?", "admin").Order("id ASC").First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toUserEntity(&m), nil
}

func (r *UserRepositoryImpl) FindAll() ([]*user.User, error) {
	var ms []model.User
	if err := r.db.Order("created_at DESC").Find(&ms).Error; err != nil {
		return nil, err
	}
	users := make([]*user.User, 0, len(ms))
	for i := range ms {
		users = append(users, toUserEntity(&ms[i]))
	}
	return users, nil
}

func (r *UserRepositoryImpl) Save(u *user.User) error {
	m := model.User{
		Email:        u.Email,
		Name:         u.Name,
		PasswordHash: u.PasswordHash,
		Role:         string(u.Role),
		IsActive:     u.IsActive,
	}
	if err := r.db.Create(&m).Error; err != nil {
		return err
	}
	u.ID = m.ID
	return nil
}

func (r *UserRepositoryImpl) Update(u *user.User) error {
	return r.db.Model(&model.User{}).Where("id = ?", u.ID).Updates(map[string]interface{}{
		"name":          u.Name,
		"password_hash": u.PasswordHash,
		"is_active":     u.IsActive,
	}).Error
}

func (r *UserRepositoryImpl) Delete(id uint) error {
	return r.db.Unscoped().Delete(&model.User{}, id).Error
}
