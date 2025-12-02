package controllers

import (
	"github.com/revel/revel"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"scinceHub/app/middleware"
	"scinceHub/app/models"
)

type Admins struct {
	*revel.Controller
}

func (a Admins) LoginAdmin(login, password string) revel.Result {
	admin, err := models.GetAdminsDataByLogin(login)
	if err != nil {
		a.Response.Status = http.StatusUnauthorized
		return a.RenderTemplate("login_admin.html")
	}
	err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password))
	if err != nil {
		a.Response.Status = http.StatusUnauthorized
		return a.RenderTemplate("login_admin.html")
	}
	token, err := middleware.GenerateAdminJWT(admin.ID)

	if err != nil {
		a.Response.Status = http.StatusInternalServerError
		return a.RenderText("Ошибка генерации токена")
	}
	middleware.SetCookieData(a.Controller, "auth_token_admin", token, false)

	a.Response.Status = http.StatusFound
	return a.Redirect("/admin")
}

func (a Admins) LogoutAdmin() revel.Result {
	middleware.SetCookieData(a.Controller, "auth_token_admin", "", true)
	return a.Redirect("/")
}
