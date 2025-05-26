package controllers

import (
	"bytes"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/revel/revel"
	"io"
	"net/http"
	"os"
	"scinceHub/app/middleware"
	"scinceHub/app/models"
	"strconv"
	"strings"
	"time"
)

type Publications struct {
	*revel.Controller
}

type DeleteAuthorFromPublication struct {
	IDPublication uint64 `json:"id_publication"`
	IDAuthor      uint64 `json:"id_author"`
}

func (p Publications) CreatePublication() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		//p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
	sUserID := fmt.Sprintf("%d", userID)
	_ = models.DeleteDataFromRedis(sUserID)
	pub := new(models.Publications)
	pub.Title = p.Params.Get("title")
	pub.Abstract = p.Params.Get("abstract")
	pub.OwnerID = userID
	validate := validator.New()
	err = validate.Struct(pub)
	if err != nil {
		p.Response.Status = http.StatusUnprocessableEntity
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	fileHeader, ok := p.Params.Files["file"]
	if !ok || len(fileHeader) == 0 {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": "No file uploaded"})
	}

	file, err := fileHeader[0].Open()
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": "Не удалось открыть файл"})
	}
	defer file.Close()

	randomNumber, _ := GenerateRandomNumber()
	filePath := fmt.Sprintf("public/uploads/%d_%s_%s", userID, randomNumber, fileHeader[0].Filename)
	dst, err := os.Create(filePath)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": "Не удалось сохранить файл"})
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": "Ошибка при сохранении файла"})
	}

	pub.FileLink, err = models.PutFileInMINIO(filePath)
	if err != nil {
		revel.AppLog.Error(err.Error())
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	_ = os.Remove(filePath)
	unique := make(map[uint64]interface{}, 0)
	rawTagIDs := p.Params.Values["tags[]"]
	tagIDs := make([]uint64, 0)
	for _, ID := range rawTagIDs {
		tagID, _ := strconv.ParseUint(ID, 10, 64)
		_, ok := unique[tagID]
		if !ok {
			unique[tagID] = nil
			tagIDs = append(tagIDs, tagID)
		}
	}
	rawCoauthors := p.Params.Values["coauthors[]"]
	unique = make(map[uint64]interface{}, 0)
	coauthorIDs := make([]uint64, 0)
	coauthorIDs = append(coauthorIDs, userID)
	unique[userID] = nil
	for _, ID := range rawCoauthors {
		coauthorID, _ := strconv.ParseUint(ID, 10, 64)
		_, ok := unique[coauthorID]
		if !ok {
			unique[coauthorID] = nil
			coauthorIDs = append(coauthorIDs, coauthorID)
		}
	}
	err = models.CreatePublication(pub, tagIDs, coauthorIDs)

	if err != nil {
		revel.AppLog.Error(err.Error())
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	return p.Redirect("/profile")
}

func (p Publications) DeleteAuthorFromPublication() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		return p.Redirect("/login")
	}
	sUserID := fmt.Sprintf("%d", userID)
	_ = models.DeleteDataFromRedis(sUserID)
	dafp := new(DeleteAuthorFromPublication)
	p.Params.BindJSON(&dafp)
	fmt.Println(dafp)

	err = models.DeleteProfileFromPublication(dafp.IDPublication, dafp.IDAuthor)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Publications) GetPublicationData(id uint64) revel.Result {
	_, err := middleware.ValidateJWT(p.Request, "auth_token")
	_, err2 := middleware.ValidateAdminJWT(p.Request, "auth_token_admin")
	if err != nil && err2 != nil {
		p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
	pub, err := models.GetPublicationByID(id)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	return p.RenderJSON(pub)
}

func (p Publications) DeletePublication() revel.Result {
	_, err := middleware.ValidateJWT(p.Request, "auth_token")
	_, err2 := middleware.ValidateAdminJWT(p.Request, "auth_token_admin")
	if err != nil && err2 != nil {
		p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}

	pub := new(models.Publications)
	err = p.Params.BindJSON(pub)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	sUserID := fmt.Sprintf("%d", pub.OwnerID)
	_ = models.DeleteDataFromRedis(sUserID)

	err = models.DeletePublicationByID(pub.ID)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	err = models.RemoveFileFromMINIO(pub.FileLink)
	if err != nil {
		revel.AppLog.Error(err.Error())
	}

	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Publications) UpdatePublication() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	_, err2 := middleware.ValidateAdminJWT(p.Request, "auth_token_admin")
	if err != nil && err2 != nil {
		p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}

	pub := new(models.Publications)
	pubID := p.Params.Get("publication_id")
	pub.ID, err = strconv.ParseUint(pubID, 10, 64)
	pub.Title = p.Params.Get("title")
	pub.Abstract = p.Params.Get("abstract")
	ownerIDStr := p.Params.Get("owner_id")
	pub.OwnerID, err = strconv.ParseUint(ownerIDStr, 10, 64)
	pub.FileLink = p.Params.Get("fileLink")
	sUserID := ownerIDStr
	_ = models.DeleteDataFromRedis(sUserID)
	validate := validator.New()
	err = validate.Struct(pub)
	if err != nil {
		p.Response.Status = http.StatusUnprocessableEntity
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	if pub.OwnerID != userID && err2 != nil {
		p.Response.Status = http.StatusForbidden
		return p.RenderJSON(map[string]string{"error": "Вы не можете редактировать чужие публикации!"})
	}

	fileHeader, ok := p.Params.Files["file"]
	if ok && len(fileHeader) != 0 {
		file, err := fileHeader[0].Open()
		if err != nil {
			p.Response.Status = http.StatusInternalServerError
			return p.RenderJSON(map[string]string{"error": "Не удалось открыть файл"})
		}
		defer file.Close()
		randomNumber, _ := GenerateRandomNumber()
		filePath := fmt.Sprintf("public/uploads/%s_%s_%s", sUserID, randomNumber, fileHeader[0].Filename)
		dst, err := os.Create(filePath)
		if err != nil {
			p.Response.Status = http.StatusInternalServerError
			return p.RenderJSON(map[string]string{"error": "Не удалось сохранить файл"})
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			p.Response.Status = http.StatusInternalServerError
			return p.RenderJSON(map[string]string{"error": "Ошибка при сохранении файла"})
		}
		fileLink, err := models.PutFileInMINIO(filePath)
		if err != nil {
			p.Response.Status = http.StatusInternalServerError
			return p.RenderJSON(map[string]string{"error": "Ошибка при сохранении файла"})
		}
		err = models.RemoveFileFromMINIO(pub.FileLink)
		pub.FileLink = fileLink
		if err != nil {
			revel.AppLog.Error(err.Error())
		}
	}

	unique := make(map[uint64]interface{}, 0)
	rawTagIDs := p.Params.Values["tags[]"]
	tagIDs := make([]uint64, 0)
	for _, ID := range rawTagIDs {
		tagID, _ := strconv.ParseUint(ID, 10, 64)
		_, ok := unique[tagID]
		if !ok {
			unique[tagID] = nil
			tagIDs = append(tagIDs, tagID)
		}
	}
	rawCoauthors := p.Params.Values["coauthors[]"]
	unique = make(map[uint64]interface{}, 0)
	coauthorIDs := make([]uint64, 0)
	coauthorIDs = append(coauthorIDs, pub.OwnerID)
	unique[pub.OwnerID] = nil
	for _, ID := range rawCoauthors {
		coauthorID, _ := strconv.ParseUint(ID, 10, 64)
		_, ok := unique[coauthorID]
		if !ok {
			unique[coauthorID] = nil
			coauthorIDs = append(coauthorIDs, coauthorID)
		}
	}
	err = models.UpdatePublication(pub, tagIDs, coauthorIDs)

	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Publications) GetPublicationsData() revel.Result {
	pub, err := models.GetAllPublications()
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusOK
	return p.RenderJSON(pub)
}

func (p Publications) GetPublicationsPaginator() revel.Result {
	searchData := new(models.SearchDataForPublications)
	err := p.Params.BindJSON(searchData)
	if err != nil {
		p.Response.Status = http.StatusUnprocessableEntity
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	pubs, err := models.GetPublicationsWithSearchParams(*searchData)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusOK
	return p.RenderJSON(pubs)
}

func (p Publications) AddTagsToPublication(id uint64) revel.Result {
	var tagIDs []uint64
	err := p.Params.BindJSON(&tagIDs)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	err = models.AddTagsToPublication(id, tagIDs)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusNoContent
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Publications) DeleteTagsFromPublication(id uint64) revel.Result {
	var tagIDs []uint64
	err := p.Params.BindJSON(&tagIDs)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	err = models.DeleteTagsFromPublication(id, tagIDs)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusNoContent
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Publications) AddProfilesToPublication(id uint64) revel.Result {
	var profileIDs []uint64
	err := p.Params.BindJSON(&profileIDs)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	err = models.AddProfilesToPublication(id, profileIDs)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusNoContent
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Publications) GetFileWithPublicationList() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		//p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
	filters := new(models.PublicationDownloadFiltres)
	err = p.Params.BindJSON(filters)
	if err != nil {
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	filename, err := GetFileWithPublicationList(userID, *filters)
	if err != nil {
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	file, err := os.Open(filename)
	if err != nil {
		return p.RenderJSON(map[string]string{"error": "Ошибка при открытии файла"})
	}
	defer file.Close()

	array := strings.Split(filename, "/")
	name := array[len(array)-1]
	fmt.Println(name)

	fileData, err := io.ReadAll(file)
	if err != nil {
		return p.RenderJSON(map[string]string{"error": "Ошибка при чтении файла"})
	}
	reader := bytes.NewReader(fileData)

	go func() {
		time.Sleep(10 * time.Second)
		_ = os.Remove(filename)
	}()

	modTime := time.Now()

	return p.RenderBinary(reader, name, revel.Attachment, modTime)
}
