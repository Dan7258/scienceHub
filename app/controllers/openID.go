package controllers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/revel/revel"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"scinceHub/app/middleware"
	"scinceHub/app/models"
	"scinceHub/app/smtp"
	"strings"
)

type OpenID struct {
	*revel.Controller
}

type AccessToken struct {
	AccessToken string `json:"access_token"`
}

type UserDataYA struct {
	DefaultEmail string `json:"default_email"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
}

type UserDataVK struct {
	User UserVK `json:"user"`
}

type UserVK struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (o OpenID) GetCodeForYAToken() revel.Result {
	accessToken := new(AccessToken)
	err := o.Params.BindJSON(accessToken)
	if err != nil {
		return o.RenderJSON(map[string]string{"error": err.Error()})
	}
	req, err := http.NewRequest("GET", "https://login.yandex.ru/info?format=json", nil)
	if err != nil {
		return o.RenderJSON(map[string]string{"error": err.Error()})
	}
	req.Header.Set("Authorization", "OAuth "+accessToken.AccessToken)
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return o.RenderJSON(map[string]string{"error": err.Error()})
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return o.RenderJSON(map[string]string{"error": fmt.Sprintf("API error: %s", resp.Status)})
	}
	userdata := new(UserDataYA)

	err = json.NewDecoder(resp.Body).Decode(userdata)
	if err != nil {
		return o.RenderJSON(map[string]string{"error": "Failed to read response: " + err.Error()})
	}
	if !models.ThsProfilesIsExist(userdata.DefaultEmail) {
		err = o.RegisterWithYA(*userdata)
	}
	if err != nil {
		return o.RenderJSON(map[string]string{"error": err.Error()})
	}
	err = o.LoginWithOpenID(userdata.DefaultEmail)
	if err != nil {
		return o.RenderJSON(map[string]string{"error": err.Error()})
	}

	return o.RenderJSON(map[string]int{"status": http.StatusOK})

}

func (o OpenID) RegisterWithYA(userdata UserDataYA) error {
	profile := new(models.Profiles)
	profile.Login = userdata.DefaultEmail
	profile.FirstName = userdata.FirstName
	profile.LastName = userdata.LastName
	pass := o.GeneratePassword(12)
	hpass, err := middleware.HashPassword(pass)
	if err != nil {
		return err
	}
	profile.Password = hpass
	err = models.CreateProfile(profile)
	if err != nil {
		return err
	}
	message := fmt.Sprintf("Добро пожаловать на scienceHub, вот ваш пароль: %s. Пожалуйста, смените пароль при любой удобно возможности", pass)
	err = smtp.SendMessage(profile.Login, "Пароль от учетной записи", message)
	if err != nil {
		return err
	}
	return nil
}

func (o OpenID) GetCodeForVKToken() revel.Result {
	accessToken := new(AccessToken)
	clientIDVK := "53573299"
	err := o.Params.BindJSON(accessToken)
	if err != nil {
		return o.RenderJSON(map[string]string{"error": err.Error()})
	}
	formData := url.Values{
		"client_id":    {clientIDVK},
		"access_token": {accessToken.AccessToken},
	}

	req, err := http.NewRequest("POST", "https://id.vk.com/oauth2/user_info", strings.NewReader(formData.Encode()))
	if err != nil {
		return o.RenderJSON(map[string]string{"error": err.Error()})
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return o.RenderJSON(map[string]string{"error": err.Error()})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return o.RenderJSON(map[string]string{"error": fmt.Sprintf("API error: %s", resp.Status)})
	}
	userdata := new(UserDataVK)
	body, _ := io.ReadAll(resp.Body)

	fmt.Println(string(body))
	fmt.Println(userdata)
	err = json.NewDecoder(resp.Body).Decode(userdata)
	if err != nil {
		return o.RenderJSON(map[string]string{"error": "Failed to read response: " + err.Error()})
	}
	if !models.ThsProfilesIsExist(userdata.User.Email) {
		err = o.RegisterWithVK(*userdata)
	}
	if err != nil {
		return o.RenderJSON(map[string]string{"error": err.Error()})
	}
	err = o.LoginWithOpenID(userdata.User.Email)
	if err != nil {
		return o.RenderJSON(map[string]string{"error": err.Error()})
	}

	return o.RenderJSON(map[string]int{"status": http.StatusOK})

}

func (o OpenID) RegisterWithVK(userdata UserDataVK) error {
	profile := new(models.Profiles)
	profile.Login = userdata.User.Email
	profile.FirstName = userdata.User.FirstName
	profile.LastName = userdata.User.LastName
	pass := o.GeneratePassword(12)
	hpass, err := middleware.HashPassword(pass)
	if err != nil {
		return err
	}
	profile.Password = hpass
	err = models.CreateProfile(profile)
	if err != nil {
		return err
	}
	message := fmt.Sprintf("Добро пожаловать на scienceHub, вот ваш пароль: %s. Пожалуйста, смените пароль при любой удобно возможности", pass)
	err = smtp.SendMessage(profile.Login, "Пароль от учетной записи", message)
	if err != nil {
		return err
	}
	return nil
}

func (o OpenID) LoginWithOpenID(email string) error {
	user, err := models.GetProfileLoginData(email)
	if err != nil {
		return err
	}

	token, err := middleware.GenerateJWT(user.ID)
	if err != nil {
		return err
	}
	middleware.SetCookieData(o.Controller, "auth_token", token, false)

	return nil
}

func (o OpenID) GeneratePassword(length uint64) string {
	data := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+"
	password := ""
	max := big.NewInt(int64(len(data)))
	for i := 0; i < int(length); i++ {
		randomNumber, _ := rand.Int(rand.Reader, max)
		index := int(randomNumber.Int64())
		password += string(data[index])
	}
	return password

}
