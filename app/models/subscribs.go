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

func GetMySubscribersWithSearchParams(profileID uint64, searchData SearchDataForProfiles) ([]Profiles, error) {
	subscribers := make([]Profiles, 0)
	words := strings.Split(searchData.Stroke, " ")
	query := DB.Model(new(Profiles)).
		Joins("left join subscribs on subscribs.subscribers_id = profiles.id").
		Where("subscribs.profiles_id = ?", profileID).
		Preload("SubscribersList").
		Preload("MySubscribesList").
		Preload("Publications").
		Where("profiles.id >= ?", searchData.FirstID)

	for i := 0; searchData.Stroke != "" && i < len(words); i++ {
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
	err := query.Limit(searchData.Count).Find(&subscribers).Error
	return subscribers, err
}

func GetMySubscribesWithSearchParams(profileID uint64, searchData SearchDataForProfiles) ([]Profiles, error) {
	subscribers := make([]Profiles, 0)
	words := strings.Split(searchData.Stroke, " ")
	query := DB.Model(new(Profiles)).
		Joins("left join subscribs on subscribs.profiles_id = profiles.id").
		Where("subscribs.subscribers_id = ?", profileID).
		Preload("SubscribersList").
		Preload("MySubscribesList").
		Preload("Publications").
		Where("profiles.id >= ?", searchData.FirstID)

	for i := 0; searchData.Stroke != "" && i < len(words); i++ {
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
	err := query.Limit(searchData.Count).Find(&subscribers).Error
	return subscribers, err
}
