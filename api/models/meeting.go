package models

import (
	"errors"
	"html"
	"strings"

	"gorm.io/gorm"
)

type Meeting struct {
	gorm.Model
	UserId  string `gorm:"not null"`
	Title   string `gorm:"size:255;not null" json:"title"`
	Content string `gorm:"text;not null" json:"content"`
	Created User   `gorm:"foreignKey:UserId;references:UserId;"`
}

func (m *Meeting) Prepare(uID string) {
	m.UserId = uID
	m.Title = html.EscapeString(strings.TrimSpace(m.Title))
	m.Content = html.EscapeString(strings.TrimSpace(m.Content))
	m.Created = User{}
}

func (m *Meeting) Validate() map[string]string {
	var err error
	var errMsg = make(map[string]string)

	if m.Title == "" {
		err = errors.New("required Title")
		errMsg["Required_title"] = err.Error()
	} else if m.Content == "" {
		err = errors.New("required Content")
		errMsg["Required_content"] = err.Error()
	}
	return errMsg
}

func (m *Meeting) SaveMeeting(db *gorm.DB) (*Meeting, error) {
	err := db.Debug().Create(&m).Error
	if err != nil {
		return &Meeting{}, err
	}

	err = db.Debug().Preload("Created").Where("id = ?", m.ID).Find(&m).Error
	if err != nil {
		return &Meeting{}, err
	}

	return m, nil
}

func (m *Meeting) GetMeetings(db *gorm.DB) (*[]Meeting, error) {
	meetings := []Meeting{}
	err := db.Debug().Preload("Created").Order("created_at desc").Find(&meetings).Error
	if err != nil {
		return &[]Meeting{}, err
	}

	return &meetings, err
}

func (m *Meeting) GetUserMeetings(db *gorm.DB, id string) (*[]Meeting, error) {
	meetings := []Meeting{}
	err := db.Debug().Preload("Created").Where("user_id = ?", id).Find(&meetings).Error
	if err != nil {
		return &[]Meeting{}, err
	}

	return &meetings, nil
}

func (m *Meeting) GetMeeting(db *gorm.DB, id uint) (*Meeting, error) {
	err := db.Debug().Preload("Created").Where("id = ?", id).Take(&m).Error
	if err != nil {
		return &Meeting{}, err
	}

	return m, nil
}

func (m *Meeting) UpdateMeeting(db *gorm.DB) (*Meeting, error) {
	err := db.Debug().Model(&Meeting{}).Where("id = ?", m.ID).Updates(Meeting{
		Title:   m.Title,
		Content: m.Content,
	}).Error
	if err != nil {
		return &Meeting{}, err
	}

	err = db.Debug().Preload("Created").Where("id = ?", m.ID).Take(&m).Error
	if err != nil {
		return &Meeting{}, err
	}

	return m, nil
}

func (m *Meeting) DeleteMeeting(db *gorm.DB) (int64, error) {
	db = db.Debug().Model(&Meeting{}).Where("id = ?", m.ID).Take(&Meeting{}).Delete(&Meeting{})
	if db.Error != nil {
		return 0, db.Error
	}

	return db.RowsAffected, nil
}

func (m *Meeting) DeleteUserMeeting(db *gorm.DB, id string) (int64, error) {
	meetings := []Meeting{}
	db = db.Debug().Model(&Meeting{}).Where("user_id = ?", id).Find(&meetings).Delete(&meetings)
	if db.Error != nil {
		return 0, db.Error
	}

	return db.RowsAffected, nil
}
