package controllers

import (
	"crypto/rand"
	"fmt"
	"github.com/revel/revel"
	"math/big"
	"scinceHub/app/models"
	"time"
)

func GenerateRandomNumber() (*big.Int, error) {
	max := big.NewInt(100000)
	randomNumber, err := rand.Int(rand.Reader, max)
	if err != nil {
		fmt.Println("Ошибка при генерации случайного числа:", err)
		return nil, err
	}
	return randomNumber, nil
}

func SetVerificationEmailCode(key, code string, timeLife time.Duration) {
	key = key + "_" + "verifyEmailCode"
	err := models.SetDataInRedis(key, []byte(code), timeLife)
	if err != nil {
		revel.AppLog.Error(err.Error())
		models.SetVerificationEmailCode(key, code, timeLife)
	} else {
		revel.AppLog.Debug("установили код верификации " + code + "по ключу: " + key)
	}
}

func GetVerificationEmailCode(key string) (string, bool) {
	key = key + "_" + "verifyEmailCode"
	value, err := models.GetDataFromRedis(key)
	revel.AppLog.Info(key + "_" + "verifyEmailCode")
	if err != nil || value == nil {
		revel.AppLog.Error(err.Error())
		return models.GetVerificationEmailCode(key)
	}
	revel.AppLog.Info("получили код верификации из Redis " + string(value))
	return string(value), true
}

func DeleteVerificationEmailCode(key string) {
	key = key + "_" + "verifyEmailCode"
	err := models.DeleteDataFromRedis(key)
	if err != nil {
		revel.AppLog.Error(err.Error())
		models.DeleteVerificationEmailCode(key)
	}
}

func SetChangeEmailCode(key, code string, timeLife time.Duration) {
	key = key + "_" + "changeEmailCode"
	err := models.SetDataInRedis(key, []byte(code), timeLife)
	if err != nil {
		revel.AppLog.Error(err.Error())
		models.SetChangeEmailCode(key, code, timeLife)
	} else {
		revel.AppLog.Debug("установили код смены " + code + "по ключу: " + key)
	}

}

func GetChangeEmailCode(key string) (string, bool) {
	key = key + "_" + "changeEmailCode"
	value, err := models.GetDataFromRedis(key)
	revel.AppLog.Info(key)
	if err != nil || value == nil {
		revel.AppLog.Error(err.Error())
		return models.GetChangeEmailCode(key)
	}
	revel.AppLog.Info("получили код смены из Redis " + string(value))
	return string(value), true
}

func DeleteChangeEmailCode(key string) {
	key = key + "_" + "changeEmailCode"
	err := models.DeleteDataFromRedis(key)
	if err != nil {
		revel.AppLog.Error(err.Error())
		models.DeleteChangeEmailCode(key)
	}
}

func SetChangePasswordCode(key, code string, timeLife time.Duration) {
	key = key + "_" + "changePasswordCode"
	err := models.SetDataInRedis(key, []byte(code), timeLife)
	if err != nil {
		revel.AppLog.Error(err.Error())
		models.SetChangePasswordCode(key, code, timeLife)
	} else {
		revel.AppLog.Debug("установили код смены " + code + "по ключу: " + key)
	}

}

func GetChangePasswordCode(key string) (string, bool) {
	key = key + "_" + "changePasswordCode"
	value, err := models.GetDataFromRedis(key)
	revel.AppLog.Info(key)
	if err != nil || value == nil {
		revel.AppLog.Error(err.Error())
		return models.GetChangePasswordCode(key)
	}
	revel.AppLog.Info("получили код смены из Redis " + string(value))
	return string(value), true
}

func DeleteChangePasswordCode(key string) {
	key = key + "_" + "changePasswordCode"
	err := models.DeleteDataFromRedis(key)
	if err != nil {
		revel.AppLog.Error(err.Error())
		models.DeleteChangePasswordCode(key)
	}
}
