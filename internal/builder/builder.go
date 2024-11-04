package builder

import (
	"go-todo/configs"
	"go-todo/internal/http/router"
	"go-todo/internal/http/handler"
	"go-todo/internal/repository"
	"go-todo/internal/service"
	"go-todo/pkg/cache"
	"go-todo/pkg/route"
	"go-todo/pkg/token"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func BuildPublicRoutes(cfg *configs.Config, db *gorm.DB, rdb *redis.Client) []route.Route {
	cacheable := cache.NewCacheable(rdb)
	userRepository := repository.NewUserRepository(db)
	tokenUseCase := token.NewTokenUseCase(cfg.JWT.SecretKey)
	
	userService := service.NewUserService(userRepository, tokenUseCase, cacheable)
	userHandler := handler.NewUserHandler(userService)

	return router.PublicRoutes(userHandler)
}

func BuildPrivateRoutes(cfg *configs.Config, db *gorm.DB, rdb *redis.Client) []route.Route {
	cacheable := cache.NewCacheable(rdb)
	userRepository := repository.NewUserRepository(db)
	tokenUseCase := token.NewTokenUseCase(cfg.JWT.SecretKey)
	
	userService := service.NewUserService(userRepository, tokenUseCase, cacheable)
	userHandler := handler.NewUserHandler(userService)

	todoRepository := repository.NewTodoRepository(db)
	todoService := service.NewTodoService(todoRepository, cacheable)
	todoHandler := handler.NewTodoHandler(todoService)

	return router.PrivateRoutes(userHandler, todoHandler)
}
