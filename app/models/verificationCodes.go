package models

import (
	"github.com/revel/revel"
	"sync"
	"time"
)

type VerificationEmail struct {
	Code     string
	TimeLife time.Time
}

type ChangeEmail struct {
	Code     string
	TimeLife time.Time
}

type ChangePassword struct {
	Code     string
	TimeLife time.Time
}

type VerifyProfile struct {
	Profile     Profiles
	Code        string `json:"code"`
	OldPassword string `json:"old_password"`
}

//var VerificationEmailCodes sync.Map
//
//var ChangeEmailCodes sync.Map
//
//var ChangePasswordCodes sync.Map
//
//var mutex = new(sync.Mutex)
//
//func SetVerificationEmailCode(key, code string, timeLife time.Duration) {
//	mutex.Lock()
//	defer mutex.Unlock()
//	VerificationEmailCodes.Store(key, VerificationEmail{
//		Code:     code,
//		TimeLife: time.Now().UTC().Add(timeLife),
//	})
//}
//
//func GetVerificationEmailCode(key string) (string, bool) {
//	data, ok := VerificationEmailCodes.Load(key)
//	info := data.(VerificationEmail)
//	if !ok || time.Now().UTC().After(info.TimeLife) {
//		return "", false
//	}
//	return info.Code, ok
//}
//
//func UpdateVerificationEmailCodes() {
//	for i, v := range VerificationEmailCodes {
//		info := v.(VerificationEmail)
//		if !time.Now().UTC().After(info.TimeLife) {
//			VerificationEmailCodes.Delete(i)
//		}
//	}
//}

var VerificationEmailCodes = make(map[string]VerificationEmail)

var ChangeEmailCodes = make(map[string]ChangeEmail)

var ChangePasswordCodes = make(map[string]ChangePassword)

var mutex = new(sync.RWMutex)

func SetVerificationEmailCode(key, code string, timeLife time.Duration) {
	mutex.Lock()
	defer mutex.Unlock()
	VerificationEmailCodes[key] = VerificationEmail{
		Code:     code,
		TimeLife: time.Now().UTC().Add(timeLife),
	}
}

func GetVerificationEmailCode(key string) (string, bool) {
	mutex.RLock()
	defer mutex.RUnlock()
	data, ok := VerificationEmailCodes[key]
	if !ok || time.Now().UTC().After(data.TimeLife) {
		return "", false
	}
	return data.Code, ok
}

func UpdateVerificationEmailCodes() {
	mutex.Lock()
	defer mutex.Unlock()
	for i, v := range VerificationEmailCodes {
		if time.Now().UTC().After(v.TimeLife) {
			delete(VerificationEmailCodes, i)
		}
	}
}

func DeleteVerificationEmailCode(key string) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(VerificationEmailCodes, key)
}

func SetChangeEmailCode(key, code string, timeLife time.Duration) {
	mutex.Lock()
	defer mutex.Unlock()
	ChangeEmailCodes[key] = ChangeEmail{
		Code:     code,
		TimeLife: time.Now().UTC().Add(timeLife),
	}
}

func GetChangeEmailCode(key string) (string, bool) {
	mutex.RLock()
	defer mutex.RUnlock()
	data, ok := ChangeEmailCodes[key]
	if !ok || time.Now().UTC().After(data.TimeLife) {
		return "", false
	}
	return data.Code, ok
}

func UpdateChangeEmailCodes() {
	mutex.Lock()
	defer mutex.Unlock()
	for i, v := range ChangeEmailCodes {
		if time.Now().UTC().After(v.TimeLife) {
			delete(ChangeEmailCodes, i)
		}
	}
}

func DeleteChangeEmailCode(key string) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(ChangeEmailCodes, key)
}

func SetChangePasswordCode(key, code string, timeLife time.Duration) {
	mutex.Lock()
	defer mutex.Unlock()
	ChangePasswordCodes[key] = ChangePassword{
		Code:     code,
		TimeLife: time.Now().UTC().Add(timeLife),
	}
}

func GetChangePasswordCode(key string) (string, bool) {
	mutex.RLock()
	defer mutex.RUnlock()
	data, ok := ChangePasswordCodes[key]
	if !ok || time.Now().UTC().After(data.TimeLife) {
		return "", false
	}
	return data.Code, ok
}

func UpdateChangePasswordCodes() {
	mutex.Lock()
	defer mutex.Unlock()
	for i, v := range ChangePasswordCodes {
		if time.Now().UTC().After(v.TimeLife) {
			delete(ChangePasswordCodes, i)
		}
	}
}

func DeleteChangePasswordCode(key string) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(ChangePasswordCodes, key)
}

func StartUpdateVerifyCodeLists() {
	go func() {
		tiker := time.NewTicker(5 * time.Minute)
		defer tiker.Stop()
		for {
			UpdateVerificationEmailCodes()
			UpdateChangeEmailCodes()
			UpdateChangePasswordCodes()
			revel.AppLog.Debug("Обновили список кодов верификации")
			<-tiker.C
		}
	}()
}
