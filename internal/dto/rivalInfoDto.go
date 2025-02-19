package dto

import (
	"github.com/Catizard/lampghost_wails/internal/entity"
	"gorm.io/gorm"
)

type RivalInfoDto struct {
	gorm.Model

	Name         string
	ScoreLogPath *string
	SongDataPath *string
	PlayCount    int
	MainUser     bool

	DiffTableHeader *DiffTableHeaderDto
}

func NewRivalInfoDto(rival *entity.RivalInfo) *RivalInfoDto {
	return &RivalInfoDto{
		Model: gorm.Model{
			ID:        rival.ID,
			CreatedAt: rival.CreatedAt,
			UpdatedAt: rival.UpdatedAt,
			DeletedAt: rival.DeletedAt,
		},
		Name:         rival.Name,
		ScoreLogPath: rival.ScoreLogPath,
		SongDataPath: rival.SongDataPath,
		PlayCount:    rival.PlayCount,
		MainUser:     rival.MainUser,
	}
}

func (rival *RivalInfoDto) Entity() *entity.RivalInfo {
	return &entity.RivalInfo{
		Name:         rival.Name,
		PlayCount:    rival.PlayCount,
		ScoreLogPath: rival.ScoreLogPath,
		SongDataPath: rival.SongDataPath,
		MainUser:     rival.MainUser,
	}
}

func NewRivalInfoDtoWithDiffTable(rival *entity.RivalInfo, header *DiffTableHeaderDto) *RivalInfoDto {
	ret := NewRivalInfoDto(rival)
	ret.DiffTableHeader = header
	return ret
}
