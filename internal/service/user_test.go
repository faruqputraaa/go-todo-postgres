package service

import (
	"context"
	"encoding/json"
	"errors"
	"go-todo/internal/entity"
	mock_cache "go-todo/test/mock/pkg/cache"
	mock_token "go-todo/test/mock/pkg/token"
	mock_repository "go-todo/test/mock/repository"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func setupUserService(t *testing.T) (*gomock.Controller, UserService, *mock_repository.MockUserRepository, *mock_cache.MockCacheable, *mock_token.MockTokenUseCase) {
	ctrl := gomock.NewController(t)
	mockRepo := mock_repository.NewMockUserRepository(ctrl)
	mockCache := mock_cache.NewMockCacheable(ctrl)
	mockToken := mock_token.NewMockTokenUseCase(ctrl)
	service := NewUserService(mockRepo, mockToken, mockCache)
	return ctrl, service, mockRepo, mockCache, mockToken
}

// Kasus uji untuk FindAll

func TestUserService_FindAll_CacheHit(t *testing.T) {
	ctrl, service, _, mockCache, _ := setupUserService(t)
	defer ctrl.Finish()

	ctx := context.Background()
	expectedUsers := []entity.User{{ID: 1, Username: "user1"}, {ID: 2, Username: "user2"}}
	cachedData, _ := json.Marshal(expectedUsers)

	mockCache.EXPECT().Get("pengguna:semua").Return(string(cachedData), nil)

	users, err := service.FindAll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedUsers, users)
}

func TestUserService_FindAll_CacheMiss(t *testing.T) {
	ctrl, service, mockRepo, mockCache, _ := setupUserService(t)
	defer ctrl.Finish()

	ctx := context.Background()
	expectedUsers := []entity.User{{ID: 1, Username: "user1"}, {ID: 2, Username: "user2"}}

	mockCache.EXPECT().Get("pengguna:semua").Return("", nil)
	mockRepo.EXPECT().FindAll(ctx).Return(expectedUsers, nil)

	marshalledData, _ := json.Marshal(expectedUsers)
	mockCache.EXPECT().Set("pengguna:semua", marshalledData, 5*time.Minute).Return(nil)

	users, err := service.FindAll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedUsers, users)
}

func TestUserService_FindAll_CacheError(t *testing.T) {
	ctrl, service, mockRepo, mockCache, _ := setupUserService(t)
	defer ctrl.Finish()

	ctx := context.Background()
	expectedUsers := []entity.User{{ID: 1, Username: "user1"}, {ID: 2, Username: "user2"}}

	mockCache.EXPECT().Get("pengguna:semua").Return("", errors.New("cache error"))
	mockRepo.EXPECT().FindAll(ctx).Return(expectedUsers, nil)

	marshalledData, _ := json.Marshal(expectedUsers)
	mockCache.EXPECT().Set("pengguna:semua", marshalledData, 5*time.Minute).Return(nil)

	users, err := service.FindAll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedUsers, users)
}

// Kasus uji untuk Login

func TestUserService_Login_ValidCredentials(t *testing.T) {
	ctrl, service, mockRepo, _, mockToken := setupUserService(t)
	defer ctrl.Finish()

	ctx := context.Background()
	username, password := "user1", "password"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := entity.User{ID: 1, Username: username, Password: string(hashedPassword), Role: "user"}

	mockRepo.EXPECT().FindByUsername(ctx, username).Return(&user, nil)
	mockToken.EXPECT().GenerateAccessToken(gomock.Any()).Return("mockToken", nil)

	token, err := service.Login(ctx, username, password)
	assert.NoError(t, err)
	assert.Equal(t, "mockToken", token)
}

func TestUserService_Login_InvalidCredentials(t *testing.T) {
	ctrl, service, mockRepo, _, _ := setupUserService(t)
	defer ctrl.Finish()

	ctx := context.Background()
	username, password := "user1", "wrongPassword"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := entity.User{ID: 1, Username: username, Password: string(hashedPassword)}

	mockRepo.EXPECT().FindByUsername(ctx, username).Return(&user, nil)

	_, err := service.Login(ctx, username, password)
	assert.ErrorIs(t, err, ErrKredensialTidakValid)
}

// Kasus uji untuk CreateUser

func TestUserService_CreateUser_NewUsername(t *testing.T) {
	ctrl, service, mockRepo, mockCache, _ := setupUserService(t)
	defer ctrl.Finish()

	ctx := context.Background()
	user := &entity.User{Username: "newUser", Password: "password"}
	expectedUser := &entity.User{ID: 1, Username: "newUser"}

	mockRepo.EXPECT().FindByUsername(ctx, user.Username).Return(nil, errors.New("not found"))
	mockRepo.EXPECT().Create(ctx, user).Return(expectedUser, nil)
	mockCache.EXPECT().Delete("pengguna:semua").Return(nil)

	createdUser, err := service.CreateUser(ctx, user)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, createdUser)
}

func TestUserService_CreateUser_ExistingUsername(t *testing.T) {
	ctrl, service, mockRepo, _, _ := setupUserService(t)
	defer ctrl.Finish()

	ctx := context.Background()
	user := &entity.User{Username: "existingUser"}

	mockRepo.EXPECT().FindByUsername(ctx, user.Username).Return(user, nil)

	_, err := service.CreateUser(ctx, user)
	assert.ErrorIs(t, err, ErrUsernameSudahAda)
}

// Kasus uji untuk UpdateUser

func TestUserService_UpdateUser_ValidID(t *testing.T) {
	ctrl, service, mockRepo, mockCache, _ := setupUserService(t)
	defer ctrl.Finish()

	ctx := context.Background()
	existingUser := &entity.User{ID: 1, Username: "user1", FullName: "Old Name", Role: "user"}
	updateData := &entity.User{ID: 1, FullName: "New Name"} // Hanya memperbarui FullName

	// Mengharapkan repository untuk mengambil pengguna yang ada
	mockRepo.EXPECT().FindByID(ctx, existingUser.ID).Return(existingUser, nil)

	// Mengharapkan repository untuk memperbarui pengguna dengan hanya field yang tidak kosong
	expectedUpdatedUser := &entity.User{ID: 1, Username: "user1", FullName: "New Name", Role: "user"}
	mockRepo.EXPECT().Update(ctx, expectedUpdatedUser).Return(expectedUpdatedUser, nil)

	// Mengharapkan cache dihapus setelah pembaruan
	mockCache.EXPECT().Delete("pengguna:semua").Return(nil)

	// Menjalankan fungsi pembaruan service
	result, err := service.UpdateUser(ctx, updateData)
	assert.NoError(t, err)
	assert.Equal(t, "New Name", result.FullName)
	assert.Equal(t, "user1", result.Username) // Memastikan field yang tidak berubah tidak terpengaruh
	assert.Equal(t, "user", result.Role)
}

func TestUserService_UpdateUser_PartialUpdate(t *testing.T) {
	ctrl, service, mockRepo, mockCache, _ := setupUserService(t)
	defer ctrl.Finish()

	ctx := context.Background()
	existingUser := &entity.User{ID: 2, Username: "user2", FullName: "Old Name", Role: "user"}
	updateData := &entity.User{ID: 2, Username: "newUser2"} // Hanya memperbarui Username

	// Mock pengambilan pengguna yang ada
	mockRepo.EXPECT().FindByID(ctx, existingUser.ID).Return(existingUser, nil)

	// Mengharapkan hanya Username yang diperbarui di repository
	expectedUpdatedUser := &entity.User{ID: 2, Username: "newUser2", FullName: "Old Name", Role: "user"}
	mockRepo.EXPECT().Update(ctx, expectedUpdatedUser).Return(expectedUpdatedUser, nil)

	// Mengharapkan cache dihapus
	mockCache.EXPECT().Delete("pengguna:semua").Return(nil)

	// Menjalankan fungsi pembaruan service
	result, err := service.UpdateUser(ctx, updateData)
	assert.NoError(t, err)
	assert.Equal(t, "newUser2", result.Username)
	assert.Equal(t, "Old Name", result.FullName) // Memastikan field lain tetap sama
	assert.Equal(t, "user", result.Role)
}

func TestUserService_UpdateUser_InvalidID(t *testing.T) {
	ctrl, service, _, _, _ := setupUserService(t)
	defer ctrl.Finish()

	ctx := context.Background()
	updateData := &entity.User{ID: -1, FullName: "Invalid ID Test"}

	// Menjalankan fungsi pembaruan service dengan ID tidak valid
	_, err := service.UpdateUser(ctx, updateData)
	assert.Error(t, err)
	assert.Equal(t, "ID pengguna tidak valid", err.Error())
}

func TestUserService_UpdateUser_NotFound(t *testing.T) {
	ctrl, service, mockRepo, _, _ := setupUserService(t)
	defer ctrl.Finish()

	ctx := context.Background()
	updateData := &entity.User{ID: 3, FullName: "Non-Existent User"}

	// Mengharapkan repository mengembalikan error yang menunjukkan pengguna tidak ditemukan
	mockRepo.EXPECT().FindByID(ctx, updateData.ID).Return(nil, ErrPenggunaTidakDitemukan)

	// Menjalankan fungsi pembaruan service
	_, err := service.UpdateUser(ctx, updateData)
	assert.ErrorIs(t, err, ErrPenggunaTidakDitemukan)
}

// Kasus uji untuk DeleteUser

func TestUserService_DeleteUser_ValidID(t *testing.T) {
	ctrl, service, mockRepo, mockCache, _ := setupUserService(t)
	defer ctrl.Finish()

	ctx := context.Background()
	userID := int64(1)

	mockRepo.EXPECT().FindByID(ctx, userID).Return(&entity.User{ID: userID}, nil)
	mockRepo.EXPECT().Delete(ctx, userID).Return(nil)
	mockCache.EXPECT().Delete("pengguna:semua").Return(nil)

	err := service.DeleteUser(ctx, userID)
	assert.NoError(t, err)
}

func TestUserService_DeleteUser_InvalidID(t *testing.T) {
	ctrl, service, _, _, _ := setupUserService(t)
	defer ctrl.Finish()

	ctx := context.Background()
	userID := int64(-1)

	err := service.DeleteUser(ctx, userID)
	assert.Error(t, err)
	assert.Equal(t, "ID pengguna tidak valid", err.Error())
}
