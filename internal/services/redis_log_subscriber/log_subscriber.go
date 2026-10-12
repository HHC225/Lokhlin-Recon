// Package redis_log_subscriber -----------------------------
// @file      : log_subscriber.go
// @time      : 2025/01/XX XX:XX
// -------------------------------------------
package redis_log_subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"recon/internal/services/project"
	"recon/internal/services/task/task"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"recon/internal/config"
	"recon/internal/database/mongodb"
	redisdb "recon/internal/database/redis"
	"recon/internal/logger"
	"recon/internal/models"
	"recon/internal/services/task/common"
)

var (
	GET_LOG_NAME []string
	logNameMutex sync.RWMutex

	LOG_INFO     map[string][]string
	logInfoMutex sync.RWMutex
)
var taskService task.Service
var projectService project.Service

func init() {
	LOG_INFO = make(map[string][]string)
	GET_LOG_NAME = make([]string, 0)
	taskService = task.NewService()
	projectService = project.NewService()
}

func SubscribeLogChannel() {
	channelName := "logs"
	logger.Info("Subscribed to channel", zap.String("channel", channelName))

	for {
		trySubscribe(channelName)

		time.Sleep(2 * time.Second)
	}
}

func trySubscribe(channelName string) {
	ctx := context.Background()
	pubsub := redisdb.Client.Subscribe(ctx, channelName)
	defer pubsub.Close()

	_, err := pubsub.Receive(ctx)
	if err != nil {
		logger.Error("Failed to subscribe to channel", zap.String("channel", channelName), zap.Error(err))
		return
	}

	ch := pubsub.Channel()
	logger.Info("Successfully subscribed to channel", zap.String("channel", channelName))

	for msg := range ch {
		if err := handleMessage(ctx, msg.Payload); err != nil {
			logger.Error("Failed to handle message", zap.Error(err))
			continue
		}
	}
}

var (
	logCounters   = make(map[string]int)
	logCountersMu sync.Mutex
)

func handleMessage(ctx context.Context, payload string) error {

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	logName, ok := data["name"].(string)
	if !ok {
		return fmt.Errorf("invalid log name in message")
	}

	logContent, ok := data["log"].(string)
	if !ok {
		return fmt.Errorf("invalid log content in message")
	}

	//logNameMutex.RLock()
	//inList := contains(GET_LOG_NAME, logName)
	//logNameMutex.RUnlock()
	//
	//if inList {
	//	logInfoMutex.Lock()
	//	if LOG_INFO[logName] == nil {
	//		LOG_INFO[logName] = make([]string, 0)
	//	}
	//	LOG_INFO[logName] = append(LOG_INFO[logName], logContent)
	//	logInfoMutex.Unlock()
	//}
	logger.Info(fmt.Sprintf("%s %s", logName, logContent))

	if strings.Contains(logContent, "Register Success") {
		if err := checkNodeTask(ctx, logName); err != nil {
			logger.Error("Failed to check node task", zap.String("node", logName), zap.Error(err))
		}
	}
	if strings.Contains(logContent, "Task end") {
		go func() {
			err := taskService.TaskProgress(context.Background())
			if err != nil {
				return
			}

			if err := projectService.RefreshProjectAssociations(context.Background()); err != nil {
				logger.Error("Failed to refresh project asset associations", zap.Error(err))
			}
		}()
	}

	logKey := fmt.Sprintf("log:%s", logName)
	if err := redisdb.Client.RPush(ctx, logKey, logContent).Err(); err != nil {
		return fmt.Errorf("failed to push log to Redis: %w", err)
	}

	logCountersMu.Lock()
	defer logCountersMu.Unlock()

	logCounters[logName]++
	if logCounters[logName] > config.GlobalConfig.Log.TotalLogs {

		if err := redisdb.Client.Del(ctx, logKey).Err(); err != nil {
			logger.Error("Failed to delete log key", zap.String("key", logKey), zap.Error(err))
		}
		logCounters[logName] = 0
	}

	return nil
}

func checkNodeTask(ctx context.Context, nodeName string) error {

	filter := bson.M{
		"progress": bson.M{"$ne": 100},
		"status":   1,
		"$or": []bson.M{
			{"node": nodeName},
			{"allNode": true},
		},
	}

	opts := options.Find()
	cursor, err := mongodb.DB.Collection("task").Find(ctx, filter, opts)
	if err != nil {
		return fmt.Errorf("failed to find tasks: %w", err)
	}
	defer cursor.Close(ctx)

	var tasks []models.Task
	if err := cursor.All(ctx, &tasks); err != nil {
		return fmt.Errorf("failed to decode tasks: %w", err)
	}

	if len(tasks) == 0 {
		return nil
	}

	taskService := common.NewService()

	for _, taskDoc := range tasks {

		scanTemplate, err := taskService.GetScanTemplate(context.Background(), &taskDoc)
		if err != nil {
			logger.Error("Failed to get scan template", zap.String("task_id", taskDoc.ID.Hex()), zap.Error(err))
			continue
		}

		taskDataJSON, err := json.Marshal(scanTemplate)
		if err != nil {
			logger.Error("Failed to marshal task data", zap.String("task_id", taskDoc.ID.Hex()), zap.Error(err))
			continue
		}

		nodeTaskKey := fmt.Sprintf("NodeTask:%s", nodeName)
		if err := redisdb.Client.RPush(ctx, nodeTaskKey, taskDataJSON).Err(); err != nil {
			logger.Error("Failed to push node task", zap.String("node", nodeName), zap.Error(err))
			continue
		}
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func AddLogName(name string) {
	logNameMutex.Lock()
	defer logNameMutex.Unlock()

	if !contains(GET_LOG_NAME, name) {
		GET_LOG_NAME = append(GET_LOG_NAME, name)
	}
}

func RemoveLogName(name string) {
	logNameMutex.Lock()
	defer logNameMutex.Unlock()

	for i, n := range GET_LOG_NAME {
		if n == name {
			GET_LOG_NAME = append(GET_LOG_NAME[:i], GET_LOG_NAME[i+1:]...)
			break
		}
	}
}

func GetLogInfo(name string) []string {
	logInfoMutex.RLock()
	defer logInfoMutex.RUnlock()

	logs := LOG_INFO[name]
	if logs == nil {
		return []string{}
	}
	return logs
}

func PopLogInfo(name string) string {
	logInfoMutex.Lock()
	defer logInfoMutex.Unlock()

	logs := LOG_INFO[name]
	if logs == nil || len(logs) == 0 {
		return ""
	}

	log := logs[0]
	LOG_INFO[name] = logs[1:]
	return log
}
