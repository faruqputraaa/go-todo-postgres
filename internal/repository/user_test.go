package repository

import (
	"context"
	"errors"
	"go-todo/internal/entity"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// TestUserRepository_FindAll menguji fungsi FindAll pada UserRepository
func TestUserRepository_FindAll(t *testing.T) {
	db, mock := setupMockDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	rows := sqlmock.NewRows([]string{"id", "username", "full_name", "role"}).
		AddRow(1, "user1", "User One", "admin").
		AddRow(2, "user2", "User Two", "user")

	mock.ExpectQuery("SELECT \\* FROM `users`").WillReturnRows(rows)

	users, err := repo.FindAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, users, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestUserRepository_FindByID menguji fungsi FindByID pada UserRepository
func TestUserRepository_FindByID(t *testing.T) {
	db, mock := setupMockDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	row := sqlmock.NewRows([]string{"id", "username", "full_name", "role"}).
		AddRow(1, "user1", "User One", "admin")

	// Gunakan regexp.QuoteMeta untuk menghindari masalah escaping pada regex
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE id = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs(1, 1).
		WillReturnRows(row)

	user, err := repo.FindByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, "user1", user.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestUserRepository_FindByID_NotFound menguji fungsi FindByID pada UserRepository ketika data tidak ditemukan
func TestUserRepository_FindByID_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	// Gunakan regexp.QuoteMeta untuk menghindari masalah escaping pada regex
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE id = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs(1, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err := repo.FindByID(context.Background(), 1)
	assert.ErrorIs(t, err, ErrPenggunaTidakDitemukan)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestUserRepository_FindByUsername menguji fungsi FindByUsername pada UserRepository
func TestUserRepository_FindByUsername(t *testing.T) {
	db, mock := setupMockDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	row := sqlmock.NewRows([]string{"id", "username", "full_name", "role"}).
		AddRow(1, "user1", "User One", "admin")

	// Gunakan regexp.QuoteMeta dan tambahkan dua argumen pada WithArgs
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE username = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs("user1", 1). // Argumen pertama untuk username, kedua untuk LIMIT
		WillReturnRows(row)

	user, err := repo.FindByUsername(context.Background(), "user1")
	assert.NoError(t, err)
	assert.Equal(t, "user1", user.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestUserRepository_FindByUsername_NotFound menguji fungsi FindByUsername pada UserRepository ketika data tidak ditemukan
func TestUserRepository_FindByUsername_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	// Gunakan regexp.QuoteMeta untuk menghindari masalah escaping pada regex
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE username = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs("user1", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err := repo.FindByUsername(context.Background(), "user1")
	assert.ErrorIs(t, err, ErrPenggunaTidakDitemukan)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestUserRepository_Create menguji fungsi Create pada UserRepository
func TestUserRepository_Create(t *testing.T) {
	db, mock := setupMockDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	user := &entity.User{
		Username: "newuser",
		Password: "password123",
		FullName: "New User",
		Role:     "user",
	}

	// Urutan kolom sesuai dengan query sebenarnya: `username`, `password`, `role`, `full_name`
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `users` (`username`,`password`,`role`,`full_name`) VALUES (?,?,?,?)")).
		WithArgs(user.Username, user.Password, user.Role, user.FullName).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	createdUser, err := repo.Create(context.Background(), user)
	assert.NoError(t, err)
	assert.Equal(t, "newuser", createdUser.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestUserRepository_Update menguji fungsi Update pada UserRepository
func TestUserRepository_Update(t *testing.T) {
	db, mock := setupMockDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	user := &entity.User{
		ID:       1,
		Username: "updateduser",
		FullName: "Updated User",
		Role:     "user",
		Password: "newpassword123",
	}

	// Ekspektasi untuk query `UPDATE`
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `users` SET `full_name`=?,`password`=?,`role`=?,`username`=? WHERE id = ?")).
		WithArgs(user.FullName, user.Password, user.Role, user.Username, user.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Ekspektasi untuk query `SELECT` setelah `UPDATE`
	row := sqlmock.NewRows([]string{"id", "username", "full_name", "role", "password"}).
		AddRow(1, "updateduser", "Updated User", "user", "newpassword123")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE id = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs(1, 1).
		WillReturnRows(row)

	updatedUser, err := repo.Update(context.Background(), user)
	assert.NoError(t, err)
	assert.Equal(t, "updateduser", updatedUser.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestUserRepository_Update_NotFound menguji fungsi Update pada UserRepository ketika data tidak ditemukan
func TestUserRepository_Update_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	user := &entity.User{
		ID:       1,
		Username: "updateduser",
		FullName: "Updated User",
		Role:     "user",
	}

	// Sesuaikan urutan kolom dan simulasi "not found" dengan RowsAffected = 0
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `users` SET `full_name`=?,`role`=?,`username`=? WHERE id = ?")).
		WithArgs(user.FullName, user.Role, user.Username, user.ID).
		WillReturnResult(sqlmock.NewResult(0, 0)) // Simulasi RowsAffected = 0 untuk not found
	mock.ExpectCommit()

	_, err := repo.Update(context.Background(), user)
	assert.ErrorIs(t, err, ErrPenggunaTidakDitemukan) // Pastikan error sesuai dengan ErrPenggunaTidakDitemukan
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestUserRepository_Delete menguji fungsi Delete pada UserRepository
func TestUserRepository_Delete(t *testing.T) {
	db, mock := setupMockDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `users` WHERE id = ?").WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1)) // RowsAffected = 1
	mock.ExpectCommit()

	err := repo.Delete(context.Background(), 1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestUserRepository_Delete_NotFound menguji fungsi Delete pada UserRepository ketika data tidak ditemukan
func TestUserRepository_Delete_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `users` WHERE id = ?").WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 0)) // Tidak ada baris yang terpengaruh
	mock.ExpectCommit()

	err := repo.Delete(context.Background(), 1)
	assert.ErrorIs(t, err, ErrPenggunaTidakDitemukan)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestUserRepository_FindAll_Error menguji fungsi FindAll pada UserRepository ketika terjadi error pada database
func TestUserRepository_FindAll_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	// Simulasikan error saat query `FindAll`
	mock.ExpectQuery("SELECT \\* FROM `users`").WillReturnError(errors.New("database error"))

	_, err := repo.FindAll(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "terjadi kesalahan pada database")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestUserRepository_Create_Error menguji fungsi Create pada UserRepository ketika terjadi error pada database
func TestUserRepository_Create_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	user := &entity.User{
		Username: "newuser",
		Password: "password123",
		FullName: "New User",
		Role:     "user",
	}

	// Simulasikan error saat insert `Create`
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `users` (`username`,`password`,`role`,`full_name`) VALUES (?,?,?,?)")).
		WithArgs(user.Username, user.Password, user.Role, user.FullName).
		WillReturnError(errors.New("insert error"))
	mock.ExpectRollback()

	_, err := repo.Create(context.Background(), user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "terjadi kesalahan pada database")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestUserRepository_Update_Error menguji fungsi Update pada UserRepository ketika terjadi error pada database
func TestUserRepository_Update_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	user := &entity.User{
		ID:       1,
		Username: "updateduser",
		FullName: "Updated User",
		Role:     "user",
		Password: "newpassword123",
	}

	// Simulasikan error saat `Update`
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `users` SET `full_name`=?,`password`=?,`role`=?,`username`=? WHERE id = ?")).
		WithArgs(user.FullName, user.Password, user.Role, user.Username, user.ID).
		WillReturnError(errors.New("update error"))
	mock.ExpectRollback()

	_, err := repo.Update(context.Background(), user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "terjadi kesalahan pada database")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestUserRepository_FindByID_Error menguji fungsi FindByID pada UserRepository ketika terjadi error pada database
func TestUserRepository_FindByID_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	// Simulasikan error selain `ErrRecordNotFound`
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE id = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs(1, 1).
		WillReturnError(errors.New("database error"))

	_, err := repo.FindByID(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "terjadi kesalahan pada database")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestUserRepository_FindByUsername_Error menguji fungsi FindByUsername pada UserRepository ketika terjadi error pada database
func TestUserRepository_FindByUsername_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	// Simulasikan error saat query `FindByUsername`
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE username = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs("user1", 1).
		WillReturnError(errors.New("database error"))

	_, err := repo.FindByUsername(context.Background(), "user1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "terjadi kesalahan pada database")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestUserRepository_Update_FindError menguji fungsi Update pada UserRepository ketika terjadi error pada database saat FindByID setelah Update
func TestUserRepository_Update_FindError(t *testing.T) {
	db, mock := setupMockDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	user := &entity.User{
		ID:       1,
		Username: "updateduser",
		FullName: "Updated User",
		Role:     "user",
		Password: "newpassword123",
	}

	// Simulasi pembaruan berhasil
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `users` SET `full_name`=?,`password`=?,`role`=?,`username`=? WHERE id = ?")).
		WithArgs(user.FullName, user.Password, user.Role, user.Username, user.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Simulasikan error saat `FindByID` setelah `Update`
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE id = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs(user.ID, 1).
		WillReturnError(errors.New("database error"))

	_, err := repo.Update(context.Background(), user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "terjadi kesalahan pada database")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestUserRepository_Delete_Error menguji fungsi Delete pada UserRepository ketika terjadi error pada database
func TestUserRepository_Delete_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	// Simulasikan error saat `Delete`
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `users` WHERE id = ?").WithArgs(1).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "terjadi kesalahan pada database")
	assert.NoError(t, mock.ExpectationsWereMet())
}
