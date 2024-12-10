package entity

import (
	"gorm.io/gorm"
)

type RivalInfo struct {
	gorm.Model
	Name         string
	ScoreLogPath *string
	SongDataPath *string
	PlayCount    int
}

func (RivalInfo) TableName() string {
	return "rival_info"
}
