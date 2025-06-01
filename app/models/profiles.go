package models

import (
	"fmt"
	"gorm.io/gorm"
	"os"
	"scinceHub/app/middleware"
	"strconv"
	"strings"
)

type Profiles struct {
	ID               uint64         `json:"id" gorm:"primaryKey"`
	Login            string         `json:"login" gorm:"size:1000;not null;unique" validate:"omitempty,email,min=4,max=1000"`
	Password         string         `json:"password" gorm:"size:1000;not null" validate:"omitempty,min=8,max=1000"`
	FirstName        string         `json:"first_name" gorm:"size:1000;not null" validate:"omitempty,min=3,max=1000"`
	LastName         string         `json:"last_name" gorm:"size:1000;not null" validate:"omitempty,min=3,max=1000"`
	MiddleName       string         `json:"middle_name" gorm:"size:1000;" validate:"max=1000"`
	Gender           int            `json:"gender" validate:"omitempty,min=1,max=3"`
	Country          string         `json:"country" gorm:"size:100;" validate:"max=100"`
	AcademicDegree   string         `json:"academic_degree" gorm:"size:1000;" validate:"max=1000"`
	VAC              string         `json:"vac" gorm:"size:1000;" validate:"max=1000"`
	Appointment      string         `json:"appointment" gorm:"size:1000;" validate:"max=1000"`
	Publications     []Publications `gorm:"many2many:profile_publications;"`
	SubscribersList  []Profiles     `gorm:"many2many:subscribs;joinForeignKey:profiles_id;joinReferences:subscribers_id"`
	MySubscribesList []Profiles     `gorm:"many2many:subscribs;joinForeignKey:subscribers_id;joinReferences:profiles_id"`
}

type GetSearchingDataFromProfiles struct {
	Data     []Profiles `json:"data"`
	MaxPages int64      `json:"max_pages"`
}

type ProfileWithSubscribitionStatus struct {
	Profile      Profiles
	Isubscribed  bool
	IsSubscribed bool
}

type SearchDataForProfiles struct {
	Stroke string `json:"stroke"`
	Paginator
}

func CreateProfile(profile *Profiles) error {
	profile.Gender = 3
	result := DB.Create(profile)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func GetProfileByID(ID uint64) (*Profiles, error) {
	profile := new(Profiles)
	result := DB.Select("id, first_name, last_name, middle_name, gender, academic_degree, country, vac, appointment").
		Preload("Publications.Profiles", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, first_name, last_name, middle_name")
		}).
		Preload("Publications.Tags").
		Preload("SubscribersList", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, first_name, last_name, middle_name")
		}).
		Preload("MySubscribesList", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, first_name, last_name, middle_name")
		}).
		First(profile, ID)
	if result.Error != nil {
		return nil, result.Error
	}
	return profile, nil
}

func ThsProfilesIsExist(login string) bool {
	profile := new(Profiles)
	result := DB.Select("login").Where("login = ?", login).First(profile)
	if result.Error != nil {
		return false
	}
	return true
}

func GetUserProfile(ID uint64) (*Profiles, error) {
	profile := new(Profiles)
	result := DB.Preload("Publications", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at desc")
	}).
		Preload("Publications.Profiles", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, first_name, last_name, middle_name")
		}).
		Preload("Publications.Tags").
		Preload("SubscribersList", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, first_name, last_name, middle_name")
		}).
		Preload("MySubscribesList", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, first_name, last_name, middle_name")
		}).
		First(profile, ID)
	if result.Error != nil {
		return nil, result.Error
	}
	return profile, nil
}

func GetProfileLoginData(login string) (*Profiles, error) {
	profile := new(Profiles)
	result := DB.Select("id, login, password").Where("login = ?", login).First(profile)
	if result.Error != nil {
		return nil, result.Error
	}
	return profile, nil
}

func GetAllProfiles() ([]Profiles, error) {
	var profiles []Profiles
	result := DB.Select("id, first_name, last_name, middle_name, country, academic_degree, vac, appointment").
		Preload("Publications").
		Preload("SubscribersList").
		Preload("MySubscribesList").
		Find(&profiles)

	if result.Error != nil {
		return nil, result.Error
	}
	return profiles, nil
}

func GetAuthorsWithSearchParams(data SearchDataForProfiles) (GetSearchingDataFromProfiles, error) {
	searchData := new(GetSearchingDataFromProfiles)
	searchData.Data = make([]Profiles, 0)
	var count int64
	query := DB.Model(new(Profiles)).Select("id, first_name, last_name, middle_name, gender, country, academic_degree, vac, appointment").
		Preload("Publications").
		Preload("SubscribersList").
		Preload("MySubscribesList")

	words := strings.Split(data.Stroke, " ")
	for i := 0; data.Stroke != "" && i < len(words); i++ {
		likeword := "%" + words[i] + "%"

		if i == 0 {
			query.Where("first_name ILIKE ? OR last_name ILIKE ? OR middle_name ILIKE ?", likeword, likeword, likeword)
		} else {
			query.Or("first_name ILIKE ? OR last_name ILIKE ? OR middle_name ILIKE ?", likeword, likeword, likeword)
		}
		if middleware.IsInteger(words[i]) {
			id, _ := strconv.Atoi(words[i])
			query.Or("id = ?", id)
		}
	}
	switch data.Sort {
	case SortNameAsc:
		query = query.Order("last_name ASC")
	default:
		query = query.Order("last_name DESC")
	}
	if query.Error != nil {
		return *searchData, query.Error
	}
	query.Count(&count)
	data.Page--
	if data.Page < 0 {
		data.Page = 0
	}
	err := query.Offset(data.Page * data.Count).Limit(data.Count).Find(&searchData.Data).Error
	searchData.MaxPages = count / int64(data.Count)
	if count%int64(data.Count) > 0 {
		searchData.MaxPages++
	}
	return *searchData, err
}

func GetAllProfileIDAndNames() ([]Profiles, error) {
	var profiles []Profiles
	result := DB.Select("id, first_name, last_name, middle_name").Find(&profiles)
	if result.Error != nil {
		return nil, result.Error
	}
	return profiles, nil
}

func GetProfileNameByID(id uint64) (Profiles, error) {
	var profile Profiles
	result := DB.Select("id, first_name, last_name, middle_name, gender").First(&profile, id)
	if result.Error != nil {
		return profile, result.Error
	}
	return profile, nil
}

func DeleteProfileByID(ID uint64) error {
	profile := new(Profiles)
	err := DB.Model(new(Subscribs)).Where("profiles_id = ? OR subscribers_id = ?", ID, ID).Delete(new(Subscribs)).Error
	if err != nil {
		return err
	}
	var publications []Publications
	_ = DB.Model(new(Publications)).Where("owner_id = ?", ID).Find(&publications)
	err = DB.Model(new(Publications)).Where("owner_id = ?", ID).Delete(new(Publications)).Error
	if err != nil {
		return err
	}
	result := DB.Delete(profile, ID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("Профиль с ID %d не найден", ID)
	}
	for _, publication := range publications {
		_ = RemoveFileFromMINIO(publication.FileLink)
	}
	return nil
}

func RemoveFilesByUserID(ID uint64) error {
	dir := "public/uploads/"
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		f := strings.Split(file.Name(), "_")
		if !file.IsDir() && f[0] == fmt.Sprintf("%d", ID) {
			os.Remove(dir + file.Name())
		}
	}
	return nil
}

func DeleteProfileByLogin(login string) error {
	profile := new(Profiles)
	result := DB.Delete(profile, login)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("Профиль с login: %s не найден", login)
	}
	return nil
}

func UpdateProfileByID(ID uint64, updProfile *Profiles) error {
	result := DB.Model(new(Profiles)).Where("id = ?", ID).Updates(updProfile)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("Профиль с ID %d не найден", ID)
	}
	return nil
}

func UpdateProfileByLogin(login string, updProfile *Profiles) error {
	result := DB.Model(new(Profiles)).Where("login = ?", login).Updates(updProfile)
	if result.Error != nil {
		return result.Error
	}
	//if result.RowsAffected == 0 {
	//	return fmt.Errorf("Профиль с login: %s не найден", login)
	//}
	return nil
}

func AddPublicationsToProfile(ID uint64, pubIDs []uint64) error {
	profile := new(Profiles)
	var pubs []Publications
	result := DB.First(profile, ID)
	if result.Error != nil {
		return result.Error
	}
	result = DB.Find(&pubs, pubIDs)
	if result.Error != nil {
		return result.Error
	}
	err := DB.Model(profile).Association("Publications").Append(pubs)
	if err != nil {
		return err
	}
	return nil
}

func DeletePublicationsFromProfile(ID uint64, pubIDs []uint64) error {
	profile := new(Profiles)
	var pubs []Publications
	result := DB.First(profile, ID)
	if result.Error != nil {
		return result.Error
	}
	result = DB.Find(&pubs, pubIDs)
	if result.Error != nil {
		return result.Error
	}
	err := DB.Model(profile).Association("Publications").Delete(pubs)
	if err != nil {
		return err
	}
	return nil
}
