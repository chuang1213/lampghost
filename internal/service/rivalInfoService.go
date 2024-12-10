package service

import (
	"fmt"

	"github.com/Catizard/lampghost_wails/internal/entity"
	"github.com/charmbracelet/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type RivalInfoService struct {
	db *gorm.DB
}

func NewRivalInfoService(db *gorm.DB) *RivalInfoService {
	return &RivalInfoService{
		db: db,
	}
}

func (s *RivalInfoService) FindRivalInfoList() ([]*entity.RivalInfo, int, error) {
	var rivals []*entity.RivalInfo
	if err := s.db.Find(&rivals).Error; err != nil {
		return nil, 0, err
	}
	return rivals, len(rivals), nil
}

func (s *RivalInfoService) FindRivalInfoByID(rivalID uint) (*entity.RivalInfo, error) {
	out := entity.RivalInfo{}
	if err := s.db.First(&out, rivalID).Error; err != nil {
		log.Debugf("[RivalInfoService] FindRivalInfoByID with ID=%d failed: %v\n", rivalID, err)
		return nil, err
	}
	return &out, nil
}

func (s *RivalInfoService) SyncRivalScoreLog(rivalInfo *entity.RivalInfo) error {
	log.Debug("[Service] calling RivalInfoService.SyncRivalScoreLog")
	if rivalInfo.ScoreLogPath == nil || *rivalInfo.ScoreLogPath == "" {
		return fmt.Errorf("Cannot sync rival %s's score log: score log file path is empty!", rivalInfo.Name)
	}
	dsn := *rivalInfo.ScoreLogPath
	log.Debugf("[RivalInfoService] Trying to read from %s", dsn)
	scoreLogDB, err := gorm.Open(sqlite.Open(dsn))
	if err != nil {
		return err
	}
	scoreLogService := NewScoreLogService(scoreLogDB)
	rawScoreLog, n, err := scoreLogService.FindScoreLogList()
	if err != nil {
		return err
	}
	log.Debugf("[RivalInfoService] Read %d logs from %s", n, *rivalInfo.ScoreLogPath)
	if n == 0 {
		return nil
	}

	rivalScoreLog := make([]entity.RivalScoreLog, len(rawScoreLog))
	for i, rawLog := range rawScoreLog {
		rivalLog := entity.FromRawScoreLogToRivalScoreLog(rawLog)
		rivalLog.RivalId = rivalInfo.ID
		rivalScoreLog[i] = rivalLog
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("rival_id = ?", rivalInfo.ID).Delete(&entity.RivalScoreLog{}).Error; err != nil {
			return err
		}

		if err := tx.Create(&rivalScoreLog).Error; err != nil {
			return err
		}
		if err := tx.Model(rivalInfo).Updates(entity.RivalInfo{
			PlayCount: n,
		}).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	log.Debugf("[RivalInfoService] Sync rival %s's %d scorelogs successfully!", rivalInfo.Name, n)
	return nil
}

func (s *RivalInfoService) SyncRivalScoreLogByID(rivalID uint) error {
	if rivalInfo, err := s.FindRivalInfoByID(rivalID); err != nil {
		return err
	} else {
		return s.SyncRivalScoreLog(rivalInfo)
	}
}
