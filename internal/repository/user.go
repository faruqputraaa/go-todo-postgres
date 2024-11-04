package repository

import (
	"context"
	"errors"
	"fmt"
	"go-todo/internal/entity"

	"gorm.io/gorm"
)

// UserRepository mendefinisikan operasi database untuk entitas User.
type UserRepository interface {
	FindAll(ctx context.Context) ([]entity.User, error)                        
	FindByID(ctx context.Context, id int64) (*entity.User, error)              
	FindByUsername(ctx context.Context, username string) (*entity.User, error) 
	Create(ctx context.Context, user *entity.User) (*entity.User, error)       
	Update(ctx context.Context, user *entity.User) (*entity.User, error)       
	Delete(ctx context.Context, id int64) error                                
}

var (
	ErrPenggunaTidakDitemukan  = errors.New("pengguna tidak ditemukan")
	ErrUsernameTelahDigunakan  = errors.New("username telah digunakan")
	ErrDatabaseError           = errors.New("terjadi kesalahan pada database")
)

type userRepository struct {
	db *gorm.DB // gorm.DB untuk operasi database
}

// NewUserRepository inisialisasi UserRepository baru.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db}
}

// FindAll mengambil semua pengguna dari database.
func (r *userRepository) FindAll(ctx context.Context) ([]entity.User, error) {
	users := make([]entity.User, 0)
	if err := r.db.WithContext(ctx).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return users, nil
}

// FindByID mencari pengguna berdasarkan ID.
func (r *userRepository) FindByID(ctx context.Context, id int64) (*entity.User, error) {
	user := new(entity.User)
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPenggunaTidakDitemukan
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return user, nil
}

// FindByUsername mencari pengguna berdasarkan username.
func (r *userRepository) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	user := new(entity.User)
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPenggunaTidakDitemukan
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return user, nil
}

// Create menambahkan pengguna baru ke database.
func (r *userRepository) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return user, nil
}

// Update memperbarui data pengguna.
func (r *userRepository) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	// Siapkan map untuk field yang akan diupdate
	updates := map[string]interface{}{
		"username": user.Username,
		"full_name": user.FullName,
		"role":      user.Role,
		"password":  user.Password,
	}

	// Hapus field yang kosong atau nil
	for key, value := range updates {
		if value == "" {
			delete(updates, key)
		}
	}

	// Lakukan update
	result := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", user.ID).
		Updates(updates)

	if result.Error != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, ErrPenggunaTidakDitemukan
	}

	// Ambil data terbaru setelah update
	updatedUser, err := r.FindByID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

// Delete menghapus pengguna berdasarkan ID.
func (r *userRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.User{})
	if result.Error != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseError, result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrPenggunaTidakDitemukan
	}
	return nil
}
