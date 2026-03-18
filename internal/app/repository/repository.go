package repository

import (
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db 			*gorm.DB
	minio 		*minio.Client
	minioBucket string 
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

	return &Repository{
		db:			 db,
		minio: 		 minioClient,
		minioBucket: "orbits",
	}, nil
}

func (r *Repository) DB() *gorm.DB {
	return r.db
}