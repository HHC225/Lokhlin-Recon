//go:build ignore

// examples----------------------------------
// @file      : basic_usage.go
// @time      : 2025/01/24 21:00
// -------------------------------------------

// ============================================

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"recon/internal/utils/ct"
	"recon/internal/utils/ct/models"
	"recon/internal/utils/ct/types"
)

func main() {

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
			{
				URL:           "https://ct.googleapis.com/logs/us1/argon2026h2/",
				BatchSize:     500,
				ParallelFetch: 10,
				BufferSize:    1000,
				Timeout:       30 * time.Second,
				MaxRetries:    3,
			},
			{
				URL:           "https://ct.googleapis.com/logs/us1/argon2027h1/",
				BatchSize:     500,
				ParallelFetch: 10,
				BufferSize:    1000,
				Timeout:       30 * time.Second,
				MaxRetries:    3,
			},
			{
				URL:           "https://ct.googleapis.com/logs/eu1/xenon2026h1/",
				BatchSize:     500,
				ParallelFetch: 10,
				BufferSize:    1000,
				Timeout:       30 * time.Second,
				MaxRetries:    3,
			},
			{
				URL:           "https://ct.googleapis.com/logs/eu1/xenon2026h2/",
				BatchSize:     500,
				ParallelFetch: 10,
				BufferSize:    1000,
				Timeout:       30 * time.Second,
				MaxRetries:    3,
			},
			{
				URL:           "https://ct.googleapis.com/logs/eu1/xenon2027h1/",
				BatchSize:     500,
				ParallelFetch: 10,
				BufferSize:    1000,
				Timeout:       30 * time.Second,
				MaxRetries:    3,
			},
			{
				URL:           "https://ct.cloudflare.com/logs/nimbus2026/",
				BatchSize:     500,
				ParallelFetch: 10,
				BufferSize:    1000,
				Timeout:       30 * time.Second,
				MaxRetries:    3,
			},
			{
				URL:           "https://ct.cloudflare.com/logs/nimbus2027/",
				BatchSize:     500,
				ParallelFetch: 10,
				BufferSize:    1000,
				Timeout:       30 * time.Second,
				MaxRetries:    3,
			},
		},

		Domains: []string{
			"example.com",
			"google.com",
			"github.com",
			"example.test",
			"jgeai.site",
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

	sdk, err := ct.NewCTWatcherSDK(opts)
	if err != nil {
		log.Fatalf("Failed to create SDK: %v", err)
	}
	defer sdk.Close()

	sdk.OnSubdomainFound(func(event models.SubdomainEvent) {
		fmt.Printf("[Subdomain found] %s -> %s | certificate index: %d | time: %s\n",
			event.RootDomain,
			event.Subdomain,
			event.Certificate.Index,
			event.DiscoveredAt.Format("2006-01-02 15:04:05"))
	})

	sdk.OnError(func(err error, context map[string]interface{}) {
		errorType := "unknown"
		if ctxType, ok := context["error_type"]; ok {
			if et, ok := ctxType.(string); ok {
				errorType = et
			}
		}
		fmt.Printf("[Error] Type: %s | message: %v\n", errorType, err)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("Starting CT log watchers...")
	if err := sdk.Start(ctx); err != nil {
		log.Fatalf("Failed to start SDK: %v", err)
	}

	fmt.Println("CT watchers started; watching for subdomains...")
	fmt.Println("Press Ctrl+C to stop watching")

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		var lastProcessedCount int64 = 0
		var lastReportTime = time.Now()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
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

				fmt.Printf("\n=== Watcher status report ===\n")
				fmt.Printf("Runtime: %v\n", status.Uptime)
				fmt.Printf("Total certificates processed: %d\n", status.ProcessedCertificates)
				fmt.Printf("Total subdomains found: %d\n", status.DiscoveredSubdomains)
				fmt.Printf("Average QPS: %.2f certs/sec\n", avgQps)
				fmt.Printf("Recent QPS: %.2f certs/sec (last 30 seconds)\n", incrementalQps)
				fmt.Printf("Active watchers: %d\n", status.ActiveWatchers)
				fmt.Printf("Total domains: %d\n", stats.TotalDomains)

				if len(status.WatcherStatuses) > 0 {
					fmt.Printf("\n--- Watcher details ---\n")
					totalWatcherProcessed := int64(0)
					for url, watcherStatus := range status.WatcherStatuses {
						fmt.Printf("Watcher: %s\n", url)
						fmt.Printf("  Status: %s\n", watcherStatus.Status)
						fmt.Printf("  Certificates processed: %d\n", watcherStatus.ProcessedCount)
						fmt.Printf("  Last index: %d\n", watcherStatus.LastIndex)
						fmt.Printf("  Error count: %d\n", watcherStatus.ErrorCount)
						totalWatcherProcessed += watcherStatus.ProcessedCount
					}
					fmt.Printf("\nTotal watcher certificates processed: %d\n", totalWatcherProcessed)
				}

				fmt.Printf("========================\n\n")
			}
		}
	}()

	go func() {
		time.Sleep(2 * time.Minute)
		fmt.Println("Adding domain: test.com")
		if err := sdk.AddDomains([]string{"test.com"}); err != nil {
			fmt.Printf("Failed to add domain: %v\n", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\nStop signal received; stopping CT watchers...")

	cancel()

	time.Sleep(5 * time.Second)

	fmt.Println("CT watchers fully stopped")
}

// func main() {
//     runCTExample()
// }
