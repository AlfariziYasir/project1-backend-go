package models

import (
	"errors"
	"html"
	"strings"

	"gorm.io/gorm"
)

type Report struct {
	gorm.Model
	UserId    string `gorm:"not null" json:"user_id"`
	MeetingId uint   `gorm:"not null" json:"post_id"`
	Body      string `gorm:"text;not null" json:"body"`
	User      User   `gorm:"ForeignKey:UserId" json:"user"`
}

func (r *Report) Prepare() {
	r.UserId = html.EscapeString(strings.TrimSpace(r.UserId))
	r.Body = html.EscapeString(strings.TrimSpace(r.Body))
	r.User = User{}
}

func (r *Report) Validate(action string) map[string]string {
	var errMsg = make(map[string]string)
	var err error

	switch strings.ToLower(action) {
	case "update":
		if r.Body == "" {
			err = errors.New("required report comment")
			errMsg["Required_body"] = err.Error()
		}
		if r.UserId == "" {
			err = errors.New("required User ID")
			errMsg["Required_user_id"] = err.Error()
		}
	default:
		if r.Body == "" {
			err = errors.New("required report comment")
			errMsg["Required_body"] = err.Error()
		}
		if r.UserId == "" {
			err = errors.New("required User ID")
			errMsg["Required_user_id"] = err.Error()
		}
	}
	return errMsg
}

func (r *Report) SaveReport(db *gorm.DB) (*Report, error) {
	err := db.Debug().Model(&Report{}).Create(&r).Error
	if err != nil {
		return &Report{}, err
	}

	err = db.Debug().Preload("User").Where(&Report{UserId: r.UserId}).First(&r).Error
	if err != nil {
		return &Report{}, err
	}

	return r, nil
}

func (r *Report) GetReports(db *gorm.DB, mID uint) (*[]Report, error) {
	rs := []Report{}
	err := db.Debug().Preload("User").Where(&Report{MeetingId: mID}).Order("created_at desc").Find(&rs).Error
	if err != nil {
		return &[]Report{}, err
	}

	return &rs, nil
}

func (r *Report) UpdateReport(db *gorm.DB) (*Report, error) {
	err := db.Debug().Model(&Report{}).Where("id = ?", r.ID).Update("body", r.Body).Error
	if err != nil {
		return &Report{}, err
	}

	err = db.Debug().Model(&User{}).Where("user_id = ?", r.UserId).Take(&r.User).Error
	if err != nil {
		return &Report{}, err
	}

	return r, nil
}

func (r *Report) DeleteReport(db *gorm.DB, uid string) (int64, error) {
	db = db.Debug().Model(&Report{}).Where("user_id = ?", uid).Take(&Report{}).Delete(&Report{})
	if db.Error != nil {
		return 0, db.Error
	}

	return db.RowsAffected, nil
}
