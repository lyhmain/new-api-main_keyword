package setting

import (
	"strings"
)

var CheckSensitiveEnabled = true
var CheckSensitiveOnPromptEnabled = true

var CheckSensitiveOnCompletionEnabled = true

var StopOnSensitiveEnabled = false

var StreamCacheQueueLength = 0

var SensitiveWords = []string{
	"test_sensitive",
}

var ReplacementMap = make(map[string]string)

var CustomReplacementEnabled = false

func SensitiveWordsToString() string {
	return strings.Join(SensitiveWords, "\n")
}

func SensitiveWordsFromString(s string) {
	SensitiveWords = []string{}
	sw := strings.Split(s, "\n")
	for _, w := range sw {
		w = strings.TrimSpace(w)
		if w != "" {
			SensitiveWords = append(SensitiveWords, w)
		}
	}
}

func ShouldCheckPromptSensitive() bool {
	return CheckSensitiveEnabled && CheckSensitiveOnPromptEnabled
}

func ShouldCheckCompletionSensitive() bool {
	return CheckSensitiveEnabled && CheckSensitiveOnCompletionEnabled
}

func GetReplacementMap() map[string]string {
	return ReplacementMap
}

func SetReplacementMap(m map[string]string) {
	ReplacementMap = m
}

func ClearReplacementMap() {
	ReplacementMap = make(map[string]string)
}

func IsCustomReplacementEnabled() bool {
	return CustomReplacementEnabled
}

func SetCustomReplacementEnabled(enabled bool) {
	CustomReplacementEnabled = enabled
}

// Keyword replacement settings
var KeywordReplacementEnabled = true
var KeywordAuditEnabled = true

// ShouldApplyKeywordReplacement returns whether keyword replacement should be applied
func ShouldApplyKeywordReplacement() bool {
	return KeywordReplacementEnabled
}

// ShouldApplyKeywordAudit returns whether keyword audit should be applied
func ShouldApplyKeywordAudit() bool {
	return KeywordAuditEnabled
}

// ResponseKeywordReplacementEnabled - whether to replace keywords in AI responses
var ResponseKeywordReplacementEnabled = true

// ShouldApplyResponseKeywordReplacement returns whether response keyword replacement should be applied
func ShouldApplyResponseKeywordReplacement() bool {
	return ResponseKeywordReplacementEnabled
}
