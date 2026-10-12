// helper-------------------------------------
// @file      : shodan.go
// @time      : 2025/12/21 19:47
// -------------------------------------------

package helper

import (
	"fmt"
	"regexp"
	"strings"
)

func ConvertShodanToCustomQuery(shodanQueryValue interface{}) string {
	var queries []string

	if arr, ok := shodanQueryValue.([]interface{}); ok {
		for _, item := range arr {
			if str, ok := item.(string); ok {
				converted := parseShodanQuery(str)
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

	if str, ok := shodanQueryValue.(string); ok {
		return parseShodanQuery(str)
	}

	return ""
}

func parseShodanQuery(query string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return ""
	}

	return parseShodanQueryWithBrackets(query)
}

func parseShodanQueryWithBrackets(query string) string {
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
		parsedInner := parseShodanQueryWithBrackets(innerQuery)

		query = query[:start] + "(" + parsedInner + ")" + query[end+1:]
	}

	return parseShodanLogicExpression(query)
}

func parseShodanLogicExpression(query string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return ""
	}

	if hasOperator(query, "||") {
		parts := splitByShodanOperatorSymbol(query, "||")
		var results []string
		for _, part := range parts {
			parsed := parseShodanLogicExpression(strings.TrimSpace(part))
			if parsed != "" {
				results = append(results, parsed)
			}
		}
		if len(results) > 0 {
			return strings.Join(results, " || ")
		}
		return ""
	}

	if strings.Contains(strings.ToUpper(query), " OR ") {
		parts := splitByShodanOperator(query, " OR ")
		var results []string
		for _, part := range parts {
			parsed := parseShodanLogicExpression(strings.TrimSpace(part))
			if parsed != "" {
				results = append(results, parsed)
			}
		}
		if len(results) > 0 {
			return strings.Join(results, " || ")
		}
		return ""
	}

	if hasOperator(query, "&&") {
		parts := splitByShodanOperatorSymbol(query, "&&")
		var results []string
		for _, part := range parts {
			parsed := parseShodanLogicExpression(strings.TrimSpace(part))
			if parsed != "" {
				results = append(results, parsed)
			}
		}
		if len(results) > 0 {
			return strings.Join(results, " && ")
		}
		return ""
	}

	if strings.Contains(strings.ToUpper(query), " AND ") {
		parts := splitByShodanOperator(query, " AND ")
		var results []string
		for _, part := range parts {
			parsed := parseShodanLogicExpression(strings.TrimSpace(part))
			if parsed != "" {
				results = append(results, parsed)
			}
		}
		if len(results) > 0 {
			return strings.Join(results, " && ")
		}
		return ""
	}

	return parseShodanFilters(query)
}

func splitByShodanOperator(query, operator string) []string {
	var parts []string
	var current strings.Builder
	depth := 0
	operatorUpper := strings.ToUpper(operator)
	operatorLen := len(operator)

	for i := 0; i < len(query); i++ {
		if query[i] == '(' {
			depth++
			current.WriteByte(query[i])
		} else if query[i] == ')' {
			depth--
			current.WriteByte(query[i])
		} else if depth == 0 {

			if i+operatorLen <= len(query) {
				substr := strings.ToUpper(query[i : i+operatorLen])
				if substr == operatorUpper {
					parts = append(parts, current.String())
					current.Reset()
					i += operatorLen - 1
					continue
				}
			}
			current.WriteByte(query[i])
		} else {
			current.WriteByte(query[i])
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

func hasOperator(query, operator string) bool {
	inQuotes := false
	quoteChar := byte(0)
	depth := 0

	for i := 0; i <= len(query)-len(operator); i++ {
		char := query[i]

		if (char == '"' || char == '\'') && (i == 0 || query[i-1] != '\\') {
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
				quoteChar = 0
			}
			continue
		}

		if inQuotes {
			continue
		}

		if char == '(' {
			depth++
			continue
		} else if char == ')' {
			depth--
			continue
		}

		if depth == 0 && query[i:i+len(operator)] == operator {
			return true
		}
	}

	return false
}

func splitByShodanOperatorSymbol(query, operator string) []string {
	var parts []string
	var current strings.Builder
	depth := 0
	inQuotes := false
	quoteChar := byte(0)
	operatorLen := len(operator)

	for i := 0; i < len(query); i++ {
		char := query[i]

		if (char == '"' || char == '\'') && (i == 0 || query[i-1] != '\\') {
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
				quoteChar = 0
			}
			current.WriteByte(char)
			continue
		}

		if inQuotes {
			current.WriteByte(char)
			continue
		}

		if char == '(' {
			depth++
			current.WriteByte(char)
		} else if char == ')' {
			depth--
			current.WriteByte(char)
		} else if depth == 0 {

			if i+operatorLen <= len(query) && query[i:i+operatorLen] == operator {
				parts = append(parts, current.String())
				current.Reset()
				i += operatorLen - 1
				continue
			}
			current.WriteByte(char)
		} else {
			current.WriteByte(char)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

func parseShodanFilters(query string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return ""
	}

	query = strings.Trim(query, "()")
	query = strings.TrimSpace(query)

	filters := splitShodanFilters(query)
	if len(filters) == 0 {
		return ""
	}

	var results []string
	for _, filter := range filters {
		filter = strings.TrimSpace(filter)
		if filter == "" {
			continue
		}
		parsed := parseShodanFilter(filter)
		if parsed != "" {
			results = append(results, parsed)
		}
	}

	if len(results) == 0 {
		return ""
	}

	if len(results) == 1 {
		return results[0]
	}
	return strings.Join(results, " && ")
}

func splitShodanFilters(query string) []string {
	var filters []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(query); i++ {
		char := query[i]

		if (char == '"' || char == '\'') && (i == 0 || query[i-1] != '\\') {
			if !inQuotes {
				inQuotes = true
				quoteChar = char
				current.WriteByte(char)
			} else if char == quoteChar {
				inQuotes = false
				quoteChar = 0
				current.WriteByte(char)
			} else {
				current.WriteByte(char)
			}
			continue
		}

		if inQuotes {
			current.WriteByte(char)
			continue
		}

		if char == ' ' {
			if current.Len() > 0 {
				filters = append(filters, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteByte(char)
	}

	if current.Len() > 0 {
		filters = append(filters, current.String())
	}

	return filters
}

func parseShodanFilter(filter string) string {
	filter = strings.TrimSpace(filter)
	if filter == "" {
		return ""
	}

	fieldMap := map[string]string{
		"http.title":        "title",
		"http.html":         "body",
		"http.component":    "app",
		"http.server":       "header",
		"http.favicon.hash": "icon",
		"http.status":       "statuscode",
		"title":             "title",
		"html":              "body",
		"component":         "app",
		"server":            "header",
		"favicon.hash":      "icon",
		"status":            "statuscode",
	}

	isNot := false
	if strings.HasPrefix(filter, "-") {
		isNot = true
		filter = strings.TrimPrefix(filter, "-")
		filter = strings.TrimSpace(filter)
	} else {

		filterUpper := strings.ToUpper(filter)
		if strings.HasPrefix(filterUpper, "NOT ") {
			isNot = true

			notIndex := strings.Index(filterUpper, "NOT ")
			if notIndex == 0 {
				filter = filter[4:]
				filter = strings.TrimSpace(filter)
			}
		}
	}

	pattern := regexp.MustCompile(`^([\w.]+)\s*:\s*(.+)$`)
	matches := pattern.FindStringSubmatch(filter)
	if len(matches) == 3 {
		field := strings.ToLower(matches[1])
		value := strings.TrimSpace(matches[2])

		if field == "country" {
			return ""
		}

		value = strings.Trim(value, `"'`)

		customField, exists := fieldMap[field]
		if !exists {

			customField = "body"
		}

		operator := "="
		if isNot {
			operator = "!="
		}

		return fmt.Sprintf(`%s%s"%s"`, customField, operator, value)
	}

	if !strings.Contains(filter, ":") {
		value := strings.Trim(filter, `"'`)
		operator := "="
		if isNot {
			operator = "!="
		}
		return fmt.Sprintf(`body%s"%s"`, operator, value)
	}

	return ""
}
