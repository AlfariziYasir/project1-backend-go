package models

import (
	"html"
	"strings"

	"gorm.io/gorm"
)

type ResetPassword struct {
	gorm.Model
	Email string `gorm:"size:100;not null" json:"email"`
	Token string `gorm:"size:255;not null" json:"token"`
}

func (rp *ResetPassword) Prepare() {
	rp.Email = html.EscapeString(strings.TrimSpace(rp.Email))
	rp.Token = html.EscapeString(strings.TrimSpace(rp.Token))
}

func (rp *ResetPassword) SaveData(db *gorm.DB) (*ResetPassword, error) {
	err := db.Debug().Model(&ResetPassword{}).Create(&rp).Error
	if err != nil {
		return &ResetPassword{}, err
	}

	return rp, nil
}

func (rp *ResetPassword) DeleteData(db *gorm.DB) (int64, error) {
	db = db.Debug().Model(&ResetPassword{}).Where("id = ?", rp.ID).Take(&ResetPassword{}).Delete(&ResetPassword{})
	if db.Error != nil {
		return 0, db.Error
	}

	return db.RowsAffected, nil
}
