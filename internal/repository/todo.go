package repository

import (
	"context"
	"go-todo/internal/entity"

	"gorm.io/gorm"
)

// TodoRepository mendefinisikan operasi CRUD untuk entity Todo.
type TodoRepository interface {
	FindAll(ctx context.Context) ([]entity.Todo, error)
	FindByID(ctx context.Context, id int64) (*entity.Todo, error)
	Create(ctx context.Context, todo entity.Todo) (entity.Todo, error)
	Update(ctx context.Context, todo entity.Todo) (entity.Todo, error)
	Delete(ctx context.Context, id int64) error
}

type todoRepository struct {
	db *gorm.DB
}

// NewTodoRepository menginisialisasi repository Todo baru.
func NewTodoRepository(db *gorm.DB) TodoRepository {
	return &todoRepository{db}
}

// FindAll mengambil semua todo dari database.
func (r *todoRepository) FindAll(ctx context.Context) ([]entity.Todo, error) {
	var todos []entity.Todo
	if err := r.db.WithContext(ctx).Find(&todos).Error; err != nil {
		return nil, err
	}
	return todos, nil
}

// FindByID mengambil satu todo berdasarkan ID dari database.
func (r *todoRepository) FindByID(ctx context.Context, id int64) (*entity.Todo, error) {
	todo := new(entity.Todo)
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(todo).Error; err != nil {
		return nil, err
	}
	return todo, nil
}

// Create menambahkan todo baru ke dalam database.
func (r *todoRepository) Create(ctx context.Context, todo entity.Todo) (entity.Todo, error) {
	if err := r.db.WithContext(ctx).Create(&todo).Error; err != nil {
		return entity.Todo{}, err
	}
	return todo, nil
}

// Update memperbarui todo yang ada di database.
func (r *todoRepository) Update(ctx context.Context, todo entity.Todo) (entity.Todo, error) {
	// Menghapus kondisi `Where` yang eksplisit
	if err := r.db.WithContext(ctx).Model(&todo).
		Select("Title", "Content", "DueDate", "Completed", "UserID").
		Updates(todo).Error; err != nil {
		return entity.Todo{}, err
	}
	return todo, nil
}

// Delete menghapus todo berdasarkan ID dari database.
func (r *todoRepository) Delete(ctx context.Context, id int64) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.Todo{}).Error; err != nil {
		return err
	}
	return nil
}
