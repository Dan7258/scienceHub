package middleware

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/revel/revel"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"strconv"
	"time"
)

func InitENV() {
	err := godotenv.Load(".env")
	if err != nil {
		revel.AppLog.Error("SECRET_KEY environment variable not set")
		return
	}
	if len([]byte(os.Getenv("SECRET_KEY"))) == 0 {
		revel.AppLog.Error("SECRET_KEY environment variable not set")
		return
	}
}

func GenerateJWT(userID uint64) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = userID
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET_KEY")))

	if err != nil {
		fmt.Errorf("Something Went Wrong: %s", err.Error())
		return "", err
	}

	return tokenString, nil

}

func ValidateJWT(request *revel.Request, cookieName string) (uint64, error) {
	cookie, err := request.Cookie(cookieName)
	if err != nil {
		return 0, err
	}
	token, err := jwt.Parse(cookie.GetValue(), func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("There was an error in parsing the token")
		}
		return []byte(os.Getenv("SECRET_KEY")), nil
	})
	if err != nil {
		return 0, err
	}
	exp := token.Claims.(jwt.MapClaims)["exp"].(float64)
	userID := uint64(token.Claims.(jwt.MapClaims)["sub"].(float64))

	if time.Until(time.Unix(int64(exp), 0)) < time.Minute*30 {
		GenerateJWT(userID)
	}

	return userID, nil
}

func GenerateAdminJWT(userID uint64) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = userID
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET_KEY_ADMIN")))

	if err != nil {
		fmt.Errorf("Something Went Wrong: %s", err.Error())
		return "", err
	}

	return tokenString, nil

}

func ValidateAdminJWT(request *revel.Request, cookieName string) (uint64, error) {
	cookie, err := request.Cookie(cookieName)
	if err != nil {
		return 0, err
	}
	token, err := jwt.Parse(cookie.GetValue(), func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("There was an error in parsing the token")
		}
		return []byte(os.Getenv("SECRET_KEY_ADMIN")), nil
	})
	if err != nil {
		return 0, err
	}
	exp := token.Claims.(jwt.MapClaims)["exp"].(float64)
	userID := uint64(token.Claims.(jwt.MapClaims)["sub"].(float64))

	if time.Until(time.Unix(int64(exp), 0)) < time.Minute*30 {
		GenerateAdminJWT(userID)
	}

	return userID, nil
}

func SetCookieData(ctrl *revel.Controller, cookieName, cookieValue string, remove bool) {
	cookie := &http.Cookie{}
	cookie.Name = cookieName
	cookie.Value = cookieValue
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Secure = true
	cookie.SameSite = http.SameSiteStrictMode

	if remove {
		cookie.Expires = time.Unix(0, 0)
		cookie.MaxAge = -1
	}
	ctrl.SetCookie(cookie)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func IsInteger(stroke string) bool {
	_, err := strconv.Atoi(stroke)
	return err == nil
}
