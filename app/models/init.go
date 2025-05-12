package models

import (
	"context"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	"github.com/revel/revel"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"time"
)

var DB *gorm.DB
var RDB *redis.Client
var minioClient *minio.Client
var ES *elasticsearch.Client

func InitDB() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_SSLMODE"))
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database")
	}

	fmt.Println("Connected to the database successfully!")

	// Получаем список таблиц
	var tables []string
	result := DB.Raw("SELECT tablename FROM pg_tables WHERE schemaname = 'public'").Scan(&tables)
	if result.Error != nil {
		fmt.Println("Error fetching tables:", result.Error)
		return
	}

	fmt.Println("Tables in the database:")
	for _, table := range tables {
		fmt.Println("-", table)
	}
}

func InitRDB() {
	RDB = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pong, err := RDB.Ping(ctx).Result()
	if err != nil {
		revel.AppLog.Panic("Error pinging redis:" + err.Error())
	}
	revel.AppLog.Info("Revel connected: " + pong)
}

func InitMINIO() {
	endpoint := os.Getenv("S3_HOST")
	accessKeyID := os.Getenv("MINIO_ROOT_USER")
	secretAccessKey := os.Getenv("MINIO_ROOT_PASSWORD")
	useSSL := false
	var err error
	minioClient, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		revel.AppLog.Fatal("Error creating minio client:", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	backetName := os.Getenv("S3_BUCKET_NAME")
	if ok, _ := minioClient.BucketExists(ctx, backetName); !ok {
		err = minioClient.MakeBucket(ctx, backetName, minio.MakeBucketOptions{})

	}
	if err != nil {
		revel.AppLog.Fatal("Error creating bucket:", err)
	}
	policy := `{
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Principal": "*",
                "Action": ["s3:GetObject"],
                "Resource": ["arn:aws:s3:::%s/*"]
            }
        ]
    }`
	policy = fmt.Sprintf(policy, backetName)
	err = minioClient.SetBucketPolicy(ctx, backetName, policy)
	if err != nil {
		revel.AppLog.Fatal("Error setting bucket policy:", err)
	}
	revel.AppLog.Info("Connected to minio successfully!")
}
