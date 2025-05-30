package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/revel/revel"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"scinceHub/app/middleware"
	"scinceHub/app/models"
	"scinceHub/app/smtp"
	"time"
)

type Profiles struct {
	*revel.Controller
}

func (p Profiles) SendVerificationCodeForRegister() revel.Result {
	profile := new(models.Profiles)
	err := p.Params.BindJSON(profile)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	if models.ThsProfilesIsExist(profile.Login) {
		p.Response.Status = http.StatusConflict
		return p.RenderJSON(map[string]string{"error": "Пользователь с таким email уже существует"})
	}
	randomNumber, err := GenerateRandomNumber()
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": "Не удалось сгенерировать код подтверждения"})
	}

	verificationCode := fmt.Sprintf("%06d", randomNumber)
	SetVerificationEmailCode(profile.Login, verificationCode, 5*time.Minute)

	err = smtp.SendMessage(profile.Login, "Подтверждение почты", verificationCode)
	if err != nil {
		p.Response.Status = http.StatusUnprocessableEntity
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	revel.AppLog.Info("message sent")
	return p.RenderJSON(map[string]string{"message": "Код подтверждения отправлен"})
}

func (p Profiles) SendVerificationCodeForChangeEmail() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		p.Response.Status = http.StatusUnauthorized
		return nil
	}
	sUserID := fmt.Sprintf("%d", userID)
	profile := new(models.Profiles)
	err = p.Params.BindJSON(profile)
	if err != nil {
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]int{"status": http.StatusBadRequest})
	}
	_, ok := GetChangePasswordCode(sUserID)
	if ok {
		p.Response.Status = http.StatusConflict
		return p.RenderJSON(map[string]string{"error": "Завершите смену пароля!"})
	}

	if models.ThsProfilesIsExist(profile.Login) {
		p.Response.Status = http.StatusConflict
		return p.RenderJSON(map[string]string{"error": "Пользователь с таким email уже существует"})
	}

	randomNumber, err := GenerateRandomNumber()
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": "Не удалось сгенерировать код подтверждения"})
	}

	verificationCode := fmt.Sprintf("%06d", randomNumber)
	SetChangeEmailCode(sUserID, verificationCode, 5*time.Minute)

	err = smtp.SendMessage(profile.Login, "Смена почты", verificationCode)
	if err != nil {
		p.Response.Status = http.StatusUnprocessableEntity
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	revel.AppLog.Info("message sent")
	return p.RenderJSON(map[string]string{"message": "Код смены почты"})
}

func (p Profiles) VerifyAndChangeEmail() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		p.Response.Status = http.StatusUnauthorized
		return nil
	}
	sUserID := fmt.Sprintf("%d", userID)
	var vprofile = new(models.VerifyProfile)
	err = p.Params.BindJSON(vprofile)
	if err != nil {
		return p.RenderJSON(map[string]int{"status": http.StatusBadRequest})
	}
	validate := validator.New()
	err = validate.Struct(vprofile.Profile)
	if err != nil || vprofile.Code == "" {
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]int{"status": http.StatusBadRequest})
	}
	changeEmailCode, ok := GetChangeEmailCode(sUserID)
	if !ok {
		p.Response.Status = http.StatusNotFound
		return p.RenderJSON(map[string]string{"error": "Код подтверждения не найден или истек его срок. Пожалуйста, запросите новый код."})
	}

	if changeEmailCode != vprofile.Code {
		p.Response.Status = http.StatusUnauthorized
		return p.RenderJSON(map[string]string{"error": "Неверный код подтверждения"})
	}

	err = models.UpdateProfileByID(userID, &vprofile.Profile)
	if err != nil {
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]int{"status": http.StatusInternalServerError})
	}
	DeleteChangeEmailCode(sUserID)
	_ = models.DeleteDataFromRedis(sUserID)
	p.Response.Status = http.StatusOK
	return nil
}

func (p Profiles) StopChangeEmail() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		p.Response.Status = http.StatusUnauthorized
		return nil
	}
	sUserID := fmt.Sprintf("%d", userID)
	DeleteChangeEmailCode(sUserID)
	p.Response.Status = http.StatusOK
	return nil
}

func (p Profiles) SendVerificationCodeForChangePassword() revel.Result {
	profile := new(models.Profiles)
	err := p.Params.BindJSON(profile)
	if err != nil {
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]int{"status": http.StatusBadRequest})
	}

	ok := models.ThsProfilesIsExist(profile.Login)
	if !ok {
		return p.RenderJSON(map[string]string{"error": "Пользователя с такой почтой не существует"})
	}
	randomNumber, err := GenerateRandomNumber()
	if err != nil {
		return p.RenderJSON(map[string]int{"status": http.StatusInternalServerError})
	}

	verificationCode := fmt.Sprintf("%06d", randomNumber)
	pdata, _ := models.GetProfileLoginData(profile.Login)
	sUserID := fmt.Sprintf("%d", pdata.ID)
	SetChangePasswordCode(sUserID, verificationCode, 5*time.Minute)

	err = smtp.SendMessage(profile.Login, "Смена пароля", verificationCode)
	if err != nil {
		p.Response.Status = http.StatusUnprocessableEntity
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	revel.AppLog.Info("message sent")
	return p.RenderJSON(map[string]string{"message": "Код смены пароля"})
}

func (p Profiles) VerifyAndChangePassword() revel.Result {
	var vprofile = new(models.VerifyProfile)
	err := p.Params.BindJSON(vprofile)
	if err != nil {
		return p.RenderJSON(map[string]int{"status": http.StatusBadRequest})
	}
	validate := validator.New()
	err = validate.Struct(vprofile.Profile)
	if err != nil || vprofile.Code == "" {
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]int{"status": http.StatusBadRequest})
	}
	revel.AppLog.Debugf("%v\n", vprofile)
	pdata, _ := models.GetProfileLoginData(vprofile.Profile.Login)
	sUserID := fmt.Sprintf("%d", pdata.ID)
	changePasswordCode, ok := GetChangePasswordCode(sUserID)
	if !ok {
		return p.RenderJSON(map[string]int{"status": http.StatusNotFound})
	}
	if changePasswordCode != vprofile.Code {
		return p.RenderJSON(map[string]int{"status": http.StatusNotAcceptable})
	}
	hashPassword, err := middleware.HashPassword(vprofile.Profile.Password)
	if err != nil {
		return p.RenderJSON(map[string]int{"status": http.StatusInternalServerError})
	}
	vprofile.Profile.Password = hashPassword
	err = models.UpdateProfileByID(pdata.ID, &vprofile.Profile)
	if err != nil {
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]int{"status": http.StatusInternalServerError})
	}
	DeleteChangePasswordCode(sUserID)
	_ = models.DeleteDataFromRedis(sUserID)
	p.Response.Status = http.StatusOK
	return nil
}

func (p Profiles) StopChangePassword() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		p.Response.Status = http.StatusUnauthorized
		return nil
	}
	sUserID := fmt.Sprintf("%d", userID)

	DeleteChangePasswordCode(sUserID)
	p.Response.Status = http.StatusOK
	return nil
}

func (p Profiles) VerifyAndCreateUser() revel.Result {
	var vprofile = new(models.VerifyProfile)
	err := p.Params.BindJSON(vprofile)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": "Неверный запрос"})
	}
	validate := validator.New()
	err = validate.Struct(vprofile.Profile)
	if err != nil || vprofile.Code == "" {
		p.Response.Status = http.StatusBadRequest
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	VerificationEmailCode, ok := GetVerificationEmailCode(vprofile.Profile.Login)
	if !ok {
		p.Response.Status = http.StatusNotFound
		return p.RenderJSON(map[string]string{"error": "Код подтверждения не найден или не действителен. Пожалуйста, запросите новый код."})
	}

	if VerificationEmailCode != vprofile.Code {
		p.Response.Status = http.StatusUnauthorized
		return p.RenderJSON(map[string]string{"error": "Неверный код подтверждения"})
	}

	hashPassword, err := middleware.HashPassword(vprofile.Profile.Password)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	vprofile.Profile.Password = hashPassword
	err = models.CreateProfile(&vprofile.Profile)

	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusCreated
	return nil
}

func (p Profiles) Login(login, password string) revel.Result {
	user, err := models.GetProfileLoginData(login)
	if err != nil {
		p.Response.Status = http.StatusUnauthorized
		return nil
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		p.Response.Status = http.StatusUnauthorized
		return nil
	}
	token, err := middleware.GenerateJWT(user.ID)

	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": "Ошибка генерации токена"})
	}
	middleware.SetCookieData(p.Controller, "auth_token", token, false)

	p.Response.Status = http.StatusOK
	return nil
}

func (p Profiles) Logout() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err == nil {
		sUserID := fmt.Sprintf("%d", userID)
		_ = models.DeleteDataFromRedis(sUserID)
	}
	middleware.SetCookieData(p.Controller, "auth_token", "", true)
	p.Response.Status = http.StatusNoContent
	return nil
}

func (p Profiles) GetProfileByID(id uint64) revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		p.Response.Status = http.StatusUnauthorized
		return nil
	}
	profile, err := models.GetProfileByID(id)
	if err != nil {
		p.Response.Status = http.StatusNotFound
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	profileWithSubscribitionStatus := new(models.ProfileWithSubscribitionStatus)
	profileWithSubscribitionStatus.Profile = *profile
	profileWithSubscribitionStatus.Isubscribed = models.CheckMySubscribesForProfile(userID, id)
	profileWithSubscribitionStatus.IsSubscribed = models.CheckMySubscribesForProfile(id, userID)
	p.Response.Status = http.StatusOK
	return p.RenderJSON(profileWithSubscribitionStatus)
}

func (p Profiles) GetUserData() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		p.Response.Status = http.StatusUnauthorized
		return nil
	}
	profile := new(models.Profiles)
	sUserID := fmt.Sprintf("%d", userID)
	data, err := models.GetDataFromRedis(sUserID)
	if data != nil && err == nil {
		err = json.Unmarshal(data, profile)
		if err == nil {
			revel.AppLog.Debug("Данные профиля получили с redis")
			return p.RenderJSON(profile)
		}
	}
	profile, err = models.GetUserProfile(userID)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	data, err = json.Marshal(profile)
	if err == nil {
		_ = models.SetDataInRedis(sUserID, data, time.Hour)
	}

	return p.RenderJSON(profile)
}

func (p Profiles) GetUsersDataForCreatePublication() revel.Result {
	_, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		p.Response.Status = http.StatusUnauthorized
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	profile, _ := models.GetAllProfileIDAndNames()

	return p.RenderJSON(profile)
}

func (p Profiles) DeleteProfileByID(id uint64) revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	_, err2 := middleware.ValidateAdminJWT(p.Request, "auth_token_admin")
	if (err != nil || userID != id) && err2 != nil {
		p.Response.Status = http.StatusUnauthorized
		return nil

	}
	sUserID := fmt.Sprintf("%d", id)
	err = models.DeleteProfileByID(id)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	middleware.SetCookieData(p.Controller, "auth_token", "", true)
	_ = models.DeleteDataFromRedis(sUserID)
	p.Response.Status = http.StatusOK
	return nil
}

func (p Profiles) DeleteProfileByLogin(login string) revel.Result {
	_, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		p.Response.Status = http.StatusUnauthorized
		return nil
	}
	err = models.DeleteProfileByLogin(login)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusNoContent
	p.Response.Status = http.StatusOK
	return nil
}

func (p Profiles) UpdateProfileByID() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		p.Response.Status = http.StatusUnauthorized
		return nil
	}
	sUserID := fmt.Sprintf("%d", userID)
	profile := new(models.Profiles)
	err = p.Params.BindJSON(profile)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	validate := validator.New()
	err = validate.Struct(profile)
	if err != nil {
		p.Response.Status = http.StatusUnprocessableEntity
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	err = models.UpdateProfileByID(userID, profile)

	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	p.Response.Status = http.StatusNoContent
	_ = models.DeleteDataFromRedis(sUserID)
	p.Response.Status = http.StatusNoContent
	return nil
}

func (p Profiles) UpdateProfileByLogin(login string) revel.Result {
	profile := new(models.Profiles)
	err := p.Params.BindJSON(profile)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	validate := validator.New()
	err = validate.Struct(profile)
	if err != nil {
		p.Response.Status = http.StatusUnprocessableEntity
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	err = models.UpdateProfileByLogin(login, profile)

	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	p.Response.Status = http.StatusNoContent
	p.Response.Status = http.StatusNoContent
	return nil
}

func (p Profiles) GetAllProfiles() revel.Result {
	_, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		p.Response.Status = http.StatusUnauthorized
		return nil
	}
	profiles, err := models.GetAllProfiles()
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusOK
	return p.RenderJSON(profiles)
}

func (p Profiles) GetAuthorsPaginator() revel.Result {
	_, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		p.Response.Status = http.StatusUnauthorized
		return nil
	}
	searchData := new(models.SearchDataForProfiles)
	err = p.Params.BindJSON(&searchData)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	profiles, err := models.GetAuthorsWithSearchParams(*searchData)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusOK
	return p.RenderJSON(profiles)
}

func (p Profiles) GetMySubscribersPaginator() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		p.Response.Status = http.StatusUnauthorized
		return nil
	}
	searchData := new(models.SearchDataForProfiles)
	err = p.Params.BindJSON(&searchData)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	profiles, err := models.GetMySubscribersWithSearchParams(userID, *searchData)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusOK
	return p.RenderJSON(profiles)
}

func (p Profiles) GetMySubscribesPaginator() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		p.Response.Status = http.StatusUnauthorized
		return nil
	}
	searchData := new(models.SearchDataForProfiles)
	err = p.Params.BindJSON(&searchData)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	profiles, err := models.GetMySubscribesWithSearchParams(userID, *searchData)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusOK
	return p.RenderJSON(profiles)
}

func (p Profiles) AddSubscriberToProfile(id uint64) revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		p.Response.Status = http.StatusUnauthorized
		return nil
	}
	sUserID := fmt.Sprintf("%d", userID)
	err = models.AddSubscriberToProfile(userID, id)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	_ = models.DeleteDataFromRedis(sUserID)
	_ = models.DeleteDataFromRedis(fmt.Sprintf("%d", id))
	p.Response.Status = http.StatusNoContent
	return nil
}

func (p Profiles) DeleteSubscriberFromProfile(id uint64) revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		p.Response.Status = http.StatusUnauthorized
		return nil
	}
	sUserID := fmt.Sprintf("%d", userID)
	err = models.DeleteSubscriberFromProfile(userID, id)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	_ = models.DeleteDataFromRedis(sUserID)
	_ = models.DeleteDataFromRedis(fmt.Sprintf("%d", id))
	p.Response.Status = http.StatusNoContent
	return nil
}
