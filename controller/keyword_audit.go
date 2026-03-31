package controller

import (
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

func GetKeywordAudits(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	filters := make(map[string]interface{})

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			filters["user_id"] = uint(userID)
		}
	}

	if username := c.Query("username"); username != "" {
		filters["username"] = username
	}

	if keyword := c.Query("keyword"); keyword != "" {
		filters["keyword"] = keyword
	}

	if action := c.Query("action"); action != "" {
		filters["action"] = action
	}

	if processedStr := c.Query("processed"); processedStr != "" {
		processed := processedStr == "true"
		filters["processed"] = processed
	}

	if startDate := c.Query("start_date"); startDate != "" {
		filters["start_date"] = startDate
	}

	if endDate := c.Query("end_date"); endDate != "" {
		filters["end_date"] = endDate
	}

	audits, total, err := model.GetAllKeywordAudits(page, pageSize, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取审计记录失败",
		})
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"items":      audits,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
			"total_page": totalPages,
		},
	})
}

func GetKeywordAudit(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的ID",
		})
		return
	}

	audit, err := model.GetKeywordAuditByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "审计记录不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    audit,
	})
}

type UpdateKeywordAuditRequest struct {
	Note string `json:"note"`
}

func UpdateKeywordAudit(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的ID",
		})
		return
	}

	var req UpdateKeywordAuditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的请求参数",
		})
		return
	}

	audit, err := model.GetKeywordAuditByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "审计记录不存在",
		})
		return
	}

	audit.Note = req.Note

	if err := model.UpdateKeywordAudit(audit); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "更新审计记录失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "审计记录更新成功",
		"data":    audit,
	})
}

func MarkAuditProcessed(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的ID",
		})
		return
	}

	var req UpdateKeywordAuditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Note = ""
	}

	userInfo, _ := c.Get("user")

	var username string
	if userInfo != nil {
		if u, ok := userInfo.(model.User); ok {
			username = u.Username
		}
	}

	if err := model.MarkAuditAsProcessed(uint(id), username, req.Note); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "标记处理失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "审计记录已标记为已处理",
	})
}

func DeleteKeywordAudit(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的ID",
		})
		return
	}

	if err := model.DeleteKeywordAudit(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "删除审计记录失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "审计记录删除成功",
	})
}

func GetKeywordAuditStats(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "7")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 {
		days = 7
	}

	stats, err := model.GetKeywordAuditStats(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取统计信息失败",
		})
		return
	}

	unprocessedCount, _ := model.GetUnprocessedKeywordAuditsCount()
	stats["unprocessed"] = unprocessedCount

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

func DeleteOldKeywordAudits(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的天数",
		})
		return
	}

	rowsAffected, err := model.DeleteOldKeywordAudits(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "删除旧记录失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "成功删除 " + strconv.FormatInt(rowsAffected, 10) + " 条旧记录",
	})
}
