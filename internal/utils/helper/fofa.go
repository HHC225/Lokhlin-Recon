// helper-------------------------------------
// @file      : fofa.go
// @time      : 2025/12/21 19:39
// -------------------------------------------

package helper

import (
	"fmt"
	"regexp"
	"strings"
)

func ConvertFofaToCustomQuery(fofaQueryValue interface{}) string {
	var queries []string

	if arr, ok := fofaQueryValue.([]interface{}); ok {
		for _, item := range arr {
			if str, ok := item.(string); ok {
				converted := parseFofaQuery(str)
				if converted != "" {
					queries = append(queries, converted)
				}
			}
		}

		if len(queries) > 0 {
			return strings.Join(queries, " && ")
		}
		return ""
	}

	if str, ok := fofaQueryValue.(string); ok {
		return parseFofaQuery(str)
	}

	return ""
}

func parseFofaQuery(query string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return ""
	}

	return parseFofaQueryWithBrackets(query)
}

func parseFofaQueryWithBrackets(query string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return ""
	}

	for {

		start := strings.Index(query, "(")
		if start == -1 {

			break
		}

		depth := 0
		end := -1
		for i := start; i < len(query); i++ {
			if query[i] == '(' {
				depth++
			} else if query[i] == ')' {
				depth--
				if depth == 0 {
					end = i
					break
				}
			}
		}

		if end == -1 {

			break
		}

		innerQuery := query[start+1 : end]
		parsedInner := parseFofaQueryWithBrackets(innerQuery)

		query = query[:start] + "(" + parsedInner + ")" + query[end+1:]
	}

	return parseFofaLogicExpression(query)
}

func parseFofaLogicExpression(query string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return ""
	}

	if strings.Contains(query, " || ") {
		parts := splitByOperator(query, " || ")
		var results []string
		for _, part := range parts {
			parsed := parseFofaLogicExpression(strings.TrimSpace(part))
			if parsed != "" {
				results = append(results, parsed)
			}
		}
		if len(results) > 0 {
			return strings.Join(results, " || ")
		}
		return ""
	}

	if strings.Contains(query, " && ") {
		parts := splitByOperator(query, " && ")
		var results []string
		for _, part := range parts {
			parsed := parseFofaLogicExpression(strings.TrimSpace(part))
			if parsed != "" {
				results = append(results, parsed)
			}
		}
		if len(results) > 0 {
			return strings.Join(results, " && ")
		}
		return ""
	}

	return parseFofaCondition(query)
}

func splitByOperator(query, operator string) []string {
	var parts []string
	var current strings.Builder
	depth := 0

	for i := 0; i < len(query); i++ {
		if query[i] == '(' {
			depth++
			current.WriteByte(query[i])
		} else if query[i] == ')' {
			depth--
			current.WriteByte(query[i])
		} else if depth == 0 && i+len(operator) <= len(query) && query[i:i+len(operator)] == operator {
			parts = append(parts, current.String())
			current.Reset()
			i += len(operator) - 1
		} else {
			current.WriteByte(query[i])
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

func parseFofaCondition(condition string) string {
	condition = strings.TrimSpace(condition)
	if condition == "" {
		return ""
	}

	condition = strings.Trim(condition, "()")
	condition = strings.TrimSpace(condition)

	fieldMap := map[string]string{
		"app":           "app",
		"title":         "title",
		"body":          "body",
		"header":        "header",
		"icon_hash":     "icon",
		"protocol":      "service",
		"banner":        "banner",
		"banner_hash":   "banner",
		"base_protocol": "service",
		"status_code":   "statuscode",
		"js_name":       "body",
	}

	patterns := []struct {
		regex   *regexp.Regexp
		opIndex int // 0: ==, 1: *=, 2: =, 3: !=
	}{

		{regexp.MustCompile(`^(\w+)\s*==\s*["']([^"']*)["']$`), 0},

		{regexp.MustCompile(`^(\w+)\s*\*=\s*["']([^"']*)["']$`), 1},

		{regexp.MustCompile(`^(\w+)\s*=\s*["']([^"']*)["']$`), 2},

		{regexp.MustCompile(`^(\w+)\s*!=\s*["']([^"']*)["']$`), 3},

		{regexp.MustCompile(`^(\w+)\s*==\s*(.+)$`), 0},

		{regexp.MustCompile(`^(\w+)\s*\*=\s*(.+)$`), 1},

		{regexp.MustCompile(`^(\w+)\s*=\s*(.+)$`), 2},

		{regexp.MustCompile(`^(\w+)\s*!=\s*(.+)$`), 3},
	}

	operators := []string{"==", "=", "=", "!="}

	for _, pattern := range patterns {
		matches := pattern.regex.FindStringSubmatch(condition)
		if len(matches) == 3 {
			field := strings.ToLower(matches[1])
			value := strings.TrimSpace(matches[2])

			value = strings.Trim(value, `"'`)

			customField, exists := fieldMap[field]
			if !exists {

				customField = "body"
			}

			operator := operators[pattern.opIndex]
			return fmt.Sprintf(`%s%s"%s"`, customField, operator, value)
		}
	}

	hasKnownField := false
	lowerCondition := strings.ToLower(condition)
	for field := range fieldMap {
		if strings.HasPrefix(lowerCondition, field+"=") ||
			strings.HasPrefix(lowerCondition, field+"==") ||
			strings.HasPrefix(lowerCondition, field+"!=") ||
			strings.HasPrefix(lowerCondition, field+"*=") {
			hasKnownField = true
			break
		}
	}

	if !hasKnownField && condition != "" {

		value := strings.Trim(condition, `"'`)

		value = strings.TrimPrefix(value, "=")
		value = strings.TrimPrefix(value, "==")
		value = strings.TrimPrefix(value, "!=")
		value = strings.TrimPrefix(value, "*=")
		value = strings.TrimSpace(value)
		if value != "" {
			return fmt.Sprintf(`body="%s"`, value)
		}
	}

	return ""
}
