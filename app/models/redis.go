package models

import (
	"context"
	"time"
)

func SetDataInRedis(key string, value []byte, timeLive time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := RDB.Set(ctx, key, value, timeLive).Err()
	if err != nil {
		return err
	}
	return nil
}

func GetDataFromRedis(key string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return RDB.Get(ctx, key).Bytes()
}

func DeleteDataFromRedis(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := RDB.Del(ctx, key).Err()
	if err != nil {
		return err
	}
	return nil
}
