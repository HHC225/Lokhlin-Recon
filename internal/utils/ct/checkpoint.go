// ct-----------------------------------------
// @file      : checkpoint.go
// @time      : 2025/01/24 21:00
// -------------------------------------------

// ============================================

package ct

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	sasRedis "recon/internal/database/redis"
	"recon/internal/utils/ct/models"
)

type RedisCheckpointManager struct {
	client    *redis.Client
	keyPrefix string
	ctx       context.Context
}

func NewRedisCheckpointManager(keyPrefix string) (*RedisCheckpointManager, error) {

	redisClient := sasRedis.Client
	if redisClient == nil {
		return nil, fmt.Errorf("Redis client is not initialized")
	}

	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Redis connection failed: %w", err)
	}

	return &RedisCheckpointManager{
		client:    redisClient,
		keyPrefix: keyPrefix,
		ctx:       ctx,
	}, nil
}

func (rcm *RedisCheckpointManager) SaveCheckpoint(cp *models.Checkpoint) error {
	if cp == nil || !cp.IsValid() {
		return fmt.Errorf("Invalid checkpoint data")
	}

	key := rcm.makeKey(cp.URL)
	data, err := json.Marshal(cp)
	if err != nil {
		return fmt.Errorf("Failed to serialize checkpoint: %w", err)
	}

	err = rcm.client.Set(rcm.ctx, key, data, 0).Err()
	if err != nil {
		return fmt.Errorf("Failed to save checkpoint to Redis: %w", err)
	}

	return nil
}

func (rcm *RedisCheckpointManager) LoadCheckpoint(url string) (*models.Checkpoint, error) {
	key := rcm.makeKey(url)

	data, err := rcm.client.Get(rcm.ctx, key).Result()
	if err != nil {
		if err.Error() == "redis: nil" {

			return &models.Checkpoint{
				URL:                url,
				LastProcessedIndex: 0,
				Timestamp:          time.Now(),
				Version:            1,
			}, nil
		}
		return nil, fmt.Errorf("Failed to load checkpoint from Redis: %w", err)
	}

	var cp models.Checkpoint
	if err := json.Unmarshal([]byte(data), &cp); err != nil {
		return nil, fmt.Errorf("Failed to deserialize checkpoint: %w", err)
	}

	return &cp, nil
}

func (rcm *RedisCheckpointManager) DeleteCheckpoint(url string) error {
	key := rcm.makeKey(url)

	err := rcm.client.Del(rcm.ctx, key).Err()
	if err != nil {
		return fmt.Errorf("Failed to delete checkpoint: %w", err)
	}

	return nil
}

func (rcm *RedisCheckpointManager) ListCheckpoints() ([]*models.Checkpoint, error) {
	pattern := rcm.keyPrefix + "checkpoint:*"

	keys, err := rcm.client.Keys(rcm.ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("Failed to list checkpoint keys: %w", err)
	}

	checkpoints := make([]*models.Checkpoint, 0, len(keys))

	for _, key := range keys {
		data, err := rcm.client.Get(rcm.ctx, key).Result()
		if err != nil {
			continue
		}

		var cp models.Checkpoint
		if err := json.Unmarshal([]byte(data), &cp); err != nil {
			continue
		}

		checkpoints = append(checkpoints, &cp)
	}

	return checkpoints, nil
}

func (rcm *RedisCheckpointManager) CleanupExpired(olderThan time.Duration) error {
	checkpoints, err := rcm.ListCheckpoints()
	if err != nil {
		return fmt.Errorf("Failed to list checkpoints: %w", err)
	}

	cutoffTime := time.Now().Add(-olderThan)
	deletedCount := 0

	for _, cp := range checkpoints {
		if cp.Timestamp.Before(cutoffTime) {
			if err := rcm.DeleteCheckpoint(cp.URL); err != nil {
				continue
			}
			deletedCount++
		}
	}

	return nil
}

func (rcm *RedisCheckpointManager) BatchSaveCheckpoints(checkpoints []*models.Checkpoint) error {
	if len(checkpoints) == 0 {
		return nil
	}

	pipe := rcm.client.Pipeline()

	for _, cp := range checkpoints {
		if cp == nil || !cp.IsValid() {
			continue
		}

		key := rcm.makeKey(cp.URL)
		data, err := json.Marshal(cp)
		if err != nil {
			continue
		}

		pipe.Set(rcm.ctx, key, data, 0)
	}

	_, err := pipe.Exec(rcm.ctx)
	if err != nil {
		return fmt.Errorf("Failed to save checkpoint batch: %w", err)
	}

	return nil
}

func (rcm *RedisCheckpointManager) GetCheckpointStats() (*CheckpointStats, error) {
	checkpoints, err := rcm.ListCheckpoints()
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve checkpoint statistics: %w", err)
	}

	stats := &CheckpointStats{
		TotalCheckpoints: len(checkpoints),
	}

	if len(checkpoints) > 0 {
		oldest := checkpoints[0].Timestamp
		newest := checkpoints[0].Timestamp

		totalIndexes := int64(0)

		for _, cp := range checkpoints {
			if cp.Timestamp.Before(oldest) {
				oldest = cp.Timestamp
			}
			if cp.Timestamp.After(newest) {
				newest = cp.Timestamp
			}
			totalIndexes += cp.LastProcessedIndex
		}

		stats.OldestCheckpoint = oldest
		stats.NewestCheckpoint = newest
		stats.AverageIndex = totalIndexes / int64(len(checkpoints))
	}

	return stats, nil
}

func (rcm *RedisCheckpointManager) Close() error {
	return nil
}

func (rcm *RedisCheckpointManager) makeKey(url string) string {
	return rcm.keyPrefix + "checkpoint:" + url
}

type CheckpointStats struct {
	TotalCheckpoints int       `json:"total_checkpoints"`
	OldestCheckpoint time.Time `json:"oldest_checkpoint"`
	NewestCheckpoint time.Time `json:"newest_checkpoint"`
	AverageIndex     int64     `json:"average_index"`
}

type AutoCheckpointSaver struct {
	manager     *RedisCheckpointManager
	interval    time.Duration
	checkpoints map[string]*models.Checkpoint
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	isRunning   bool
}

func NewAutoCheckpointSaver(manager *RedisCheckpointManager, interval time.Duration) *AutoCheckpointSaver {
	ctx, cancel := context.WithCancel(context.Background())

	return &AutoCheckpointSaver{
		manager:     manager,
		interval:    interval,
		checkpoints: make(map[string]*models.Checkpoint),
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (acs *AutoCheckpointSaver) Start() error {
	if acs.isRunning {
		return fmt.Errorf("Autosaver is already running")
	}

	acs.isRunning = true
	acs.saveAllCheckpoints()
	go func() {
		ticker := time.NewTicker(acs.interval)
		defer ticker.Stop()

		for {
			select {
			case <-acs.ctx.Done():

				acs.saveAllCheckpoints()
				return
			case <-ticker.C:

				acs.saveAllCheckpoints()
			}
		}
	}()

	return nil
}

func (acs *AutoCheckpointSaver) Stop() error {
	if !acs.isRunning {
		return nil
	}

	acs.cancel()
	acs.isRunning = false

	return nil
}

func (acs *AutoCheckpointSaver) UpdateCheckpoint(cp *models.Checkpoint) {
	acs.mu.Lock()
	defer acs.mu.Unlock()

	if cp != nil && cp.IsValid() {
		cp.Timestamp = time.Now()
		acs.checkpoints[cp.URL] = cp
	}
}

func (acs *AutoCheckpointSaver) RemoveCheckpoint(url string) {
	acs.mu.Lock()
	defer acs.mu.Unlock()

	delete(acs.checkpoints, url)
}

func (acs *AutoCheckpointSaver) saveAllCheckpoints() {
	acs.mu.RLock()
	checkpoints := make([]*models.Checkpoint, 0, len(acs.checkpoints))
	for _, cp := range acs.checkpoints {
		checkpoints = append(checkpoints, cp)
	}
	acs.mu.RUnlock()

	if len(checkpoints) > 0 {

		if err := acs.manager.BatchSaveCheckpoints(checkpoints); err != nil {

		}
	}
}

func (acs *AutoCheckpointSaver) GetPendingCheckpoints() int {
	acs.mu.RLock()
	defer acs.mu.RUnlock()
	return len(acs.checkpoints)
}
