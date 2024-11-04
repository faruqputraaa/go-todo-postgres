package handler

import (
	"errors"
	"go-todo/internal/entity"
	"go-todo/internal/service"
	"go-todo/pkg/response"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	userService service.UserService
}

// NewUserHandler membuat instance baru dari UserHandler
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetAllUsers menangani permintaan untuk mendapatkan semua pengguna
func (h *UserHandler) GetAllUsers(c echo.Context) error {
	users, err := h.userService.FindAll(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError,
			response.ErrorResponse(http.StatusInternalServerError, "Gagal mengambil data pengguna"))
	}

	return c.JSON(http.StatusOK,
		response.SuccessResponse("Data pengguna berhasil diambil", users))
}

// CreateUser menangani permintaan untuk membuat pengguna baru
func (h *UserHandler) CreateUser(c echo.Context) error {
	var req struct {
		Username string `json:"username" validate:"required"`
		Password string `json:"password" validate:"required"`
		FullName string `json:"full_name" validate:"required"`
		Role     string `json:"role" validate:"required"`
	}

	// Validasi permintaan
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest,
			response.ErrorResponse(http.StatusBadRequest, "Format permintaan tidak valid"))
	}

	user := &entity.User{
		Username: req.Username,
		Password: req.Password, 
		FullName: req.FullName,
		Role:     req.Role,
	}

	createdUser, err := h.userService.CreateUser(c.Request().Context(), user)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrUsernameSudahAda) {
			status = http.StatusConflict
		}
		return c.JSON(status, response.ErrorResponse(status, err.Error()))
	}

	return c.JSON(http.StatusCreated,
		response.SuccessResponse("Pengguna berhasil dibuat", createdUser))
}

// LoginUser menangani permintaan login
func (h *UserHandler) LoginUser(c echo.Context) error {
	var req struct {
		Username string `json:"username" validate:"required"`
		Password string `json:"password" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest,
			response.ErrorResponse(http.StatusBadRequest, "Format permintaan tidak valid"))
	}

	if req.Username == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest,
			response.ErrorResponse(http.StatusBadRequest, "Username dan password harus diisi"))
	}

	token, err := h.userService.Login(c.Request().Context(), req.Username, req.Password)
	if err != nil {
		status := http.StatusUnauthorized
		if err == service.ErrServerInternal {
			status = http.StatusInternalServerError
		}
		return c.JSON(status, response.ErrorResponse(status, err.Error()))
	}

	return c.JSON(http.StatusOK,
		response.SuccessResponse("Login berhasil", map[string]string{"token": token}))
}

// UpdateUser menangani permintaan untuk memperbarui data pengguna
func (h *UserHandler) UpdateUser(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest,
			response.ErrorResponse(http.StatusBadRequest, "ID pengguna tidak valid"))
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		FullName string `json:"full_name"`
		Role     string `json:"role"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest,
			response.ErrorResponse(http.StatusBadRequest, "Format permintaan tidak valid"))
	}

	user := &entity.User{
		ID:       id,
		Username: req.Username,
		Password: req.Password, // Password akan di-hash di service layer
		FullName: req.FullName,
		Role:     req.Role,
	}

	updatedUser, err := h.userService.UpdateUser(c.Request().Context(), user)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrPenggunaTidakDitemukan) {
			status = http.StatusNotFound
		}
		return c.JSON(status, response.ErrorResponse(status, err.Error()))
	}

	return c.JSON(http.StatusOK,
		response.SuccessResponse("Data pengguna berhasil diperbarui", updatedUser))
}

// DeleteUser menangani permintaan untuk menghapus pengguna
func (h *UserHandler) DeleteUser(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest,
			response.ErrorResponse(http.StatusBadRequest, "ID pengguna tidak valid"))
	}

	if err := h.userService.DeleteUser(c.Request().Context(), id); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrPenggunaTidakDitemukan) {
			status = http.StatusNotFound
		}
		return c.JSON(status, response.ErrorResponse(status, err.Error()))
	}

	return c.JSON(http.StatusOK,
		response.SuccessResponse("Pengguna berhasil dihapus", nil))
}
