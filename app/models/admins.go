package models

type Admin struct {
	ID       uint64 `gorm:"primary_key;" json:"id"`
	Login    string `json:"login" gorm:"size:1000;not null;unique" validate:"omitempty,min=4,max=1000"`
	Password string `json:"password" gorm:"size:1000;not null" validate:"omitempty,min=8,max=1000"`
}

func GetAdminsDataByLogin(login string) (*Admin, error) {
	admin := new(Admin)
	ok := DB.Migrator().HasTable(admin)
	if !ok {
		err := DB.Migrator().CreateTable(admin)
		if err != nil {
			return nil, err
		}
	}
	result := DB.Where("login = ?", login).First(admin)
	if result.Error != nil {
		return nil, result.Error
	}
	return admin, nil
}
