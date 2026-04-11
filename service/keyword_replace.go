package service

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
)

type KeywordMatch struct {
	Keyword     string
	Replacement string
	StartPos    int
	EndPos      int
	IsRegex     bool
}

type ReplaceResult struct {
	Text            string
	Replaced        bool
	MatchCount      int
	TriggeredAudit  bool
	Blocked         bool
	MatchedKeywords []string
}

func KeywordReplace(text string, enableAudit bool) ReplaceResult {
	result := ReplaceResult{
		Text:           text,
		Replaced:       false,
		MatchCount:     0,
		TriggeredAudit: false,
		Blocked:        false,
	}

	if len(text) == 0 {
		return result
	}

	replacements, err := model.GetCompiledKeywordReplacementsCache()
	if err != nil || len(replacements) == 0 {
		return result
	}

	var matches []KeywordMatch
	lowerText := strings.ToLower(text)

	for _, kr := range replacements {
		if !kr.Enabled {
			continue
		}

		keyword := kr.Keyword
		replacement := kr.Replacement

		if kr.IsRegex && kr.Regex != nil {
			// Use pre-compiled regex from cache
			loc := kr.Regex.FindStringIndex(text)
			if loc != nil {
				matches = append(matches, KeywordMatch{
					Keyword:     keyword,
					Replacement: replacement,
					StartPos:    loc[0],
					EndPos:      loc[1],
					IsRegex:     true,
				})
			}
		} else {
			if kr.CaseSensitive {
				idx := strings.Index(text, keyword)
				if idx >= 0 {
					matches = append(matches, KeywordMatch{
						Keyword:     keyword,
						Replacement: replacement,
						StartPos:    idx,
						EndPos:      idx + utf8.RuneCountInString(keyword),
						IsRegex:     false,
					})
				}
			} else {
				idx := strings.Index(lowerText, strings.ToLower(keyword))
				if idx >= 0 {
					matches = append(matches, KeywordMatch{
						Keyword:     keyword,
						Replacement: replacement,
						StartPos:    idx,
						EndPos:      idx + utf8.RuneCountInString(keyword),
						IsRegex:     false,
					})
				}
			}
		}
	}

	if len(matches) == 0 {
		return result
	}

	result.MatchCount = len(matches)
	result.Replaced = true
	result.MatchedKeywords = make([]string, 0, len(matches))

	for _, match := range matches {
		result.MatchedKeywords = append(result.MatchedKeywords, match.Keyword)
	}

	if enableAudit && setting.StopOnSensitiveEnabled {
		result.Blocked = true
		result.TriggeredAudit = true
		return result
	}

	result.Text = applyReplacements(text, matches)
	result.TriggeredAudit = enableAudit

	return result
}

func KeywordReplaceWithMap(text string, customMap map[string]string) string {
	if len(text) == 0 || len(customMap) == 0 {
		return text
	}

	result := text

	for keyword, replacement := range customMap {
		if keyword == "" {
			continue
		}

		result = strings.ReplaceAll(result, keyword, replacement)
	}

	return result
}

func ChineseKeywordReplace(text string) string {
	if len(text) == 0 {
		return text
	}

	replacements, err := model.GetCompiledKeywordReplacementsCache()
	if err != nil || len(replacements) == 0 {
		return text
	}

	result := text

	for _, kr := range replacements {
		if !kr.Enabled {
			continue
		}

		keyword := kr.Keyword
		replacement := kr.Replacement

		if kr.IsRegex {
			var pattern string
			if kr.CaseSensitive {
				pattern = keyword
			} else {
				pattern = "(?i)" + keyword
			}

			re, err := regexp.Compile(pattern)
			if err != nil {
				continue
			}

			result = re.ReplaceAllString(result, replacement)
		} else {
			if kr.CaseSensitive {
				result = strings.ReplaceAll(result, keyword, replacement)
			} else {
				result = strings.ReplaceAll(strings.ToLower(result), strings.ToLower(keyword), replacement)
			}
		}
	}

	return result
}

func CheckKeywordAudit(text string, userID uint, username string, channelID uint, channelName string, modelName string, requestType string, ipAddress string, userAgent string, tokenID uint, tokenName string) {
	if len(text) == 0 {
		return
	}

	replacements, err := model.GetCompiledKeywordReplacementsCache()
	if err != nil || len(replacements) == 0 {
		return
	}

	lowerText := strings.ToLower(text)
	var matchedKeywords []string

	for _, kr := range replacements {
		if !kr.Enabled || kr.AuditThreshold == 0 {
			continue
		}

		keyword := kr.Keyword

		if kr.IsRegex {
			var pattern string
			if kr.CaseSensitive {
				pattern = keyword
			} else {
				pattern = "(?i)" + keyword
			}

			re, err := regexp.Compile(pattern)
			if err != nil {
				continue
			}

			if re.MatchString(text) {
				matchedKeywords = append(matchedKeywords, kr.Keyword)
			}
		} else {
			checkText := text
			if !kr.CaseSensitive {
				checkText = lowerText
				keyword = strings.ToLower(keyword)
			}

			if strings.Contains(checkText, keyword) {
				matchedKeywords = append(matchedKeywords, kr.Keyword)
			}
		}
	}

	if len(matchedKeywords) > 0 {
		for _, keyword := range matchedKeywords {
			audit := &model.KeywordAudit{
				Keyword:     keyword,
				Context:     text,
				Action:      "audit",
				UserID:      userID,
				Username:    username,
				ChannelID:   channelID,
				ChannelName: channelName,
				Model:       modelName,
				RequestType: requestType,
				IPAddress:   ipAddress,
				UserAgent:   userAgent,
				TokenID:     tokenID,
				TokenName:   tokenName,
			}
			model.CreateKeywordAudit(audit)
		}
	}
}

func applyReplacements(text string, matches []KeywordMatch) string {
	if len(matches) == 0 {
		return text
	}

	var builder strings.Builder
	builder.Grow(len(text))

	lastPos := 0
	for _, match := range matches {
		if match.StartPos < lastPos {
			continue
		}

		builder.WriteString(text[lastPos:match.StartPos])
		builder.WriteString(match.Replacement)
		lastPos = match.EndPos
	}

	builder.WriteString(text[lastPos:])

	return builder.String()
}

func GetCustomReplacementMap() map[string]string {
	replacements, err := model.GetCompiledKeywordReplacementsCache()
	if err != nil || len(replacements) == 0 {
		return nil
	}

	customMap := make(map[string]string)
	for _, kr := range replacements {
		if kr.Enabled {
			customMap[kr.Keyword] = kr.Replacement
		}
	}

	if len(customMap) == 0 {
		return nil
	}

	return customMap
}

// ReplaceKeywordsInResponse applies keyword replacement to response content
// Uses pre-compiled regex for better performance
func ReplaceKeywordsInResponse(responseBody []byte) ([]byte, bool) {
	if !setting.ShouldApplyResponseKeywordReplacement() {
		return responseBody, false
	}

	compiledReplacements, err := model.GetCompiledKeywordReplacementsCache()
	if err != nil || len(compiledReplacements) == 0 {
		return responseBody, false
	}

	modified := false
	result := string(responseBody)
	lowerResult := strings.ToLower(result)

	for _, kr := range compiledReplacements {
		if !kr.Enabled {
			continue
		}

		keyword := kr.Keyword
		replacement := kr.Replacement

		if kr.IsRegex && kr.Regex != nil {
			if kr.Regex.MatchString(result) {
				result = kr.Regex.ReplaceAllString(result, replacement)
				modified = true
				lowerResult = strings.ToLower(result)
			}
		} else if !kr.IsRegex {
			if kr.CaseSensitive {
				if strings.Contains(result, keyword) {
					result = strings.ReplaceAll(result, keyword, replacement)
					modified = true
				}
			} else {
				lowerKeyword := strings.ToLower(keyword)
				if strings.Contains(lowerResult, lowerKeyword) {
					var builder strings.Builder
					lastPos := 0
					for {
						idx := strings.Index(lowerResult[lastPos:], lowerKeyword)
						if idx == -1 {
							builder.WriteString(result[lastPos:])
							break
						}
						actualPos := lastPos + idx
						builder.WriteString(result[lastPos:actualPos])
						builder.WriteString(replacement)
						lastPos = actualPos + len(keyword)
						modified = true
					}
					result = builder.String()
					lowerResult = strings.ToLower(result)
				}
			}
		}
	}

	if modified {
		return []byte(result), true
	}

	return responseBody, false
}
