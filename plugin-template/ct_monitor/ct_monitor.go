// Package plugin plugin-------------------------------------
// @file      : ct_monitor.go
// @time      : 2026/1/25 13:05
// -------------------------------------------
package plugin

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"recon/internal/utils/helper"

	"recon/internal/models"

	"recon/internal/utils/ct"
	ct_models "recon/internal/utils/ct/models"

	"go.mongodb.org/mongo-driver/bson"
	"recon/internal/options"
	"recon/internal/utils"
	"recon/internal/utils/ct/types"
)

func GetName() string {
	return "CT Monitor"
}

func Cycle() string {
	return "1"
}

func Install() error {
	return nil
}

var blacklist = []string{
	"google.com",
}

var subdomainBlacklist = []string{
	".prod.",
}
var CTSDK *ct.CTWatcherSDK
var OP options.PluginOption

var (
	processedSubdomains = make(map[string]struct{})
	subdomainMu         sync.RWMutex
)

func Execute(op options.PluginOption) error {

	ctx := utils.GetContext(op.PluginHash)

	opts := &types.SDKOptions{

		LogServers: []*types.LogServerConfig{
			{
				URL:           "https://ct.googleapis.com/logs/us1/argon2026h1/",
				BatchSize:     500,
				ParallelFetch: 10,
				BufferSize:    1000,
				Timeout:       30 * time.Second,
				MaxRetries:    3,
			},
			//{
			//	URL:           "https://ct.googleapis.com/logs/us1/argon2026h2/",
			//	BatchSize:     500,
			//	ParallelFetch: 10,
			//	BufferSize:    1000,
			//	Timeout:       30 * time.Second,
			//	MaxRetries:    3,
			//},
			//{
			//	URL:           "https://ct.googleapis.com/logs/us1/argon2027h1/",
			//	BatchSize:     500,
			//	ParallelFetch: 10,
			//	BufferSize:    1000,
			//	Timeout:       30 * time.Second,
			//	MaxRetries:    3,
			//},
			{
				URL:           "https://ct.googleapis.com/logs/eu1/xenon2026h1/",
				BatchSize:     500,
				ParallelFetch: 10,
				BufferSize:    1000,
				Timeout:       30 * time.Second,
				MaxRetries:    3,
			},
			//{
			//	URL:           "https://ct.googleapis.com/logs/eu1/xenon2026h2/",
			//	BatchSize:     500,
			//	ParallelFetch: 10,
			//	BufferSize:    1000,
			//	Timeout:       30 * time.Second,
			//	MaxRetries:    3,
			//},
			//{
			//	URL:           "https://ct.googleapis.com/logs/eu1/xenon2027h1/",
			//	BatchSize:     500,
			//	ParallelFetch: 10,
			//	BufferSize:    1000,
			//	Timeout:       30 * time.Second,
			//	MaxRetries:    3,
			//},
			//{
			//	URL:           "https://ct.cloudflare.com/logs/nimbus2026/",
			//	BatchSize:     500,
			//	ParallelFetch: 10,
			//	BufferSize:    1000,
			//	Timeout:       30 * time.Second,
			//	MaxRetries:    3,
			//},
			//{
			//	URL:           "https://ct.cloudflare.com/logs/nimbus2027/",
			//	BatchSize:     500,
			//	ParallelFetch: 10,
			//	BufferSize:    1000,
			//	Timeout:       30 * time.Second,
			//	MaxRetries:    3,
			//},
		},

		DomainMatcher: &types.DomainMatcherConfig{
			EnableBloomFilter: true,
			BloomFilterSize:   1000000,
			BloomFilterHash:   5,
			EnableLRUCache:    true,
			LRUCacheSize:      10000,
			CaseSensitive:     false,
			MaxDomainLength:   253,
		},

		Performance: &types.PerformanceConfig{
			MaxWorkers:        10,
			ChannelBufferSize: 100,
			ProcessingTimeout: 5 * time.Second,
			BatchProcessSize:  50,
		},

		Redis: &types.RedisConfig{
			KeyPrefix: "ct_watcher:",
		},
	}

	filter := bson.M{
		"domain": bson.M{
			"$nin": blacklist,
		},
	}
	var rootDomains []bson.M
	err := op.CommonRepo.FindField(ctx, "RootDomain", filter, 0, []string{"domain"}, &rootDomains)
	if err != nil {
		return err
	}

	domains := make([]string, 0, len(rootDomains))
	for _, doc := range rootDomains {
		if domain, ok := doc["domain"].(string); ok {
			domains = append(domains, domain)
		}
	}

	opts.Domains = domains

	sdk, err := ct.NewCTWatcherSDK(opts)
	if err != nil {
		log.Fatalf("Failed to create SDK: %v", err)
	}
	defer sdk.Close()

	CTSDK = sdk
	OP = op

	taskTemplateId := ""
	ctTmp, err := op.TemplateService.List(context.Background(), 1, 20, "CT_Monitor")
	if err != nil {
		op.Log(fmt.Sprintf("Search Template error: %s", err.Error()), "e")
	}
	if ctTmp != nil {
		if len(ctTmp.List) == 0 {

			template := models.ScanTemplate{
				Name:         "CT_Monitor",
				AssetMapping: []string{"3a0d994a12305cb15a5cb7104d819623"}, // httpx
				Parameters: models.Parameters{
					AssetMapping: map[string]string{
						"3a0d994a12305cb15a5cb7104d819623": "-cdncheck false -screenshot true -tlsprobe false -st 10 -fr true -bh false -t 30",
					},
				},
				ParameterLists: models.Parameters{
					AssetMapping: map[string]string{
						"3a0d994a12305cb15a5cb7104d819623": "[{\"name\":\"cdncheck\",\"type\":\"bool\",\"defaultValue\":\"false\"},{\"name\":\"screenshot\",\"type\":\"bool\",\"defaultValue\":\"true\"},{\"name\":\"tlsprobe\",\"type\":\"bool\",\"defaultValue\":\"false\"},{\"name\":\"st\",\"type\":\"string\",\"defaultValue\":\"10\"},{\"name\":\"fr\",\"type\":\"bool\",\"defaultValue\":\"true\"},{\"name\":\"bh\",\"type\":\"bool\",\"defaultValue\":\"false\"},{\"name\":\"t\",\"type\":\"string\",\"defaultValue\":\"30\"}]",
					},
				},
			}
			templateID, err := op.TemplateService.Save(context.Background(), "", &template)
			if err != nil {
				op.Log(fmt.Sprintf("Failed to create template: %s", err.Error()), "e")
				return err
			}
			taskTemplateId = templateID
		} else {
			taskTemplateId = ctTmp.List[0].ID
		}
	}

	sdk.OnSubdomainFound(func(event ct_models.SubdomainEvent) {

		subdomainMu.RLock()
		_, exists := processedSubdomains[event.Subdomain]
		subdomainMu.RUnlock()

		if exists {

			return
		}

		for _, domain := range subdomainBlacklist {
			if strings.Contains(event.Subdomain, domain) {
				return
			}
		}

		subdomainMu.Lock()
		processedSubdomains[event.Subdomain] = struct{}{}
		subdomainMu.Unlock()

		subdomainKey := "ct_watcher:subdomains"
		err := op.CommonRepo.SAdd(ctx, subdomainKey, event.Subdomain)
		if err != nil {
			op.Log(fmt.Sprintf("[Error] Failed to store subdomain %s in Redis: %v", event.Subdomain, err), "e")
		}
		op.Log(fmt.Sprintf("[Subdomain found] %s -> %s | certificate index: %d | time: %s\n",
			fmt.Sprintf("<color value=\"#ff6b6b\" bold>%s</color>", event.RootDomain),
			fmt.Sprintf("<color value=\"#ff6b6b\" bold>%s</color>", event.Subdomain),
			event.Certificate.Index,
			event.DiscoveredAt.Format("2006-01-02 15:04:05")))

		notificationMsg := fmt.Sprintf("New subdomains: \n%v", event.Subdomain)
		op.Notification(notificationMsg)
	})

	sdk.OnError(func(err error, context map[string]interface{}) {
		errorType := "unknown"
		if ctxType, ok := context["error_type"]; ok {
			if et, ok := ctxType.(string); ok {
				errorType = et
			}
		}
		op.Log(fmt.Sprintf("[Error] Type: %s | message: %v\n", errorType, err), "e")
	})

	op.Log("Starting CT log watchers...")
	if err := sdk.Start(ctx); err != nil {
		log.Fatalf("Failed to start SDK: %v", err)
		op.Log(fmt.Sprintf("Failed to start SDK: %v", err), "e")
		return err
	}
	op.Log("CT watchers started; watching for subdomains...")

	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		var lastProcessedCount int64 = 0
		var lastReportTime = time.Now()

		reportStatus := func() {
			currentTime := time.Now()
			status := sdk.GetStatus()
			stats := sdk.GetStats()

			uptimeSeconds := status.Uptime.Seconds()
			avgQps := float64(0)
			if uptimeSeconds > 0 {
				avgQps = float64(status.ProcessedCertificates) / uptimeSeconds
			}

			timeDiff := currentTime.Sub(lastReportTime).Seconds()
			countDiff := status.ProcessedCertificates - lastProcessedCount
			incrementalQps := float64(0)
			if timeDiff > 0 {
				incrementalQps = float64(countDiff) / timeDiff
			}

			lastProcessedCount = status.ProcessedCertificates
			lastReportTime = currentTime

			var logContent string
			logContent += "\n=== Watcher status report ===\n"
			logContent += fmt.Sprintf("Runtime: %v\n", status.Uptime)
			logContent += fmt.Sprintf("Total certificates processed: %d\n", status.ProcessedCertificates)
			logContent += fmt.Sprintf("Total subdomains found: %d\n", status.DiscoveredSubdomains)
			logContent += fmt.Sprintf("Average QPS: %.2f certs/sec\n", avgQps)
			logContent += fmt.Sprintf("Recent QPS: %.2f certs/sec (last 30 seconds)\n", incrementalQps)
			logContent += fmt.Sprintf("Active watchers: %d\n", status.ActiveWatchers)
			logContent += fmt.Sprintf("Total domains: %d\n", stats.TotalDomains)

			if len(status.WatcherStatuses) > 0 {
				totalWatcherProcessed := int64(0)
				for url, watcherStatus := range status.WatcherStatuses {
					logContent += fmt.Sprintf("\nWatcher: %s\n", url)
					logContent += fmt.Sprintf("  Status: %s\n", watcherStatus.Status)
					logContent += fmt.Sprintf("  Certificates processed: %d\n", watcherStatus.ProcessedCount)
					logContent += fmt.Sprintf("  Last index: %d\n", watcherStatus.LastIndex)
					logContent += fmt.Sprintf("  Error count: %d\n", watcherStatus.ErrorCount)
					totalWatcherProcessed += watcherStatus.ProcessedCount
				}
				logContent += fmt.Sprintf("\nTotal watcher certificates processed: %d\n", totalWatcherProcessed)
			}
			logContent += "========================\n\n"
			op.Log(logContent)
		}

		reportStatus()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				reportStatus()
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				processSubdomains(ctx, op, taskTemplateId)
			}
		}
	}()

	<-ctx.Done()
	op.Log("CT watchers stopped")
	return nil
}

func processSubdomains(ctx context.Context, op options.PluginOption, taskTemplateId string) {
	subdomainKey := "ct_watcher:subdomains"

	count, err := op.CommonRepo.SCard(ctx, subdomainKey)
	if err != nil {
		op.Log(fmt.Sprintf("[Error] Failed to check subdomain set: %v", err), "e")
		return
	}

	if count == 0 {
		op.Log("Redis subdomain set is empty; skipping")
		return
	}

	subdomains, err := op.CommonRepo.SMembers(ctx, subdomainKey)
	if err != nil {
		op.Log(fmt.Sprintf("[Error] Failed to list subdomains: %v", err), "e")
		return
	}

	op.Log(fmt.Sprintf("Retrieved %d subdomains from Redis; creating a task", len(subdomains)))
	target := ""
	for _, subdomain := range subdomains {
		if subdomain != "" {
			target += subdomain + "\n"
		}
	}

	if target != "" {
		task := models.Task{
			Name:           fmt.Sprintf("[ct monitor]-%v[plugin]", helper.GetNowTimeString()),
			AllNode:        true,
			Duplicates:     "None",
			ScheduledTasks: false,
			Template:       taskTemplateId,
			TargetSource:   "general",
			Target:         target,
		}
		_, err := op.TaskCommonService.Insert(context.Background(), &task)
		if err != nil {
			op.Log(fmt.Sprintf("[Error] Failed to create task: %v", err), "e")
			return
		}

		members := make([]interface{}, len(subdomains))
		for i, subdomain := range subdomains {
			members[i] = subdomain
		}
		err = op.CommonRepo.SRem(ctx, subdomainKey, members...)
		if err != nil {
			op.Log(fmt.Sprintf("[Error] Failed to delete Redis subdomains: %v", err), "e")
		} else {
			op.Log(fmt.Sprintf("Deleted %d processed subdomains from Redis", len(subdomains)))
		}
	}

}

func TaskEnd(task models.Task, plgHash string) error {
	if CTSDK == nil {
		OP.Log("CT SDK is not initialized")
		return fmt.Errorf("CTSDK is null")
	}

	if strings.Contains(task.Name, "[plugin]") {
		return nil
	}
	filter := bson.M{
		"taskName": task.Name,
	}
	var rootDomains []bson.M
	err := OP.CommonRepo.FindField(context.Background(), "RootDomain", filter, 0, []string{"domain"}, &rootDomains)
	if err != nil {
		return err
	}

	domains := make([]string, 0, len(rootDomains))
	for _, doc := range rootDomains {
		if domain, ok := doc["domain"].(string); ok {
			domains = append(domains, domain)
		}
	}
	err = CTSDK.AddDomains(domains)
	if err != nil {
		OP.Log(fmt.Sprintf("CT SDK failed to add domain: %v", err), "e")
		return err
	}
	OP.Log(fmt.Sprintf("Added %v watched domains to CT monitoring", len(domains)))
	return nil
}
