// nuclei_monitor-------------------------------------
// @file      : nuclei_monitor.go
// @time      : 2025/12/20 20:04
// -------------------------------------------

package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"recon/internal/utils/helper"

	"recon/internal/models"
	"recon/internal/options"
	"recon/internal/utils"
	"recon/internal/utils/random"
)

type templateResult struct {
	TemplateID string `json:"template_id"`
	Severity   string `json:"severity"`
	Name       string `json:"name"`
	Raw        string `json:"raw"`
}

type apiResponse struct {
	Results []templateResult `json:"results"`
}

func GetName() string {
	return "Nuclei Monitor"
}

func formatSeverityWithColor(severity string) string {
	severityUpper := strings.ToUpper(severity)
	var color string
	var bold bool

	switch severityUpper {
	case "CRITICAL":
		color = "#ff6b6b"
		bold = true
	case "HIGH":
		color = "#ff8787"
		bold = true
	case "MEDIUM":
		color = "#ffd43b"
		bold = false
	case "LOW":
		color = "#51cf66"
		bold = false
	case "INFO":
		color = "#74c0fc"
		bold = false
	default:

		return severity
	}

	if bold {
		return fmt.Sprintf(`<color value="%s" bold>%s</color>`, color, severity)
	}
	return fmt.Sprintf(`<color value="%s">%s</color>`, color, severity)
}

func formatAssetCountWithColor(count int64) string {
	if count > 0 {
		return fmt.Sprintf(`<color value="#ff6b6b" bold>%v</color>`, count)
	}
	return fmt.Sprintf("%v", count)
}

func Cycle() string {
	return "0 * * * *"
}

func Install() error {
	return nil
}

//

//

var CreatTask = false

var SkipScanName = []string{
	"WordPress",
}

func Execute(op options.PluginOption) error {
	value := op.GetStringVariable("nuclei_init")
	if value != "true" {

		tempDir := os.TempDir()
		fileName := random.GenerateString(5) + ".zip"
		filePath := filepath.Join(tempDir, fileName)
		op.Log("download ....https://codeload.github.com/projectdiscovery/nuclei-templates")
		err := utils.DownloadFile("https://codeload.github.com/projectdiscovery/nuclei-templates/zip/refs/heads/main", filePath, 10*time.Minute)
		if err != nil {
			op.Log(fmt.Sprintf("download fail: %s", err.Error()), "e")
			return err
		}
		_, err = op.PocService.ImportPoc(context.Background(), filePath)
		if err != nil {
			op.Log(fmt.Sprintf("ImportPoc fail: %s", err.Error()), "e")
			return err
		}
		op.SetStringVariable("nuclei_init", "true")
		op.Log("download nuclei-templates success")
		return nil
	}

	ctx := context.Background()
	offset := 0
	limit := 20

	for {
		results, err := fetchPocTemplates(offset, limit)
		if err != nil {
			op.Log(fmt.Sprintf("Failed to fetch POC template: %s", err.Error()), "e")
			return err
		}

		if len(results) == 0 {
			break
		}

		for _, result := range results {
			result.TemplateID = strings.Replace(result.TemplateID, "-draft", "", 1)

			exists, err := op.PocService.TemplateIdExists(ctx, result.TemplateID)
			if err != nil {
				op.Log(fmt.Sprintf("Failed to check TemplateId %s: %s", result.TemplateID, err.Error()), "e")
				continue
			}

			if exists {

				op.Log("All POCs are up to date")
				return nil
			}

			pocID, err := op.PocService.AddPoc(context.Background(), &models.PocAddRequest{Content: result.Raw})
			if err != nil {
				op.Log(fmt.Sprintf("AddPoc error: %s", err.Error()), "e")
				return err
			}

			time.Sleep(3 * time.Second)

			var pt models.PocTemplate
			err = utils.Unmarshal([]byte(result.Raw), &pt)
			if err != nil {
				return fmt.Errorf("failed to unmarshal poc template: %w", err)
			}
			var searchValueOld interface{}
			var customQuery string
			var scanStatus = "Not created"
			var affectedAssetCount = int64(0)
			if pt.Info.Metadata != nil {

				if fofaQueryValue := pt.Info.Metadata["fofa-query"]; fofaQueryValue != nil {
					searchValueOld = fofaQueryValue
					customQuery = helper.ConvertFofaToCustomQuery(fofaQueryValue)

				}

				if customQuery == "" {
					if shodanQueryValue := pt.Info.Metadata["shodan-query"]; shodanQueryValue != nil {
						searchValueOld = shodanQueryValue
						customQuery = helper.ConvertShodanToCustomQuery(shodanQueryValue)

					}
				}
				if customQuery != "" {

					searchRequest := models.SearchRequest{
						Index:            "asset",
						SearchExpression: customQuery,
					}
					affectedAssetCount, err = op.AssetCommonService.TotalData(context.Background(), &searchRequest)
					if err != nil {
						return err
					}

					if CreatTask && affectedAssetCount != 0 {

						skipFlag := false
						for _, v := range SkipScanName {
							if strings.Contains(result.Name, strings.ToLower(v)) {
								skipFlag = true
								scanStatus = "Skipped"
								break
							}
						}

						template := models.ScanTemplate{
							Name:              fmt.Sprintf("nuclei-template-%v-%v", result.Name, helper.GetNowTimeString()),
							VulnerabilityScan: []string{"ed93b8af6b72fe54a60efdb932cf6fbc"}, //nuclei
							VulList:           []string{pocID},
							Parameters: models.Parameters{
								VulnerabilityScan: map[string]string{
									"ed93b8af6b72fe54a60efdb932cf6fbc": "",
								},
							},
						}
						templateID, err := op.TemplateService.Save(context.Background(), "", &template)
						if err != nil {
							op.Log(fmt.Sprintf("Failed to check TemplateId %s: %s", pocID, err.Error()), "e")
							return err
						}

						if !skipFlag {
							task := models.Task{
								Name:           fmt.Sprintf("[nuclei monitor ]-%v-%v[plugin]", result.Name, helper.GetNowTimeString()),
								AllNode:        true,
								Duplicates:     "None",
								ScheduledTasks: false,
								Template:       templateID,
								TargetSource:   "asset",
								Search:         customQuery,
							}
							_, err := op.TaskCommonService.Insert(context.Background(), &task)
							if err != nil {
								return err
							}
							scanStatus = "Created"
						}
					}
				}
			}
			severityFormatted := formatSeverityWithColor(result.Severity)
			assetCountFormatted := formatAssetCountWithColor(affectedAssetCount)
			op.Log(fmt.Sprintf("New POC: TemplateID=%s, Name=%s, Severity=%s, OldSearch=%v, SaSSearch=%v, affected assets:%s, task status:%v", result.TemplateID, result.Name, severityFormatted, searchValueOld, customQuery, assetCountFormatted, scanStatus))
			notificationMsg := fmt.Sprintf("New POC: \nTemplateID:%s\nName:%s\nSeverity:%s\nOldSearch:%v\nSaSSearch:%v\nAffected assets:%v\nTask status:%v", result.TemplateID, result.Name, result.Severity, searchValueOld, customQuery, affectedAssetCount, scanStatus)
			op.Notification(notificationMsg)
		}

		if len(results) < limit {
			break
		}

		offset += limit
	}

	return nil
}

func fetchPocTemplates(offset, limit int) ([]templateResult, error) {

	reqURL := "https://api.projectdiscovery.io/v2/template/search"
	params := url.Values{}
	params.Set("scope", "public")
	params.Set("offset", strconv.Itoa(offset))
	params.Set("limit", strconv.Itoa(limit))
	params.Set("isSignedIn", "false")

	fullURL := reqURL + "?" + params.Encode()

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create request: %w", err)
	}

	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("origin", "https://cloud.projectdiscovery.io")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("referer", "https://cloud.projectdiscovery.io/")
	req.Header.Set("sec-ch-ua", `"Chromium";v="136", "Google Chrome";v="136", "Not.A/Brand";v="99"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-site")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36")
	req.Header.Set("x-team-id", "cus16bdim6vs73cqnh0g")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected HTTP status: %d", resp.StatusCode)
	}

	var apiResp apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("Failed to parse response: %w", err)
	}

	return apiResp.Results, nil
}
