package models

import (
	"fmt"
)

type Tags struct {
	ID           uint64         `json:"id" gorm:"primaryKey"`
	Name         string         `json:"name" gorm:"size:1000;not null" validate:"required,min=2,max=1000"`
	Publications []Publications `json:"publications" gorm:"many2many:publication_tags;"`
}

func CreateTag(name string) error {
	tag := new(Tags)
	tag.Name = name
	result := DB.Create(tag)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func GetTagByID(ID int) (*Tags, error) {
	tag := new(Tags)
	result := DB.Preload("Publications").First(tag, ID)
	if result.Error != nil {
		return nil, result.Error
	}
	return tag, nil
}

func GetTagsByID(IDList []uint64) ([]Tags, error) {
	tags := make([]Tags, 0)
	tag := new(Tags)
	result := DB.Model(tag).Find(&tags, IDList)
	return tags, result.Error
}

func GetTagByName(name string) (*Tags, error) {
	tag := new(Tags)
	result := DB.Preload("Publications").Where("name = ?", name).First(tag)
	if result.Error != nil {
		return nil, result.Error
	}
	return tag, nil
}

func DeleteTagByName(name string) error {
	tag := new(Tags)
	result := DB.Where("name = ?", name).Delete(tag)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("Тег с name %s не найден", name)
	}
	return nil
}

func UpdateTagByID(name string, updTag *Tags) error {
	tag := new(Tags)
	result := DB.Preload("Publications").Model(tag).Where("name = ?", name).Update("name", updTag.Name)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("Тег с name %s не найден", name)
	}
	return nil
}

func GetAllTags() ([]Tags, error) {
	var tags []Tags
	result := DB.Preload("Publications").Find(&tags)
	if result.Error != nil {
		return nil, result.Error
	}
	return tags, nil
}

func AddPublicationsToTag(ID uint64, pubIDs []uint64) error {
	tag := new(Tags)
	var pubs []Publications
	result := DB.First(tag, ID)
	if result.Error != nil {
		return result.Error
	}
	result = DB.Find(&pubs, pubIDs)
	if result.Error != nil {
		return result.Error
	}
	err := DB.Model(tag).Association("Publications").Append(pubs)
	if err != nil {
		return err
	}
	return nil
}

func DeletePublicationsFromTag(ID uint64, pubIDs []uint64) error {
	tag := new(Tags)
	var pubs []Publications
	result := DB.First(tag, ID)
	if result.Error != nil {
		return result.Error
	}
	result = DB.Find(&pubs, pubIDs)
	if result.Error != nil {
		return result.Error
	}
	err := DB.Model(tag).Association("Publications").Delete(pubs)
	if err != nil {
		return err
	}
	return nil
}
