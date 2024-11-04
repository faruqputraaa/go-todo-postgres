package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-todo/internal/entity"
	"go-todo/internal/repository"
	"go-todo/pkg/cache"
	"go-todo/pkg/token"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrKredensialTidakValid = errors.New("username atau password tidak valid")
	ErrPenggunaTidakDitemukan = errors.New("pengguna tidak ditemukan")
	ErrServerInternal = errors.New("terjadi kesalahan pada server")
	ErrUsernameSudahAda = errors.New("username sudah digunakan")
)

type UserService interface {
	FindAll(ctx context.Context) ([]entity.User, error)
	Login(ctx context.Context, username, password string) (string, error)
	CreateUser(ctx context.Context, user *entity.User) (*entity.User, error)
	UpdateUser(ctx context.Context, user *entity.User) (*entity.User, error)
	DeleteUser(ctx context.Context, id int64) error
}

type userService struct {
	userRepository repository.UserRepository
	tokenUseCase   token.TokenUseCase
	cacheable      cache.Cacheable
}

// NewUserService membuat instance baru dari UserService
func NewUserService(
	userRepository repository.UserRepository,
	tokenUseCase token.TokenUseCase,
	cacheable cache.Cacheable,
) UserService {
	return &userService{
		userRepository: userRepository,
		tokenUseCase:  tokenUseCase,
		cacheable:     cacheable,
	}
}

// FindAll mengambil semua data pengguna dengan implementasi cache
func (s *userService) FindAll(ctx context.Context) ([]entity.User, error) {
	const cacheKey = "pengguna:semua"
	
	// Coba ambil dari cache terlebih dahulu
	cachedData, err := s.cacheable.Get(cacheKey)
	if err == nil && cachedData != "" {
		var users []entity.User
		if err := json.Unmarshal([]byte(cachedData), &users); err == nil {
			return users, nil
		}
	}

	// Jika tidak ada di cache, ambil dari repository
	users, err := s.userRepository.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data pengguna: %w", err)
	}

	// Perbarui cache
	if marshalledData, err := json.Marshal(users); err == nil {
		if err := s.cacheable.Set(cacheKey, marshalledData, 5*time.Minute); err != nil {
			// Catat error cache tapi jangan gagalkan request
			fmt.Printf("kesalahan menyimpan cache: %v\n", err)
		}
	}

	return users, nil
}

// Login memproses autentikasi pengguna
func (s *userService) Login(ctx context.Context, username, password string) (string, error) {
	user, err := s.userRepository.FindByUsername(ctx, username)
	if err != nil {
		return "", ErrKredensialTidakValid
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", ErrKredensialTidakValid
	}

	claims := token.JwtCustomClaims{
		Username: user.Username,
		Role:     user.Role,
		FullName: user.FullName,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "aplikasi-todo",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token, err := s.tokenUseCase.GenerateAccessToken(claims)
	if err != nil {
		return "", fmt.Errorf("gagal membuat token: %w", err)
	}

	return token, nil
}

// CreateUser menambahkan pengguna baru
func (s *userService) CreateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	// Cek apakah username sudah ada
	if _, err := s.userRepository.FindByUsername(ctx, user.Username); err == nil {
		return nil, ErrUsernameSudahAda
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("gagal mengenkripsi password: %w", err)
	}
	user.Password = string(hashedPassword)

	// Simpan pengguna ke database
	createdUser, err := s.userRepository.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat pengguna: %w", err)
	}

	// Hapus cache agar data konsisten
	s.cacheable.Delete("pengguna:semua")

	return createdUser, nil
}

// UpdateUser memperbarui data pengguna
func (s *userService) UpdateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	if user.ID <= 0 {
		return nil, errors.New("ID pengguna tidak valid")
	}

	existingUser, err := s.userRepository.FindByID(ctx, user.ID)
	if err != nil {
		return nil, ErrPenggunaTidakDitemukan
	}

	// Update fields yang tidak kosong
	if user.FullName != "" {
		existingUser.FullName = user.FullName
	}
	if user.Role != "" {
		existingUser.Role = user.Role
	}
	if user.Username != "" {
		existingUser.Username = user.Username
	}

	// Khusus untuk password, hanya update jika ada nilai baru
	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("gagal mengenkripsi password: %w", err)
		}
		existingUser.Password = string(hashedPassword)
	}

	// Update pengguna
	updatedUser, err := s.userRepository.Update(ctx, existingUser)
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui pengguna: %w", err)
	}

	// Hapus cache
	s.cacheable.Delete("pengguna:semua")

	return updatedUser, nil
}

// DeleteUser menghapus data pengguna
func (s *userService) DeleteUser(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("ID pengguna tidak valid")
	}

	// Cek apakah pengguna ada
	if _, err := s.userRepository.FindByID(ctx, id); err != nil {
		return ErrPenggunaTidakDitemukan
	}

	if err := s.userRepository.Delete(ctx, id); err != nil {
		return fmt.Errorf("gagal menghapus pengguna: %w", err)
	}

	// Hapus cache
	s.cacheable.Delete("pengguna:semua")

	return nil
}
