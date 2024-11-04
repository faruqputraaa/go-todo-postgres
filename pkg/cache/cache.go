package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"go-todo/configs"
	"time"

	"github.com/redis/go-redis/v9"
)

// InitCache menginisialisasi dan mengembalikan klien Redis.
func InitCache(cfg configs.RedisConfig) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       0,
	})
	return rdb
}

// Cacheable mendefinisikan metode untuk caching di Redis.
type Cacheable interface {
	Set(key string, value interface{}, duration time.Duration) error
	Get(key string) (string, error)
	Delete(key string) error
}

type cacheable struct {
	rdb *redis.Client
}

// NewCacheable membuat instance Cacheable baru dengan klien Redis yang diberikan.
func NewCacheable(rdb *redis.Client) Cacheable {
	return &cacheable{
		rdb: rdb,
	}
}

// Set menyimpan nilai ke dalam Redis untuk durasi yang ditentukan.
func (c *cacheable) Set(key string, value interface{}, duration time.Duration) error {
	// Mengubah nilai ke format JSON untuk menyimpan struktur data yang kompleks.
	marshalledValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("gagal melakukan marshal pada nilai: %w", err)
	}

	// Menggunakan konteks dengan batas waktu untuk operasi Redis
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = c.rdb.Set(ctx, key, marshalledValue, duration).Err()
	if err != nil {
		return fmt.Errorf("gagal menyimpan cache: %w", err)
	}
	return nil
}

// Get mengambil nilai dari Redis berdasarkan kuncinya.
func (c *cacheable) Get(key string) (string, error) {
	// Menggunakan konteks dengan batas waktu untuk operasi Redis
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	value, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // Kunci tidak ditemukan
	} else if err != nil {
		return "", fmt.Errorf("gagal mengambil cache: %w", err)
	}
	return value, nil
}

// Delete menghapus pasangan kunci-nilai dari Redis berdasarkan kuncinya.
func (c *cacheable) Delete(key string) error {
	// Menggunakan konteks dengan batas waktu untuk operasi Redis
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := c.rdb.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("gagal menghapus cache: %w", err)
	}
	return nil
}
