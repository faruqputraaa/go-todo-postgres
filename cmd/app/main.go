package main

import (
	"context"
	"fmt"
	"go-todo/configs"
	"go-todo/internal/builder"
	"go-todo/pkg/cache"
	"go-todo/pkg/database"
	"go-todo/pkg/server"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	// cfg, err := configs.NewConfig(".env")
	cfg, err := configs.NewConfigYaml("config.yaml")
	checkError(err)

	db, err := database.InitDatabase(cfg.PostgresConfig)
	checkError(err)

	rdb := cache.InitCache(cfg.RedisConfig)

	publicRoutes := builder.BuildPublicRoutes(cfg, db, rdb)
	privateRoutes := builder.BuildPrivateRoutes(cfg, db, rdb)

	srv := server.NewServer(cfg, publicRoutes, privateRoutes)
	runServer(srv, cfg.PORT)
	waitForShutdown(srv)
}

// checkError untuk memeriksa error, jika error terjadi maka program akan berhenti
func checkError(err error) {
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}

// runServer menjalankan server pada port yang ditentukan
func runServer(srv *server.Server, port string) {
	go func() {
		addr := fmt.Sprintf(":%s", port)
		log.Printf("Server berjalan di %s", addr)
		if err := srv.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Gagal memulai server: %v", err)
		}
	}()
}

// waitForShutdown menangani proses shutdown server ketika menerima sinyal interrupt
func waitForShutdown(srv *server.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	<-quit
	log.Println("Mematikan server...")

	// Memberi waktu untuk server menyelesaikan permintaan yang sedang berjalan
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Gagal mematikan server: %v", err)
	}
	log.Println("Server berhasil dimatikan.")
}
