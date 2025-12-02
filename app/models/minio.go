package models

import (
	"context"
	"errors"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/revel/revel"
	"os"
	"path"
	"time"
)

func PutFileInMINIO(filePath string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	fileStat, err := file.Stat()
	if err != nil {
		return "", err
	}
	fileName := path.Base(filePath)
	if fileName == "" {
		return "", errors.New("file name is empty")
	}
	bucketName := os.Getenv("S3_BUCKET_NAME")
	_, err = minioClient.PutObject(ctx, bucketName, fileName, file, fileStat.Size(),
		minio.PutObjectOptions{ContentType: "application/pdf"})
	if err != nil {
		return "", err
	}
	revel.AppLog.Debugf("Successfully uploaded file %s", filePath)
	fileLink := fmt.Sprintf("https://sciencehub.tech/minio/%s/%s", os.Getenv("S3_BUCKET_NAME"), fileName)
	return fileLink, nil
}

func RemoveFileFromMINIO(fileLink string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	filename := path.Base(fileLink)
	err := minioClient.RemoveObject(ctx, os.Getenv("S3_BUCKET_NAME"), filename, minio.RemoveObjectOptions{})
	return err
}
