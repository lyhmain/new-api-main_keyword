package model

import (
	"sync"
	"time"

	"gorm.io/gorm"
)

type KeywordReplacement struct {
	ID             uint      `json:"id" gorm:"primarykey"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Keyword        string    `json:"keyword" gorm:"uniqueIndex;size:500;not null"`
	Replacement    string    `json:"replacement" gorm:"size:500;not null"`
	Enabled        bool      `json:"enabled" gorm:"default:true"`
	IsRegex        bool      `json:"is_regex" gorm:"default:false"`
	CaseSensitive  bool      `json:"case_sensitive" gorm:"default:false"`
	Priority       int       `json:"priority" gorm:"default:0"`
	Description    string    `json:"description" gorm:"size:1000"`
	AuditThreshold int       `json:"audit_threshold" gorm:"default:0"`
}

func (KeywordReplacement) TableName() string {
	return "keyword_replacements"
}

func GetAllKeywordReplacements() ([]KeywordReplacement, error) {
	var replacements []KeywordReplacement
	err := DB.Order("priority DESC, created_at DESC").Find(&replacements).Error
	return replacements, err
}

func GetEnabledKeywordReplacements() ([]KeywordReplacement, error) {
	var replacements []KeywordReplacement
	err := DB.Where("enabled = ?", true).Order("priority DESC").Find(&replacements).Error
	return replacements, err
}

func GetKeywordReplacementByID(id uint) (*KeywordReplacement, error) {
	var replacement KeywordReplacement
	err := DB.First(&replacement, id).Error
	if err != nil {
		return nil, err
	}
	return &replacement, nil
}

func CreateKeywordReplacement(replacement *KeywordReplacement) error {
	return DB.Create(replacement).Error
}

func UpdateKeywordReplacement(replacement *KeywordReplacement) error {
	return DB.Save(replacement).Error
}

func DeleteKeywordReplacement(id uint) error {
	return DB.Delete(&KeywordReplacement{}, id).Error
}

func ImportKeywordReplacements(replacements []KeywordReplacement) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		for i := range replacements {
			replacement := &replacements[i]
			var existing KeywordReplacement
			err := tx.Where("keyword = ?", replacement.Keyword).First(&existing).Error
			if err == gorm.ErrRecordNotFound {
				if err := tx.Create(replacement).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else {
				existing.Replacement = replacement.Replacement
				existing.Enabled = replacement.Enabled
				existing.IsRegex = replacement.IsRegex
				existing.CaseSensitive = replacement.CaseSensitive
				existing.Priority = replacement.Priority
				existing.Description = replacement.Description
				existing.AuditThreshold = replacement.AuditThreshold
				if err := tx.Save(&existing).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func ExportKeywordReplacements() ([]KeywordReplacement, error) {
	var replacements []KeywordReplacement
	err := DB.Order("priority DESC, created_at DESC").Find(&replacements).Error
	return replacements, err
}

func GetKeywordReplacementStats() (map[string]interface{}, error) {
	var total int64
	var enabled int64
	var regexCount int64
	var withAudit int64

	DB.Model(&KeywordReplacement{}).Count(&total)
	DB.Model(&KeywordReplacement{}).Where("enabled = ?", true).Count(&enabled)
	DB.Model(&KeywordReplacement{}).Where("is_regex = ?", true).Count(&regexCount)
	DB.Model(&KeywordReplacement{}).Where("audit_threshold > ?", 0).Count(&withAudit)

	return map[string]interface{}{
		"total":       total,
		"enabled":     enabled,
		"regex_count": regexCount,
		"with_audit":  withAudit,
		"disabled":    total - enabled,
	}, nil
}

var keywordReplacementsCache []KeywordReplacement
var keywordReplacementsCacheTime time.Time
var keywordReplacementsCacheMutex sync.RWMutex

func GetKeywordReplacementsCache() ([]KeywordReplacement, error) {
	keywordReplacementsCacheMutex.RLock()
	defer keywordReplacementsCacheMutex.RUnlock()

	if len(keywordReplacementsCache) > 0 && time.Since(keywordReplacementsCacheTime) < 5*time.Second {
		return keywordReplacementsCache, nil
	}

	keywordReplacementsCacheMutex.RUnlock()
	keywordReplacementsCacheMutex.Lock()
	defer keywordReplacementsCacheMutex.Unlock()

	if len(keywordReplacementsCache) > 0 && time.Since(keywordReplacementsCacheTime) < 5*time.Second {
		return keywordReplacementsCache, nil
	}

	replacements, err := GetEnabledKeywordReplacements()
	if err != nil {
		return nil, err
	}

	keywordReplacementsCache = replacements
	keywordReplacementsCacheTime = time.Now()

	return replacements, nil
}

func ClearKeywordReplacementsCache() {
	keywordReplacementsCacheMutex.Lock()
	defer keywordReplacementsCacheMutex.Unlock()
	keywordReplacementsCache = nil
	keywordReplacementsCacheTime = time.Time{}
}
