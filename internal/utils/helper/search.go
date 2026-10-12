// Package helper -----------------------------
// @file      : search.go
// @time      : 2025/5/2 14:34
// -------------------------------------------
package helper

import (
	"errors"
	"fmt"
	"recon/internal/logger"
	"regexp"
	"strings"
	"sync"

	"recon/internal/models"
)

var SearchKey = map[string]map[string]string{
	"SensitiveResult": {
		"url":   "url",
		"sname": "sid",
		"body":  "body",
		"info":  "match",
		"md5":   "md5",
	},
	"DirScanResult": {
		"statuscode": "status",
		"url":        "url",
		"redirect":   "msg",
		"length":     "length",
	},
	"vulnerability": {
		"url":      "url",
		"vulname":  "vulname",
		"matched":  "matched",
		"request":  "request",
		"response": "response",
		"level":    "level",
	},
	"subdomain": {
		"domain": "host",
		"ip":     "ip",
		"type":   "type",
		"value":  "value",
	},
	"asset": {
		"app":        "technologies",
		"body":       "body",
		"header":     "rawheaders",
		"title":      "title",
		"statuscode": "statuscode",
		"icon":       "faviconmmh3",
		"ip":         "ip",
		"domain":     "host",
		"port":       "port",
		"service":    "service",
		"banner":     "metadata",
		"type":       "type",
	},
	"SubdomainTakerResult": {
		"domain":   "input",
		"value":    "value",
		"type":     "cname",
		"response": "response",
	},
	"UrlScan": {
		"url":      "output",
		"input":    "input",
		"source":   "source",
		"resultId": "resultId",
		"type":     "outputtype",
	},
	"PageMonitoring": {
		"url":      "url",
		"hash":     "hash",
		"diff":     "diff",
		"response": "response",
	},
	"crawler": {
		"url":      "url",
		"method":   "method",
		"body":     "body",
		"resultId": "resultId",
	},
	"RootDomain": {
		"domain":  "domain",
		"icp":     "icp",
		"company": "company",
	},
	"app": {
		"name":        "name",
		"icp":         "icp",
		"company":     "company",
		"category":    "category",
		"description": "description",
		"url":         "url",
		"apk":         "apk",
	},
	"mp": {
		"name":        "name",
		"icp":         "icp",
		"company":     "company",
		"category":    "category",
		"description": "description",
		"url":         "url",
	},
	"IPAsset": {
		"ip":        "ip",
		"domain":    "ports.server.domain",
		"port":      "ports.port",
		"service":   "ports.server.service",
		"webServer": "ports.server.webServer",
		"app":       "ports.server.technologies",
	},
}

var (
	filterKeyCache = map[string]string{
		"app":        "technologies",
		"color":      "color",
		"status":     "status",
		"level":      "level",
		"type":       "type",
		"project":    "project",
		"port":       "port",
		"service":    "service",
		"icon":       "faviconmmh3",
		"statuscode": "statuscode",
		"sname":      "sid",
		"task":       "taskName",
		"tags":       "tags",
	}

	fuzzyQueryKeyCache = map[string]string{
		"sub_host":        "host",
		"sub_value":       "value",
		"sub_ip":          "ip",
		"port_port":       "port",
		"port_domain":     "domain,host",
		"port_ip":         "ip,host",
		"port_protocol":   "service",
		"service_service": "type,webServer,protocol",
		"service_domain":  "domain,host",
		"service_port":    "port",
		"service_ip":      "ip,host",
	}

	regexCache = struct {
		sync.RWMutex
		m map[string]*regexp.Regexp
	}{
		m: make(map[string]*regexp.Regexp),
	}
)

func getRegex(pattern string) (*regexp.Regexp, error) {
	regexCache.RLock()
	if re, ok := regexCache.m[pattern]; ok {
		regexCache.RUnlock()
		return re, nil
	}
	regexCache.RUnlock()

	regexCache.Lock()
	defer regexCache.Unlock()

	if re, ok := regexCache.m[pattern]; ok {
		return re, nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	regexCache.m[pattern] = re
	return re, nil
}

func StringToPostfix(expression string) ([]string, error) {
	if expression == "" {
		return nil, nil
	}

	operandsStack := make([]string, 0, 16)
	expressionStack := make([]string, 0, 32)
	startChar := 0
	skipFlag := false
	expFlag := false

	runes := []rune(expression)
	for i := 0; i < len(runes); i++ {
		char := runes[i]
		if skipFlag {
			skipFlag = false
			continue
		}

		if i < len(runes)-1 {
			if char == '|' && runes[i+1] == '|' {
				skipFlag = true
				operandsStack = append(operandsStack, "||")
				key := string(runes[startChar:i])
				if key != "" {
					expressionStack = append(expressionStack, key)
				}
				startChar = i + 2
				i++
			} else if char == '&' && runes[i+1] == '&' {
				skipFlag = true
				operandsStack = append(operandsStack, "&&")
				key := string(runes[startChar:i])
				if key != "" {
					expressionStack = append(expressionStack, key)
				}
				startChar = i + 2
				i++
			}
		}

		if char == '(' && (i == 0 || runes[i-1] != '\\') && !expFlag {
			startChar = i + 1
			operandsStack = append(operandsStack, "(")
		} else if char == ')' && (i == 0 || runes[i-1] != '\\') && !expFlag {
			key := string(runes[startChar:i])
			if key != "" {
				expressionStack = append(expressionStack, key)
			}
			startChar = i + 1

			for len(operandsStack) > 0 {
				popped := operandsStack[len(operandsStack)-1]
				operandsStack = operandsStack[:len(operandsStack)-1]
				if popped == "(" {
					break
				}
				if popped != "" {
					expressionStack = append(expressionStack, popped)
				}
			}
		} else if char == '"' && (i == 0 || runes[i-1] != '\\') {
			if !expFlag {
				expFlag = true
			} else {
				if i == len(runes)-1 {
					expFlag = false
					continue
				}
				tmp := strings.ReplaceAll(string(runes[i:]), " ", "")
				if strings.HasPrefix(tmp, "\"||") ||
					(strings.HasPrefix(tmp, "\")") && len(tmp) == 3) ||
					strings.HasPrefix(tmp, "\"&&") ||
					strings.HasPrefix(tmp, "\")||") ||
					strings.HasPrefix(tmp, "\")&&") ||
					(strings.HasPrefix(tmp, "\")") && len(tmp) == 2) {
					expFlag = false
				}
			}
		}
	}

	if startChar < len(runes) {
		key := string(runes[startChar:])
		if key != "" {
			expressionStack = append(expressionStack, key)
		}
	}

	for len(operandsStack) > 0 {
		popped := operandsStack[len(operandsStack)-1]
		operandsStack = operandsStack[:len(operandsStack)-1]
		if popped != "" {
			expressionStack = append(expressionStack, popped)
		}
	}

	result := make([]string, 0, len(expressionStack))
	for _, key := range expressionStack {
		if key != "" && key != " " {
			cleaned := strings.TrimSpace(key)
			cleaned = strings.ReplaceAll(cleaned, `\(`, "(")
			cleaned = strings.ReplaceAll(cleaned, `\)`, ")")
			cleaned = strings.ReplaceAll(cleaned, `\|\|`, "||")
			cleaned = strings.ReplaceAll(cleaned, `\&\&`, "&&")
			result = append(result, cleaned)
		}
	}

	return result, nil
}

func SearchToMongoDB(expressionRaw string, keyword map[string]string) ([]map[string]interface{}, error) {
	if expressionRaw == "" {
		return []map[string]interface{}{{}}, nil
	}

	keyword["task"] = "taskName"
	keyword["rootDomain"] = "rootDomain"
	keyword["tag"] = "tags"

	expression, err := StringToPostfix(expressionRaw)
	if err != nil {
		return nil, err
	}

	stack := make([]map[string]interface{}, 0, len(expression))

	for _, expr := range expression {
		if expr == "&&" {
			if len(stack) < 2 {
				return nil, errors.New("invalid expression: insufficient operands for &&")
			}
			right := stack[len(stack)-1]
			left := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			stack = append(stack, map[string]interface{}{
				"$and": []map[string]interface{}{left, right},
			})
		} else if expr == "||" {
			if len(stack) < 2 {
				return nil, errors.New("invalid expression: insufficient operands for ||")
			}
			right := stack[len(stack)-1]
			left := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			stack = append(stack, map[string]interface{}{
				"$or": []map[string]interface{}{left, right},
			})
		} else if strings.Contains(expr, "!=") {
			parts := strings.SplitN(expr, "!=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			value = strings.Trim(value, "\"")

			if mongoKey, ok := keyword[key]; ok {
				if strings.HasSuffix(key, "code") || key == "length" {
					numValue := 0
					if _, err := fmt.Sscanf(value, "%d", &numValue); err == nil {
						stack = append(stack, map[string]interface{}{
							mongoKey: map[string]interface{}{
								"$ne": numValue,
							},
						})
					}
				} else {
					if value == "" {
						stack = append(stack, map[string]interface{}{
							"$and": []map[string]interface{}{
								{mongoKey: map[string]interface{}{"$ne": ""}},
								{mongoKey: map[string]interface{}{"$ne": nil}},
								{mongoKey: map[string]interface{}{"$not": map[string]interface{}{"$size": 0}}},
							},
						})
					} else {
						stack = append(stack, map[string]interface{}{
							mongoKey: map[string]interface{}{
								"$not": map[string]interface{}{
									"$regex":   value,
									"$options": "i",
								},
							},
						})
					}
				}
			}
		} else if strings.Contains(expr, "==") {
			parts := strings.SplitN(expr, "==", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			value = strings.Trim(value, "\"")

			if mongoKey, ok := keyword[key]; ok {
				if strings.HasSuffix(key, "code") || key == "length" {
					numValue := 0
					if _, err := fmt.Sscanf(value, "%d", &numValue); err == nil {
						stack = append(stack, map[string]interface{}{
							mongoKey: map[string]interface{}{
								"$eq": numValue,
							},
						})
					}
				} else {
					stack = append(stack, map[string]interface{}{
						mongoKey: map[string]interface{}{
							"$eq": value,
						},
					})
				}
			}
		} else if strings.Contains(expr, "=") {
			parts := strings.SplitN(expr, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			value = strings.Trim(value, "\"")

			if mongoKey, ok := keyword[key]; ok {
				stack = append(stack, map[string]interface{}{
					mongoKey: map[string]interface{}{
						"$regex":   value,
						"$options": "i",
					},
				})
			}
		}
	}

	return stack, nil
}

func GetSearchQuery(req models.SearchRequest) (map[string]interface{}, error) {
	logger.Logger.Info(fmt.Sprintf("search query: %v", req))
	if req.Index == "" {
		return nil, errors.New("search type is required")
	}

	keyword := SearchKey[req.Index]
	if keyword == nil {
		return nil, errors.New("invalid search type")
	}

	query, err := SearchToMongoDB(req.SearchExpression, keyword)
	if err != nil {
		return nil, err
	}

	if len(query) == 0 {
		return map[string]interface{}{}, nil
	}

	result := query[0]

	if len(req.Filter) > 0 {
		if _, ok := result["$and"]; !ok {
			result["$and"] = make([]map[string]interface{}, 0, len(req.Filter))
		}

		for f, values := range req.Filter {
			if mongoKey, ok := filterKeyCache[f]; ok {
				tmpOr := make([]map[string]interface{}, 0, len(values))
				for _, v := range values {
					if v != "" {
						if strings.Contains(mongoKey, ",") {
							keys := strings.Split(mongoKey, ",")
							for _, key := range keys {
								tmpOr = append(tmpOr, map[string]interface{}{
									key: v,
								})
							}
						} else {
							tmpOr = append(tmpOr, map[string]interface{}{
								mongoKey: v,
							})
						}
					}
				}
				if len(tmpOr) > 0 {
					result["$and"] = append(result["$and"].([]map[string]interface{}), map[string]interface{}{
						"$or": tmpOr,
					})
				}
			}
		}
	}

	if len(req.FuzzyQuery) > 0 {
		if _, ok := result["$and"]; !ok {
			result["$and"] = make([]map[string]interface{}, 0, len(req.FuzzyQuery))
		}

		for q, value := range req.FuzzyQuery {
			if value != "" {
				tmpFq := make([]map[string]interface{}, 0, 3)
				if mongoKey, ok := fuzzyQueryKeyCache[q]; ok {
					if strings.Contains(mongoKey, ",") {
						keys := strings.Split(mongoKey, ",")
						for _, key := range keys {
							tmpFq = append(tmpFq, map[string]interface{}{
								key: map[string]interface{}{
									"$regex":   value,
									"$options": "i",
								},
							})
						}
					} else {
						tmpFq = append(tmpFq, map[string]interface{}{
							mongoKey: map[string]interface{}{
								"$regex":   value,
								"$options": "i",
							},
						})
					}
				}
				if len(tmpFq) > 0 {
					result["$and"] = append(result["$and"].([]map[string]interface{}), map[string]interface{}{
						"$or": tmpFq,
					})
				}
			}
		}
	}

	if and, ok := result["$and"].([]map[string]interface{}); ok && len(and) == 0 {
		delete(result, "$and")
	}

	return result, nil
}
