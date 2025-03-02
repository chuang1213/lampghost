package service

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/Catizard/lampghost_wails/internal/dto"
	"github.com/Catizard/lampghost_wails/internal/entity"
	"github.com/Catizard/lampghost_wails/internal/vo"
	"github.com/charmbracelet/log"
	"gorm.io/gorm"
)

type RivalTagService struct {
	db *gorm.DB
}

func NewRivalTagService(db *gorm.DB, diffTableService *DiffTableService) *RivalTagService {
	return &RivalTagService{
		db: db,
	}
}

func (s *RivalTagService) FindRivalTagList(filter *vo.RivalTagVo) ([]*dto.RivalTagDto, int, error) {
	raw, n, err := findRivalTagList(s.db, filter)
	if err != nil {
		return nil, 0, err
	}
	if n == 0 {
		return make([]*dto.RivalTagDto, 0), 0, err
	}
	ret := make([]*dto.RivalTagDto, len(raw))
	for i := range raw {
		ret[i] = dto.NewRivalTagDto(raw[i])
	}
	return ret, n, nil
}

func (s *RivalTagService) AddRivalTag(rivalTag *vo.RivalTagVo) error {
	if rivalTag == nil {
		return fmt.Errorf("AddRivalTag: rivalTag cannot be nil")
	}
	// if you don't provide the tag name, it would be supplyed with tag time
	if rivalTag.TagName == "" {
		rivalTag.TagName = rivalTag.RecordTime.Format("2006-01-02 15:04:05")
	}
	rivalTag.Generated = false
	return s.db.Create(rivalTag.Entity()).Error
}

func (s *RivalTagService) SyncRivalTagFromRawData(rivalID uint, rawScoreLog []*entity.ScoreLog, rawSongData []*entity.SongData) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := syncRivalTagFromRawData(tx, rivalID, rawScoreLog, rawSongData); err != nil {
		return err
	}
	return tx.Commit().Error
}

func (s *RivalTagService) DeleteRivalTagByID(rivalTagID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		rivalTag, err := findRivalTagByID(s.db, rivalTagID)
		if err != nil {
			return err
		}
		if rivalTag.Generated {
			return fmt.Errorf("DeleteRivalTagByID: cannot delete generated tag")
		}
		return deleteRivalTagByID(tx, rivalTagID)
	})
}

func (s *RivalTagService) RevertRivalTagEnabledState(rivalTagID uint) error {
	return s.db.Model(&entity.RivalTag{}).Where("ID = ?", rivalTagID).UpdateColumn("enabled", gorm.Expr("1 - enabled")).Error
}

/*
Sync one rival's tag list with specified raw scorelog.

For now this function only generate the first clear course tag.
The array of logs' sequence requires nothing, this function would handle it properly,
so there's no need to sort the logs before calling this function.

This function depends on SongDataService's song hash cache and it was due to a historical
problem: scorelog only contains sha256 and difficult table data only contains md5. And the
cache is here to build the relationship. However this also means we cannot call this function
before having the songdata.db loaded properly(which isn't in initialize phase) or have
a corrupted songdata cache. So we have to take a step back to generate the cache in time.
*/
func syncRivalTagFromRawData(tx *gorm.DB, rivalID uint, rawScoreLog []*entity.ScoreLog, rawSongData []*entity.SongData) error {
	courseInfoDtos, _, err := findCourseInfoList(tx, nil)
	// NOTE: we have to build the cache by using rawSongData
	if err != nil {
		return err
	}
	correctCache := generateSongHashCacheFromRawData(rawSongData)
	for _, courseInfoDto := range courseInfoDtos {
		courseInfoDto.RepairHash(correctCache)
	}
	interestHashSet := make(map[string]interface{})
	// NOTE: We cannot use sha256s directly, the hash from logs doesn't split hashes
	for _, courseInfoDto := range courseInfoDtos {
		interestHashSet[courseInfoDto.GetJoinedSha256("")] = new(interface{})
	}
	interestScoreLogs := make([]*entity.ScoreLog, 0)
	for _, scoreLog := range rawScoreLog {
		if _, ok := interestHashSet[scoreLog.Sha256]; ok {
			interestScoreLogs = append(interestScoreLogs, scoreLog)
		}
	}
	if len(interestScoreLogs) == 0 {
		log.Warn("[RivalTagService] There's no course related play record, skip tag build")
		return nil
	}
	sort.Slice(interestScoreLogs, func(i int, j int) bool {
		return interestScoreLogs[i].TimeStamp < interestScoreLogs[j].TimeStamp
	})

	tags := make([]entity.RivalTag, 0)
	for _, courseInfoDto := range courseInfoDtos {
		for _, scoreLog := range interestScoreLogs {
			if scoreLog.Clear < entity.Normal || scoreLog.Sha256 != courseInfoDto.GetJoinedSha256("") {
				continue
			}
			// see courseInfo.go for details
			scoreLogMode, err := strconv.Atoi(scoreLog.Mode)
			if err != nil {
				// do nothing
				continue
			}
			if scoreLogMode/100 != courseInfoDto.GetConstraintMode()/100 {
				continue
			}
			fct := entity.RivalTag{
				RivalId:    rivalID,
				TagName:    courseInfoDto.Name + " First Clear",
				Generated:  true,
				RecordTime: time.Unix(scoreLog.TimeStamp, 0).Local(),
			}
			tags = append(tags, fct)
			break
		}
	}

	if err := tx.Unscoped().Where("rival_id = ?", rivalID).Delete(&entity.RivalTag{}).Error; err != nil {
		return err
	}

	return tx.Create(tags).Error
}

func findRivalTagList(tx *gorm.DB, filter *vo.RivalTagVo) ([]*entity.RivalTag, int, error) {
	partial := tx.Model(&entity.RivalTag{})
	if filter != nil {
		partial = partial.Where(filter.Entity())
	}
	var out []*entity.RivalTag
	if err := partial.Scopes(
		pagination(filter.Pagination),
	).Find(&out).Error; err != nil {
		return nil, 0, err
	}

	if filter != nil && filter.Pagination != nil {
		if filter.Pagination != nil {
			count, err := selectRivalTagCount(tx, filter)
			if err != nil {
				return nil, 0, err
			}
			filter.Pagination.PageCount = calcPageCount(count, filter.Pagination.PageSize)
		}
	}
	return out, len(out), nil
}

func findRivalTagByID(tx *gorm.DB, ID uint) (*entity.RivalTag, error) {
	tag := entity.RivalTag{}
	if err := tx.First(&tag, ID).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

func selectRivalTagCount(tx *gorm.DB, filter *vo.RivalTagVo) (int64, error) {
	querying := tx.Model(&entity.RivalTag{})
	if filter != nil {
		querying = querying.Where(filter.Entity())
	}
	var count int64
	if err := querying.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func deleteRivalTagByID(tx *gorm.DB, rivalTagID uint) error {
	return tx.Unscoped().Where("ID = ?", rivalTagID).Delete(&entity.RivalTag{}).Error
}
