package handler

import (
	"context"
	"go-todo/internal/entity"
	"go-todo/internal/service"
	"go-todo/pkg/response"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type TodoHandler struct {
	todoService service.TodoService
}

// NewTodoHandler menginisialisasi handler baru untuk todo
func NewTodoHandler(todoService service.TodoService) *TodoHandler {
	return &TodoHandler{todoService}
}

// GetAllTodos menghandle request untuk mengambil semua todo
func (h *TodoHandler) GetAllTodos(c echo.Context) error {
	ctx := context.Background()
	todos, err := h.todoService.FindAll(ctx)
	if err != nil {
		log.Printf("Error saat memanggil FindAll: %v", err) // Tambahkan log ini
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "Gagal mengambil data todo"))
	}
	return c.JSON(http.StatusOK, response.SuccessResponse("Berhasil mengambil data todo", todos))
}

// CreateTodo menangani permintaan untuk membuat todo baru
func (h *TodoHandler) CreateTodo(c echo.Context) error {
	var todo entity.Todo
	if err := c.Bind(&todo); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "Permintaan tidak valid"))
	}

	ctx := context.Background()
	createdTodo, err := h.todoService.Create(ctx, todo)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "Gagal membuat todo"))
	}
	return c.JSON(http.StatusOK, response.SuccessResponse("Todo berhasil dibuat", createdTodo))
}

// UpdateTodo menangani permintaan untuk memperbarui data todo berdasarkan ID
func (h *TodoHandler) UpdateTodo(c echo.Context) error {
	// Mengonversi ID dari parameter URL menjadi int64
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "ID todo tidak valid"))
	}

	// Mengikat body permintaan ke struct Todo
	var todo entity.Todo
	if err := c.Bind(&todo); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "Permintaan tidak valid"))
	}

	// Menyiapkan konteks dan memanggil metode Update di service
	ctx := context.Background()
	updatedTodo, err := h.todoService.Update(ctx, id, todo)
	if err != nil {
		// Memeriksa apakah error disebabkan karena Todo tidak ditemukan
		if err.Error() == "todo tidak ditemukan" {
			return c.JSON(http.StatusNotFound, response.ErrorResponse(http.StatusNotFound, "Todo tidak ditemukan"))
		}

		// Mengembalikan error internal server untuk masalah lainnya
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "Gagal memperbarui todo"))
	}

	// Mengembalikan respons sukses dengan Todo yang diperbarui
	return c.JSON(http.StatusOK, response.SuccessResponse("Todo berhasil diperbarui", updatedTodo))
}

// DeleteTodo menangani permintaan untuk menghapus todo berdasarkan ID
func (h *TodoHandler) DeleteTodo(c echo.Context) error {
	// Mengonversi ID menjadi int64 untuk kesesuaian dengan tipe entity.Todo
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "ID todo tidak valid"))
	}

	ctx := context.Background()
	if err := h.todoService.Delete(ctx, id); err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "Gagal menghapus todo"))
	}
	return c.JSON(http.StatusOK, response.SuccessResponse("Todo berhasil dihapus", nil))
}
