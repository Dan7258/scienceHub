package models

import (
	"scinceHub/app/middleware"
	"strconv"
	"strings"
)

type Subscribs struct {
	ProfilesId    uint64 `json:"profiles_id" gorm:"primaryKey;not null"`
	SubscribersID uint64 `json:"subscribers_id" gorm:"primaryKey;not null"`
}

func CheckMySubscribesForProfile(ID uint64, profileID uint64) bool {
	sub := new(Subscribs)
	result := DB.Where("profiles_id = ? AND subscribers_id = ?", profileID, ID).First(sub)
	if result.RowsAffected == 0 || result.Error != nil {
		return false
	}
	return true
}

func AddSubscriberToProfile(subID uint64, profileID uint64) error {
	sub := new(Subscribs)
	sub.ProfilesId = profileID
	sub.SubscribersID = subID
	result := DB.Model(sub).Create(sub)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func DeleteSubscriberFromProfile(subID uint64, profileID uint64) error {
	sub := new(Subscribs)
	sub.ProfilesId = profileID
	sub.SubscribersID = subID
	result := DB.Model(sub).Delete(sub)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func GetMySubscribersWithSearchParams(profileID uint64, data SearchDataForProfiles) (GetSearchingDataFromProfiles, error) {
	searchData := new(GetSearchingDataFromProfiles)
	searchData.Data = make([]Profiles, 0)
	words := strings.Split(data.Stroke, " ")
	var count int64
	query := DB.Model(new(Profiles)).
		Joins("left join subscribs on subscribs.subscribers_id = profiles.id").
		Where("subscribs.profiles_id = ?", profileID).
		Preload("SubscribersList").
		Preload("MySubscribesList").
		Preload("Publications")
	for i := 0; data.Stroke != "" && i < len(words); i++ {
		likeword := "%" + words[i] + "%"

		if i == 0 {
			query.Where("profiles.first_name ILIKE ? OR profiles.last_name ILIKE ? OR profiles.middle_name ILIKE ?", likeword, likeword, likeword)
		} else {
			query.Or("profiles.first_name ILIKE ? OR profiles.last_name ILIKE ? OR profiles.middle_name ILIKE ?", likeword, likeword, likeword)
		}
		if middleware.IsInteger(words[i]) {
			id, _ := strconv.Atoi(words[i])
			query.Or("profiles.id = ?", id)
		}
	}
	switch data.Sort {
	case SortNameAsc:
		query = query.Order("profiles.last_name ASC")
	default:
		query = query.Order("profiles.last_name DESC")
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

func GetMySubscribesWithSearchParams(profileID uint64, data SearchDataForProfiles) (GetSearchingDataFromProfiles, error) {
	searchData := new(GetSearchingDataFromProfiles)
	searchData.Data = make([]Profiles, 0)
	words := strings.Split(data.Stroke, " ")
	var count int64
	query := DB.Model(new(Profiles)).
		Joins("left join subscribs on subscribs.profiles_id = profiles.id").
		Where("subscribs.subscribers_id = ?", profileID).
		Preload("SubscribersList").
		Preload("MySubscribesList").
		Preload("Publications")
	for i := 0; data.Stroke != "" && i < len(words); i++ {
		likeword := "%" + words[i] + "%"

		if i == 0 {
			query.Where("profiles.first_name ILIKE ? OR profiles.last_name ILIKE ? OR profiles.middle_name ILIKE ?", likeword, likeword, likeword)
		} else {
			query.Or("profiles.first_name ILIKE ? OR profiles.last_name ILIKE ? OR profiles.middle_name ILIKE ?", likeword, likeword, likeword)
		}
		if middleware.IsInteger(words[i]) {
			id, _ := strconv.Atoi(words[i])
			query.Or("profiles.id = ?", id)
		}
	}
	switch data.Sort {
	case SortNameAsc:
		query = query.Order("profiles.last_name ASC")
	default:
		query = query.Order("profiles.last_name DESC")
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
