package service

import (
	"context"
	"encoding/json"
	"errors"
	"go-todo/internal/entity"
	"go-todo/internal/repository"
	"go-todo/pkg/cache"
	"time"
)

type TodoService interface {
	FindAll(ctx context.Context) ([]entity.Todo, error)
	Create(ctx context.Context, todo entity.Todo) (entity.Todo, error)
	Update(ctx context.Context, id int64, todo entity.Todo) (entity.Todo, error)
	Delete(ctx context.Context, id int64) error
}

type todoService struct {
	todoRepository repository.TodoRepository
	cacheable      cache.Cacheable
}

// NewTodoService membuat instance baru dari TodoService
func NewTodoService(
	todoRepository repository.TodoRepository,
	cacheable cache.Cacheable,
) TodoService {
	return &todoService{todoRepository, cacheable}
}

// FindAll mengambil semua data todo, dengan menggunakan caching untuk meningkatkan performa
func (s *todoService) FindAll(ctx context.Context) ([]entity.Todo, error) {
	keyFindAll := "go-todo-api:todos:find-all"

	// Mencoba mengambil data dari cache
	data, err := s.cacheable.Get(keyFindAll)
	if err != nil {
		// Mengembalikan langsung jika terjadi error cache
		return nil, err
	}

	if data != "" {
		var result []entity.Todo
		if err := json.Unmarshal([]byte(data), &result); err == nil {
			return result, nil
		}
	}

	// Jika cache miss atau gagal unmarshalling, mengambil dari repository
	result, err := s.todoRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	// Menyimpan cache hanya jika repository mengembalikan hasil yang tidak nil
	if result != nil {
		dataMarshalled, err := json.Marshal(result)
		if err == nil {
			// Menyimpan data dengan masa kadaluarsa 5 menit; mengabaikan error dari Set
			_ = s.cacheable.Set(keyFindAll, dataMarshalled, 5*time.Minute)
		}
	}

	return result, nil
}

// Create menambahkan todo baru
func (s *todoService) Create(ctx context.Context, todo entity.Todo) (entity.Todo, error) {
	// Menyimpan data todo baru ke dalam repository
	createdTodo, err := s.todoRepository.Create(ctx, todo)
	if err != nil {
		return entity.Todo{}, errors.New("gagal menambahkan todo")
	}

	// Menghapus cache untuk menjaga konsistensi data
	s.cacheable.Delete("go-todo-api:todos:find-all")
	return createdTodo, nil
}


// Update memperbarui data todo berdasarkan ID
func (s *todoService) Update(ctx context.Context, id int64, todo entity.Todo) (entity.Todo, error) {
	// Mengecek apakah todo yang ingin diperbarui ada di database
	existingTodo, err := s.todoRepository.FindByID(ctx, id)
	if err != nil {
		return entity.Todo{}, errors.New("todo tidak ditemukan")
	}

	// Memperbarui field dari todo yang ada hanya jika field baru tidak kosong
	if todo.Title != "" {
		existingTodo.Title = todo.Title
	}
	if todo.Content != "" {
		existingTodo.Content = todo.Content
	}
	if !todo.DueDate.IsZero() {
		existingTodo.DueDate = todo.DueDate
	}
	// Completed field should be updated directly as it is a boolean
	existingTodo.Completed = todo.Completed

	// Menyimpan data yang telah diperbarui ke dalam repository
	updatedTodo, err := s.todoRepository.Update(ctx, *existingTodo)
	if err != nil {
		return entity.Todo{}, errors.New("gagal memperbarui todo")
	}

	// Menghapus cache untuk menjaga konsistensi data
	s.cacheable.Delete("go-todo-api:todos:find-all")
	return updatedTodo, nil
}


// Delete menghapus todo berdasarkan ID
func (s *todoService) Delete(ctx context.Context, id int64) error {
	// Mengecek apakah todo yang ingin dihapus ada di database
	_, err := s.todoRepository.FindByID(ctx, id)
	if err != nil {
		return errors.New("todo tidak ditemukan")
	}

	// Menghapus todo dari repository
	err = s.todoRepository.Delete(ctx, id)
	if err != nil {
		return errors.New("gagal menghapus todo")
	}

	// Menghapus cache untuk menjaga konsistensi data
	s.cacheable.Delete("go-todo-api:todos:find-all")
	return nil
}
