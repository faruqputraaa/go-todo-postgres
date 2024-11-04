package repository

import (
	"context"
	"errors"
	"go-todo/internal/entity"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupMockDB mengatur database mock untuk pengujian
func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	// Ganti ke MySQL Dialector
	dialector := mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	})

	gormDB, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	assert.NoError(t, err)

	return gormDB, mock
}

// TestTodoRepository_FindAll menguji fungsi FindAll dari TodoRepository
func TestTodoRepository_FindAll(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.DB()

	repo := NewTodoRepository(db)

	rows := sqlmock.NewRows([]string{"id", "title", "content"}).
		AddRow(1, "Test Todo 1", "Content 1").
		AddRow(2, "Test Todo 2", "Content 2")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `todos`")).
		WillReturnRows(rows)

	todos, err := repo.FindAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, todos, 2)
	assert.Equal(t, "Test Todo 1", todos[0].Title)
	assert.Equal(t, "Content 1", todos[0].Content)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestTodoRepository_FindByID menguji fungsi FindByID dari TodoRepository
func TestTodoRepository_FindByID(t *testing.T) {
	db, mock := setupMockDB(t)
	sqlDB, err := db.DB()
	assert.NoError(t, err)
	defer sqlDB.Close()

	repo := NewTodoRepository(db)

	row := sqlmock.NewRows([]string{"id", "title", "content"}).
		AddRow(1, "Test Todo", "Content")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `todos` WHERE id = ? ORDER BY `todos`.`id` LIMIT ?")).
		WithArgs(1, 1).
		WillReturnRows(row)

	todo, err := repo.FindByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, "Test Todo", todo.Title)
	assert.Equal(t, "Content", todo.Content)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestTodoRepository_Create menguji fungsi Create dari TodoRepository
func TestTodoRepository_Create(t *testing.T) {
	db, mock := setupMockDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewTodoRepository(db)

	// Tambahkan semua field yang diperlukan
	todo := entity.Todo{
		Title:     "New Todo",
		Content:   "New Content",
		DueDate:   time.Now(),
		Completed: false,
		UserID:    1,
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `todos` (`title`,`content`,`due_date`,`completed`,`user_id`) VALUES (?,?,?,?,?)")).
		WithArgs(todo.Title, todo.Content, todo.DueDate, todo.Completed, todo.UserID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	createdTodo, err := repo.Create(context.Background(), todo)
	assert.NoError(t, err)
	assert.Equal(t, "New Todo", createdTodo.Title)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestTodoRepository_FindAll_Error menguji fungsi FindAll dari TodoRepository ketika terjadi error
func TestTodoRepository_FindAll_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.DB()

	repo := NewTodoRepository(db)

	// Simulasi error saat query `FindAll`
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `todos`")).WillReturnError(errors.New("database error"))

	_, err := repo.FindAll(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestTodoRepository_FindByID_Error menguji fungsi FindByID dari TodoRepository ketika terjadi error
func TestTodoRepository_FindByID_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	sqlDB, err := db.DB()
	assert.NoError(t, err)
	defer sqlDB.Close()

	repo := NewTodoRepository(db)

	// Simulasi error saat query `FindByID`
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `todos` WHERE id = ? ORDER BY `todos`.`id` LIMIT ?")).
		WithArgs(1, 1).
		WillReturnError(errors.New("database error"))

	_, err = repo.FindByID(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestTodoRepository_Create_Error menguji fungsi Create dari TodoRepository ketika terjadi error
func TestTodoRepository_Create_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewTodoRepository(db)

	todo := entity.Todo{
		Title:     "New Todo",
		Content:   "New Content",
		DueDate:   time.Now(),
		Completed: false,
		UserID:    1,
	}

	// Simulasi error saat `Create`
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `todos` (`title`,`content`,`due_date`,`completed`,`user_id`) VALUES (?,?,?,?,?)")).
		WithArgs(todo.Title, todo.Content, todo.DueDate, todo.Completed, todo.UserID).
		WillReturnError(errors.New("insert error"))
	mock.ExpectRollback()

	_, err := repo.Create(context.Background(), todo)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insert error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestTodoRepository_Update menguji fungsi Update dari TodoRepository
func TestTodoRepository_Update(t *testing.T) {
	db, mock := setupMockDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewTodoRepository(db)

	// Define `todo` with fields to be updated, excluding `id` from the SET clause
	todo := entity.Todo{
		ID:        1,
		Title:     "Updated Todo",
		Content:   "Updated Content",
		DueDate:   time.Now(),
		Completed: true,
		UserID:    1,
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `todos` SET `title`=?,`content`=?,`due_date`=?,`completed`=?,`user_id`=? WHERE `id` = ?")).
		WithArgs(todo.Title, todo.Content, todo.DueDate, todo.Completed, todo.UserID, todo.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	updatedTodo, err := repo.Update(context.Background(), todo)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Todo", updatedTodo.Title)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestTodoRepository_Update_Error menguji fungsi Update dari TodoRepository ketika terjadi error
func TestTodoRepository_Update_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewTodoRepository(db)

	todo := entity.Todo{
		ID:        1,
		Title:     "Updated Todo",
		Content:   "Updated Content",
		DueDate:   time.Now(),
		Completed: true,
		UserID:    1,
	}

	// Simulate an error during the `Update` operation
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `todos` SET `title`=?,`content`=?,`due_date`=?,`completed`=?,`user_id`=? WHERE `id` = ?")).
		WithArgs(todo.Title, todo.Content, todo.DueDate, todo.Completed, todo.UserID, todo.ID).
		WillReturnError(errors.New("update error"))
	mock.ExpectRollback()

	_, err := repo.Update(context.Background(), todo)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestTodoRepository_Delete menguji fungsi Delete dari TodoRepository
func TestTodoRepository_Delete(t *testing.T) {
	db, mock := setupMockDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewTodoRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `todos` WHERE id = ?")).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1)) // RowsAffected = 1, LastInsertId = 0
	mock.ExpectCommit()

	err := repo.Delete(context.Background(), 1)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestTodoRepository_Delete_Error menguji fungsi Delete dari TodoRepository ketika terjadi error
func TestTodoRepository_Delete_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewTodoRepository(db)

	// Simulasi error saat `Delete`
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `todos` WHERE id = ?")).
		WithArgs(1).
		WillReturnError(errors.New("delete error"))
	mock.ExpectRollback()

	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete error")
	assert.NoError(t, mock.ExpectationsWereMet())
}
