package models

import (
	"errors"

	"gorm.io/gorm"
)

type Join struct {
	gorm.Model
	UserId    string `gorm:"not null" json:"user_id"`
	MeetingId uint   `gorm:"not null" json:"meeting_id"`
}

func (j *Join) SaveJoin(db *gorm.DB) (*Join, error) {
	// check if the user has joined meeting before
	err := db.Debug().Model(&Join{}).Where("user_id = ? AND meeting_id = ?", j.UserId, j.MeetingId).Take(&j).Error
	if err != nil {
		if err.Error() == "record not found" {
			// if the user has not joined meeting before, lets save new join
			err = db.Debug().Model(&Join{}).Create(&j).Error
			if err != nil {
				return &Join{}, err
			}
		}
	} else {
		// if the user has joined meeting before, create a error message
		err = errors.New("double join")
		return &Join{}, err
	}

	return j, nil
}

func (j *Join) DeleteJoin(db *gorm.DB) (*Join, error) {
	var dj *Join

	err := db.Debug().Model(&Join{}).Where("id = ?", j.ID).Take(&j).Error
	if err != nil {
		return &Join{}, err
	} else {
		dj = j
		db = db.Debug().Model(&Join{}).Where("id = ?", j.ID).Take(&Join{}).Delete(&Join{})
		if db.Error != nil {
			return &Join{}, err
		}
	}
	return dj, nil
}

func (j *Join) GetJoinInfo(db *gorm.DB, mID uint) (*[]Join, count int, error) {
	joins := []Join{}

	err := db.Debug().Model(&Join{}).Where("meeting_id = ?", mID).Find(&joins).Error
	if err != nil {
		return &[]Join{}, err
	}
	
	err = db.Debug.Model(&Join{}).Where("meeting_id = ?", mID).Count(&count).Error

	return &joins, count, nil
}

func (j *Join) DeleteUserJoin(db *gorm.DB, uid string) (int64, error) {
	joins := []Join{}
	db = db.Debug().Model(&Join{}).Where("user_id = ?", uid).Find(&joins).Delete(&joins)
	if db.Error != nil {
		return 0, db.Error
	}

	return db.RowsAffected, nil
}

func (j *Join) DeleteMeetingJoin(db *gorm.DB, mID uint) (int64, error) {
	joins := []Join{}
	db = db.Debug().Model(&Join{}).Where("meeting_id = ?", mID).Find(&joins).Delete(&joins)
	if db.Error != nil {
		return 0, db.Error
	}

	return db.RowsAffected, nil
}
