package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-todo/configs"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

// Unit test untuk InitCache
func TestInitCache(t *testing.T) {
	cfg := configs.RedisConfig{
		Host:     "localhost",
		Port:     "6379",
		Password: "",
	}
	client := InitCache(cfg)

	assert.NotNil(t, client)
	assert.Equal(t, fmt.Sprintf("%s:%s", cfg.Host, cfg.Port), client.Options().Addr)
	assert.Equal(t, cfg.Password, client.Options().Password)
}

func TestCacheable_Set(t *testing.T) {
	db, mock := redismock.NewClientMock() // Inisialisasi Redis mock
	cache := NewCacheable(db)             // Inisialisasi cache dengan mock Redis

	key := "test-key"
	value := map[string]string{"field": "value"}
	duration := 5 * time.Minute

	// Marshal value untuk mendapatkan representasi byte
	marshalledValue, err := json.Marshal(value)
	assert.NoError(t, err)

	// Sesuaikan ekspektasi mock dengan menggunakan `gomock.Any()` atau dengan bentuk byte array
	mock.ExpectSet(key, marshalledValue, duration).SetVal("OK")

	err = cache.Set(key, value, duration)
	assert.NoError(t, err)

	// Verifikasi semua ekspektasi terpenuhi
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCacheable_Get(t *testing.T) {
	db, mock := redismock.NewClientMock() // Inisialisasi Redis mock
	cache := NewCacheable(db)             // Inisialisasi cache dengan mock Redis

	key := "test-key"
	expectedValue := map[string]string{"field": "value"}
	marshalledValue, err := json.Marshal(expectedValue)
	assert.NoError(t, err)

	// Set ekspektasi Get
	mock.ExpectGet(key).SetVal(string(marshalledValue))

	result, err := cache.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, string(marshalledValue), result)

	// Verifikasi ekspektasi terpenuhi
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCacheable_Delete(t *testing.T) {
	db, mock := redismock.NewClientMock() // Inisialisasi Redis mock
	cache := NewCacheable(db)             // Inisialisasi cache dengan mock Redis

	key := "test-key"
	mock.ExpectDel(key).SetVal(1)

	err := cache.Delete(key)
	assert.NoError(t, err)

	// Verifikasi ekspektasi terpenuhi
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCacheable_Set_MarshalError(t *testing.T) {
	db, _ := redismock.NewClientMock() // Redis mock
	cache := NewCacheable(db)

	key := "test-key"
	// Buat nilai yang tidak bisa di-marshal untuk memicu error
	value := make(chan int)

	err := cache.Set(key, value, 5*time.Minute)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gagal melakukan marshal pada nilai")
}

func TestCacheable_Set_RedisError(t *testing.T) {
	db, mock := redismock.NewClientMock()
	cache := NewCacheable(db)

	key := "test-key"
	value := map[string]string{"field": "value"}
	duration := 5 * time.Minute

	// Set ekspektasi mock untuk error saat Redis Set
	marshalledValue, _ := json.Marshal(value)
	mock.ExpectSet(key, marshalledValue, duration).SetErr(errors.New("redis set error"))

	err := cache.Set(key, value, duration)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gagal menyimpan cache")
}

func TestCacheable_Get_RedisError(t *testing.T) {
	db, mock := redismock.NewClientMock()
	cache := NewCacheable(db)

	key := "test-key"

	// Set ekspektasi mock untuk error saat Redis Get
	mock.ExpectGet(key).SetErr(errors.New("redis get error"))

	result, err := cache.Get(key)
	assert.Error(t, err)
	assert.Equal(t, "", result)
	assert.Contains(t, err.Error(), "gagal mengambil cache")
}

func TestCacheable_Get_RedisNilError(t *testing.T) {
	db, mock := redismock.NewClientMock()
	cache := NewCacheable(db)

	key := "nonexistent-key"

	// Set ekspektasi mock untuk skenario di mana Redis tidak menemukan key (redis.Nil)
	mock.ExpectGet(key).RedisNil()

	result, err := cache.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, "", result) // Harus kembali string kosong karena redis.Nil diabaikan
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCacheable_Delete_RedisError(t *testing.T) {
	db, mock := redismock.NewClientMock()
	cache := NewCacheable(db)

	key := "test-key"

	// Set ekspektasi mock untuk error saat Redis Del
	mock.ExpectDel(key).SetErr(errors.New("redis delete error"))

	err := cache.Delete(key)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gagal menghapus cache")
}
