package service

import (
	"context"
	"encoding/json"
	"errors"
	"go-todo/internal/entity"
	mock_cache "go-todo/test/mock/pkg/cache"       // Mock untuk cache
	mock_repository "go-todo/test/mock/repository" // Mock untuk repository
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestTodoService_FindAll_CacheHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := mock_cache.NewMockCacheable(ctrl)
	mockRepo := mock_repository.NewMockTodoRepository(ctrl)
	service := NewTodoService(mockRepo, mockCache)

	ctx := context.Background()
	expectedTodos := []entity.Todo{{ID: 1, Title: "Test Todo 1"}, {ID: 2, Title: "Test Todo 2"}}
	cachedData, _ := json.Marshal(expectedTodos)

	mockCache.EXPECT().Get("go-todo-api:todos:find-all").Return(string(cachedData), nil)

	todos, err := service.FindAll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedTodos, todos)
}

func TestTodoService_FindAll_CacheMissAndRepoHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := mock_cache.NewMockCacheable(ctrl)
	mockRepo := mock_repository.NewMockTodoRepository(ctrl)
	service := NewTodoService(mockRepo, mockCache)

	ctx := context.Background()
	expectedTodos := []entity.Todo{{ID: 1, Title: "Test Todo 1"}, {ID: 2, Title: "Test Todo 2"}}

	mockCache.EXPECT().Get("go-todo-api:todos:find-all").Return("", nil)
	mockRepo.EXPECT().FindAll(ctx).Return(expectedTodos, nil)
	mockCache.EXPECT().Set("go-todo-api:todos:find-all", gomock.Any(), 5*time.Minute).Return(nil)

	todos, err := service.FindAll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedTodos, todos)
}

func TestTodoService_FindAll_CacheMissAndRepoNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := mock_cache.NewMockCacheable(ctrl)
	mockRepo := mock_repository.NewMockTodoRepository(ctrl)
	service := NewTodoService(mockRepo, mockCache)

	ctx := context.Background()

	// Expectations: Cache miss, repository returns nil, no Set to cache
	mockCache.EXPECT().Get("go-todo-api:todos:find-all").Return("", nil)
	mockRepo.EXPECT().FindAll(ctx).Return(nil, nil)

	todos, err := service.FindAll(ctx)
	assert.NoError(t, err)
	assert.Nil(t, todos) // Expect todos to be nil, as repository returned nil
}

func TestTodoService_FindAll_CacheError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := mock_cache.NewMockCacheable(ctrl)
	mockRepo := mock_repository.NewMockTodoRepository(ctrl)
	service := NewTodoService(mockRepo, mockCache)

	ctx := context.Background()

	// Expect cache to return an error and the repository should not be called
	mockCache.EXPECT().Get("go-todo-api:todos:find-all").Return("", errors.New("cache error"))
	// No call expected to `repository.FindAll`

	_, err := service.FindAll(ctx)
	assert.Error(t, err)
	assert.Equal(t, "cache error", err.Error())
}

func TestTodoService_FindAll_RepoErrorOnCacheMiss(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := mock_cache.NewMockCacheable(ctrl)
	mockRepo := mock_repository.NewMockTodoRepository(ctrl)
	service := NewTodoService(mockRepo, mockCache)

	ctx := context.Background()

	mockCache.EXPECT().Get("go-todo-api:todos:find-all").Return("", nil)
	mockRepo.EXPECT().FindAll(ctx).Return(nil, errors.New("repository error"))

	_, err := service.FindAll(ctx)
	assert.Error(t, err)
	assert.Equal(t, "repository error", err.Error())
}

func TestTodoService_FindAll_ErrorDuringCacheSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := mock_cache.NewMockCacheable(ctrl)
	mockRepo := mock_repository.NewMockTodoRepository(ctrl)
	service := NewTodoService(mockRepo, mockCache)

	ctx := context.Background()
	expectedTodos := []entity.Todo{{ID: 1, Title: "Test Todo 1"}, {ID: 2, Title: "Test Todo 2"}}

	// Ekspektasi: Cache miss, repository hit, dan gagal melakukan Set ke cache
	mockCache.EXPECT().Get("go-todo-api:todos:find-all").Return("", nil)
	mockRepo.EXPECT().FindAll(ctx).Return(expectedTodos, nil)
	mockCache.EXPECT().Set("go-todo-api:todos:find-all", gomock.Any(), 5*time.Minute).Return(errors.New("cache set error"))

	todos, err := service.FindAll(ctx)
	assert.NoError(t, err)                // Even with cache set error, FindAll should not return an error
	assert.Equal(t, expectedTodos, todos) // Expected data should still be returned
}

func TestTodoService_FindAll_ErrorDuringUnmarshal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := mock_cache.NewMockCacheable(ctrl)
	mockRepo := mock_repository.NewMockTodoRepository(ctrl)
	service := NewTodoService(mockRepo, mockCache)

	ctx := context.Background()
	expectedTodos := []entity.Todo{
		{ID: 1, Title: "Test Todo 1"},
		{ID: 2, Title: "Test Todo 2"},
	}

	// Cache returns invalid data, causing unmarshal to fail, which triggers a repository call.
	mockCache.EXPECT().Get("go-todo-api:todos:find-all").Return("invalid data", nil)
	mockRepo.EXPECT().FindAll(ctx).Return(expectedTodos, nil)

	// Expect a Set call to cache the data from repository after fetching from repository
	marshalledData, _ := json.Marshal(expectedTodos)
	mockCache.EXPECT().Set("go-todo-api:todos:find-all", marshalledData, 5*time.Minute).Return(nil)

	todos, err := service.FindAll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedTodos, todos)
}

func TestTodoService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_repository.NewMockTodoRepository(ctrl)
	mockCache := mock_cache.NewMockCacheable(ctrl)
	service := NewTodoService(mockRepo, mockCache)

	ctx := context.Background()
	newTodo := entity.Todo{Title: "New Todo"}
	expectedTodo := entity.Todo{ID: 1, Title: "New Todo"}

	// Test case 1: Successful creation
	mockRepo.EXPECT().Create(ctx, newTodo).Return(expectedTodo, nil)
	mockCache.EXPECT().Delete("go-todo-api:todos:find-all").Return(nil)

	createdTodo, err := service.Create(ctx, newTodo)
	assert.NoError(t, err)
	assert.Equal(t, expectedTodo, createdTodo)

	// Test case 2: Repository error on creation
	mockRepo.EXPECT().Create(ctx, newTodo).Return(entity.Todo{}, errors.New("repository error"))

	_, err = service.Create(ctx, newTodo)
	assert.Error(t, err)
	assert.Equal(t, "gagal menambahkan todo", err.Error())
}

func TestTodoService_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_repository.NewMockTodoRepository(ctrl)
	mockCache := mock_cache.NewMockCacheable(ctrl)
	service := NewTodoService(mockRepo, mockCache)

	ctx := context.Background()
	existingTodo := entity.Todo{ID: 1, Title: "Old Title", Content: "Old Content", Completed: false}
	updateData := entity.Todo{Title: "Updated Title"} // Only updating the Title field
	expectedUpdatedTodo := entity.Todo{ID: 1, Title: "Updated Title", Content: "Old Content", Completed: false}

	// Test case 1: Successful update with partial fields
	mockRepo.EXPECT().FindByID(ctx, int64(1)).Return(&existingTodo, nil)
	mockRepo.EXPECT().Update(ctx, expectedUpdatedTodo).Return(expectedUpdatedTodo, nil)
	mockCache.EXPECT().Delete("go-todo-api:todos:find-all").Return(nil)

	updatedTodo, err := service.Update(ctx, 1, updateData)
	assert.NoError(t, err)
	assert.Equal(t, expectedUpdatedTodo, updatedTodo)

	// Test case 2: Todo not found in repository
	mockRepo.EXPECT().FindByID(ctx, int64(1)).Return(nil, errors.New("todo tidak ditemukan"))

	_, err = service.Update(ctx, 1, updateData)
	assert.Error(t, err)
	assert.Equal(t, "todo tidak ditemukan", err.Error())

	// Test case 3: Repository error during update
	mockRepo.EXPECT().FindByID(ctx, int64(1)).Return(&existingTodo, nil)
	mockRepo.EXPECT().Update(ctx, expectedUpdatedTodo).Return(entity.Todo{}, errors.New("repository error"))

	_, err = service.Update(ctx, 1, updateData)
	assert.Error(t, err)
	assert.Equal(t, "gagal memperbarui todo", err.Error())
}


func TestTodoService_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_repository.NewMockTodoRepository(ctrl)
	mockCache := mock_cache.NewMockCacheable(ctrl)
	service := NewTodoService(mockRepo, mockCache)

	ctx := context.Background()

	// Test case 1: Successful deletion
	mockRepo.EXPECT().FindByID(ctx, int64(1)).Return(&entity.Todo{ID: 1}, nil)
	mockRepo.EXPECT().Delete(ctx, int64(1)).Return(nil)
	mockCache.EXPECT().Delete("go-todo-api:todos:find-all").Return(nil)

	err := service.Delete(ctx, 1)
	assert.NoError(t, err)

	// Test case 2: Todo not found
	mockRepo.EXPECT().FindByID(ctx, int64(1)).Return(nil, errors.New("todo tidak ditemukan"))

	err = service.Delete(ctx, 1)
	assert.Error(t, err)
	assert.Equal(t, "todo tidak ditemukan", err.Error())

	// Test case 3: Repository error on delete
	mockRepo.EXPECT().FindByID(ctx, int64(1)).Return(&entity.Todo{ID: 1}, nil)
	mockRepo.EXPECT().Delete(ctx, int64(1)).Return(errors.New("repository error"))

	err = service.Delete(ctx, 1)
	assert.Error(t, err)
	assert.Equal(t, "gagal menghapus todo", err.Error())
}
