package router

import (
	"go-todo/internal/http/handler"
	"go-todo/pkg/route"
	"net/http"
)

// PublicRoutes mengatur route publik untuk login dan pembuatan pengguna
func PublicRoutes(userHandler *handler.UserHandler) []route.Route {
	return []route.Route{
		{
			Method:  http.MethodPost,
			Path:    "/login",
			Handler: userHandler.LoginUser, // Route login untuk pengguna
		},
		{
			Method:  http.MethodPost,
			Path:    "/register",
			Handler: userHandler.CreateUser, // Route untuk mendaftarkan pengguna baru
		},
	}
}

// PrivateRoutes mengatur route privat untuk operasi pengguna dan todo
func PrivateRoutes(userHandler *handler.UserHandler, todoHandler *handler.TodoHandler) []route.Route {
	return []route.Route{
		// User Routes
		{
			Method:  http.MethodGet,
			Path:    "/users",
			Handler: userHandler.GetAllUsers, // Route untuk mengambil semua pengguna
			Roles:   []string{"admin"},       // Hanya dapat diakses oleh admin
		},
		{
			Method:  http.MethodPut,
			Path:    "/users/:id",
			Handler: userHandler.UpdateUser, // Route untuk memperbarui data pengguna berdasarkan ID
			Roles:   []string{"admin"},      // Hanya dapat diakses oleh admin
		},
		{
			Method:  http.MethodDelete,
			Path:    "/users/:id",
			Handler: userHandler.DeleteUser, // Route untuk menghapus pengguna berdasarkan ID
			Roles:   []string{"admin"},      // Hanya dapat diakses oleh admin
		},
		// Todo Routes
		{
			Method:  http.MethodGet,
			Path:    "/todos",
			Handler: todoHandler.GetAllTodos, // Route untuk mengambil semua todo
			Roles:   []string{"admin", "user"},
		},
		{
			Method:  http.MethodPost,
			Path:    "/todos",
			Handler: todoHandler.CreateTodo, // Route untuk membuat todo baru
			Roles:   []string{"admin", "user"},
		},
		{
			Method:  http.MethodPut,
			Path:    "/todos/:id",
			Handler: todoHandler.UpdateTodo, // Route untuk memperbarui todo berdasarkan ID
			Roles:   []string{"admin", "user"},
		},
		{
			Method:  http.MethodDelete,
			Path:    "/todos/:id",
			Handler: todoHandler.DeleteTodo, // Route untuk menghapus todo berdasarkan ID
			Roles:   []string{"admin"},      // Hanya dapat diakses oleh admin
		},
	}
}
