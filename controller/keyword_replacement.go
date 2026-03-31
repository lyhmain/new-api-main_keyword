package controller

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

func GetKeywordReplacements(c *gin.Context) {
	replacements, err := model.GetAllKeywordReplacements()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取关键词替换列表失败",
		})
		return
	}

	stats, _ := model.GetKeywordReplacementStats()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"items": replacements,
			"stats": stats,
		},
	})
}

func GetKeywordReplacement(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的ID",
		})
		return
	}

	replacement, err := model.GetKeywordReplacementByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "关键词替换不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    replacement,
	})
}

type CreateKeywordReplacementRequest struct {
	Keyword        string `json:"keyword" binding:"required"`
	Replacement    string `json:"replacement" binding:"required"`
	Enabled        *bool  `json:"enabled"`
	IsRegex        *bool  `json:"is_regex"`
	CaseSensitive  *bool  `json:"case_sensitive"`
	Priority       *int   `json:"priority"`
	Description    string `json:"description"`
	AuditThreshold *int   `json:"audit_threshold"`
}

func CreateKeywordReplacement(c *gin.Context) {
	var req CreateKeywordReplacementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的请求参数: " + err.Error(),
		})
		return
	}

	if req.Keyword == "" || req.Replacement == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "关键词和替换词不能为空",
		})
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	isRegex := false
	if req.IsRegex != nil {
		isRegex = *req.IsRegex
	}

	caseSensitive := false
	if req.CaseSensitive != nil {
		caseSensitive = *req.CaseSensitive
	}

	priority := 0
	if req.Priority != nil {
		priority = *req.Priority
	}

	auditThreshold := 0
	if req.AuditThreshold != nil {
		auditThreshold = *req.AuditThreshold
	}

	replacement := &model.KeywordReplacement{
		Keyword:        req.Keyword,
		Replacement:    req.Replacement,
		Enabled:        enabled,
		IsRegex:        isRegex,
		CaseSensitive:  caseSensitive,
		Priority:       priority,
		Description:    req.Description,
		AuditThreshold: auditThreshold,
	}

	if err := model.CreateKeywordReplacement(replacement); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "创建关键词替换失败: " + err.Error(),
		})
		return
	}

	model.ClearKeywordReplacementsCache()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "关键词替换创建成功",
		"data":    replacement,
	})
}

type UpdateKeywordReplacementRequest struct {
	Keyword        string `json:"keyword"`
	Replacement    string `json:"replacement"`
	Enabled        *bool  `json:"enabled"`
	IsRegex        *bool  `json:"is_regex"`
	CaseSensitive  *bool  `json:"case_sensitive"`
	Priority       *int   `json:"priority"`
	Description    string `json:"description"`
	AuditThreshold *int   `json:"audit_threshold"`
}

func UpdateKeywordReplacement(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的ID",
		})
		return
	}

	var req UpdateKeywordReplacementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的请求参数",
		})
		return
	}

	replacement, err := model.GetKeywordReplacementByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "关键词替换不存在",
		})
		return
	}

	if req.Keyword != "" {
		replacement.Keyword = req.Keyword
	}
	if req.Replacement != "" {
		replacement.Replacement = req.Replacement
	}
	if req.Enabled != nil {
		replacement.Enabled = *req.Enabled
	}
	if req.IsRegex != nil {
		replacement.IsRegex = *req.IsRegex
	}
	if req.CaseSensitive != nil {
		replacement.CaseSensitive = *req.CaseSensitive
	}
	if req.Priority != nil {
		replacement.Priority = *req.Priority
	}
	if req.Description != "" {
		replacement.Description = req.Description
	}
	if req.AuditThreshold != nil {
		replacement.AuditThreshold = *req.AuditThreshold
	}

	if err := model.UpdateKeywordReplacement(replacement); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "更新关键词替换失败",
		})
		return
	}

	model.ClearKeywordReplacementsCache()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "关键词替换更新成功",
		"data":    replacement,
	})
}

func DeleteKeywordReplacement(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的ID",
		})
		return
	}

	if err := model.DeleteKeywordReplacement(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "删除关键词替换失败",
		})
		return
	}

	model.ClearKeywordReplacementsCache()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "关键词替换删除成功",
	})
}

func ImportKeywordReplacements(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请上传文件",
		})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "无法打开文件",
		})
		return
	}
	defer src.Close()

	reader := csv.NewReader(src)

	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "解析CSV文件失败",
		})
		return
	}

	if len(records) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "CSV文件格式错误，至少需要包含标题行和数据行",
		})
		return
	}

	var replacements []model.KeywordReplacement

	for _, record := range records[1:] {
		if len(record) < 2 {
			continue
		}

		keyword := strings.TrimSpace(record[0])
		replacement := strings.TrimSpace(record[1])

		if keyword == "" || replacement == "" {
			continue
		}

		enabled := true
		isRegex := false
		caseSensitive := false
		priority := 0

		if len(record) > 2 && strings.TrimSpace(record[2]) != "" {
			enabled = strings.ToLower(strings.TrimSpace(record[2])) == "true" || strings.TrimSpace(record[2]) == "1"
		}
		if len(record) > 3 && strings.TrimSpace(record[3]) != "" {
			isRegex = strings.ToLower(strings.TrimSpace(record[3])) == "true" || strings.TrimSpace(record[3]) == "1"
		}
		if len(record) > 4 && strings.TrimSpace(record[4]) != "" {
			caseSensitive = strings.ToLower(strings.TrimSpace(record[4])) == "true" || strings.TrimSpace(record[4]) == "1"
		}
		if len(record) > 5 && strings.TrimSpace(record[5]) != "" {
			p, _ := strconv.Atoi(strings.TrimSpace(record[5]))
			priority = p
		}

		description := ""
		if len(record) > 6 {
			description = strings.TrimSpace(record[6])
		}

		replacements = append(replacements, model.KeywordReplacement{
			Keyword:       keyword,
			Replacement:   replacement,
			Enabled:       enabled,
			IsRegex:       isRegex,
			CaseSensitive: caseSensitive,
			Priority:      priority,
			Description:   description,
		})
	}

	if len(replacements) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "没有有效的数据可导入",
		})
		return
	}

	if err := model.ImportKeywordReplacements(replacements); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "导入失败: " + err.Error(),
		})
		return
	}

	model.ClearKeywordReplacementsCache()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("成功导入 %d 条记录", len(replacements)),
	})
}

func ExportKeywordReplacements(c *gin.Context) {
	replacements, err := model.ExportKeywordReplacements()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "导出失败",
		})
		return
	}

	records := [][]string{
		{"关键词", "替换词", "启用", "正则", "区分大小写", "优先级", "描述"},
	}

	for _, r := range replacements {
		records = append(records, []string{
			r.Keyword,
			r.Replacement,
			strconv.FormatBool(r.Enabled),
			strconv.FormatBool(r.IsRegex),
			strconv.FormatBool(r.CaseSensitive),
			strconv.Itoa(r.Priority),
			r.Description,
		})
	}

	filename := fmt.Sprintf("keyword_replacements_%s.csv", common.GetTimestamp())
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename="+filename)

	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(c.Writer)
	writer.WriteAll(records)
}

func GetKeywordReplacementStats(c *gin.Context) {
	stats, err := model.GetKeywordReplacementStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取统计信息失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

var _ = middleware.AdminAuth
