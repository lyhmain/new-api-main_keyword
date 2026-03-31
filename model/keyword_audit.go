package model

import (
	"time"
)

type KeywordAudit struct {
	ID           uint       `json:"id" gorm:"primarykey"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	Keyword      string     `json:"keyword" gorm:"size:500;not null;index"`
	Context      string     `json:"context" gorm:"type:text"`
	MatchedRule  string     `json:"matched_rule" gorm:"size:500"`
	Action       string     `json:"action" gorm:"size:20;not null"`
	ActionDetail string     `json:"action_detail" gorm:"size:1000"`
	UserID       uint       `json:"user_id" gorm:"index"`
	Username     string     `json:"username" gorm:"size:100"`
	ChannelID    uint       `json:"channel_id" gorm:"index"`
	ChannelName  string     `json:"channel_name" gorm:"size:200"`
	Model        string     `json:"model" gorm:"size:200;index"`
	RequestType  string     `json:"request_type" gorm:"size:50"`
	IPAddress    string     `json:"ip_address" gorm:"size:50"`
	UserAgent    string     `json:"user_agent" gorm:"size:500"`
	TokenID      uint       `json:"token_id" gorm:"index"`
	TokenName    string     `json:"token_name" gorm:"size:200"`
	Processed    bool       `json:"processed" gorm:"default:false"`
	ProcessedBy  string     `json:"processed_by" gorm:"size:100"`
	ProcessedAt  *time.Time `json:"processed_at"`
	Note         string     `json:"note" gorm:"size:1000"`
}

func (KeywordAudit) TableName() string {
	return "keyword_audits"
}

func GetAllKeywordAudits(page int, pageSize int, filters map[string]interface{}) ([]KeywordAudit, int64, error) {
	var audits []KeywordAudit
	var total int64

	query := DB.Model(&KeywordAudit{})

	if userID, ok := filters["user_id"].(uint); ok && userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if username, ok := filters["username"].(string); ok && username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if keyword, ok := filters["keyword"].(string); ok && keyword != "" {
		query = query.Where("keyword LIKE ?", "%"+keyword+"%")
	}
	if action, ok := filters["action"].(string); ok && action != "" {
		query = query.Where("action = ?", action)
	}
	if processed, ok := filters["processed"].(bool); ok {
		query = query.Where("processed = ?", processed)
	}
	if startDate, ok := filters["start_date"].(string); ok && startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate, ok := filters["end_date"].(string); ok && endDate != "" {
		query = query.Where("created_at <= ?", endDate)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if page > 0 {
		offset := (page - 1) * pageSize
		err = query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&audits).Error
	} else {
		err = query.Order("created_at DESC").Find(&audits).Error
	}

	return audits, total, err
}

func GetKeywordAuditByID(id uint) (*KeywordAudit, error) {
	var audit KeywordAudit
	err := DB.First(&audit, id).Error
	if err != nil {
		return nil, err
	}
	return &audit, nil
}

func CreateKeywordAudit(audit *KeywordAudit) error {
	return DB.Create(audit).Error
}

func BatchCreateKeywordAudits(audits []KeywordAudit) error {
	if len(audits) == 0 {
		return nil
	}
	return DB.Create(&audits).Error
}

func UpdateKeywordAudit(audit *KeywordAudit) error {
	return DB.Save(audit).Error
}

func MarkAuditAsProcessed(id uint, processedBy string, note string) error {
	now := time.Now()
	return DB.Model(&KeywordAudit{}).Where("id = ?", id).Updates(map[string]interface{}{
		"processed":    true,
		"processed_by": processedBy,
		"processed_at": now,
		"note":         note,
	}).Error
}

func DeleteKeywordAudit(id uint) error {
	return DB.Delete(&KeywordAudit{}, id).Error
}

func DeleteOldKeywordAudits(days int) (int64, error) {
	cutoffDate := time.Now().AddDate(0, 0, -days)
	result := DB.Where("created_at < ? AND processed = ?", cutoffDate, true).Delete(&KeywordAudit{})
	return result.RowsAffected, result.Error
}

func GetKeywordAuditStats(days int) (map[string]interface{}, error) {
	var stats = make(map[string]interface{})

	now := time.Now()
	startDate := now.AddDate(0, 0, -days)

	var total int64
	var replaced int64
	var blocked int64
	var pending int64

	DB.Model(&KeywordAudit{}).Where("created_at >= ?", startDate).Count(&total)
	DB.Model(&KeywordAudit{}).Where("created_at >= ? AND action = ?", startDate, "replace").Count(&replaced)
	DB.Model(&KeywordAudit{}).Where("created_at >= ? AND action = ?", startDate, "block").Count(&blocked)
	DB.Model(&KeywordAudit{}).Where("created_at >= ? AND processed = ?", startDate, false).Count(&pending)

	stats["total"] = total
	stats["replaced"] = replaced
	stats["blocked"] = blocked
	stats["pending"] = pending

	var topKeywords []struct {
		Keyword string
		Count   int64
	}
	DB.Model(&KeywordAudit{}).Select("keyword, COUNT(*) as count").
		Where("created_at >= ?", startDate).
		Group("keyword").
		Order("count DESC").
		Limit(10).
		Scan(&topKeywords)

	stats["top_keywords"] = topKeywords

	var topUsers []struct {
		Username string
		Count    int64
	}
	DB.Model(&KeywordAudit{}).Select("username, COUNT(*) as count").
		Where("created_at >= ? AND username != ''", startDate).
		Group("username").
		Order("count DESC").
		Limit(10).
		Scan(&topUsers)

	stats["top_users"] = topUsers

	return stats, nil
}

func GetUnprocessedKeywordAuditsCount() (int64, error) {
	var count int64
	err := DB.Model(&KeywordAudit{}).Where("processed = ?", false).Count(&count).Error
	return count, err
}
