package repository

import (
	"log"
	"os"
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db 			*gorm.DB
	minio 		*minio.Client
	minioBucket string 
	redis       *redis.Client
}

func New(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ROOT_USER")
	secretKey := os.Getenv("MINIO_ROOT_PASSWORD")

	if endpoint == "" {
		endpoint = "localhost:9000"
		accessKey = "root"
		secretKey = "rootroot"
	}

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	
	if err != nil {
		log.Println("Ошибка инициализации MinIO:", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Убедитесь, что порт совпадает с docker-compose
		Password: "password",       // Пароль из вашего docker-compose
		DB:       0,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Println("Внимание: Не удалось подключиться к Redis:", err)
	}

	return &Repository{
		db:          db,
		minio:       minioClient,
		minioBucket: "orbits",
		redis:       rdb, // Добавляем в репозиторий
	}, nil
}

func (r *Repository) DB() *gorm.DB {
	return r.db
}

func (r *Repository) Redis() *redis.Client {
	return r.redis
}