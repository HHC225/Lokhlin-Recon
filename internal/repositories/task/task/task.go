// Package task -----------------------------
// @file      : task.go
// @time      : 2025/5/4 22:04
// -------------------------------------------
package task

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	redisdriver "github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"recon/internal/database/mongodb"
	"recon/internal/database/redis"
	"recon/internal/models"
)

type Repository interface {
	Count(ctx *gin.Context, filter interface{}) (int64, error)
	Find(ctx context.Context, filter interface{}, opts *options.FindOptions) ([]models.Task, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Task, error)
	Insert(ctx context.Context, task *models.Task) (string, error)
	ClearTaskCache(ctx context.Context, id string) error
	PushTaskInfoList(ctx context.Context, id string, targetList []string) error
	FindTemplateByID(ctx context.Context, id string) (*models.ScanTemplate, error)
	GetCollectionCursor(ctx *gin.Context, collection string, filter bson.M) (*mongo.Cursor, error)
	RPushNodeTask(ctx context.Context, name string, data any) error
	CheckTaskNameExists(ctx context.Context, name string) (bool, error)
	DeleteMany(ctx *gin.Context, ids []primitive.ObjectID) (int64, error)
	UpdateOne(ctx context.Context, filter bson.M, update bson.M) error
	GetProgressFromRedis(ctx *gin.Context, key string) (map[string]string, error)

	Exists(ctx context.Context, key string) (bool, error)
	SCard(ctx context.Context, key string) (int64, error)
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) error
}

type repository struct {
	db          *mongo.Database
	redisClient *redisdriver.Client
}

func NewRepository() Repository {
	return &repository{
		db:          mongodb.DB,
		redisClient: redis.Client,
	}
}

func (r *repository) Count(ctx *gin.Context, filter interface{}) (int64, error) {
	return r.db.Collection("task").CountDocuments(ctx.Request.Context(), filter)
}

func (r *repository) Find(ctx context.Context, filter interface{}, opts *options.FindOptions) ([]models.Task, error) {
	cursor, err := r.db.Collection("task").Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tasks []models.Task
	if err = cursor.All(ctx, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *repository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Task, error) {
	var task models.Task
	err := r.db.Collection("task").FindOne(ctx, bson.M{"_id": id}).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find task by ID: %w", err)
	}
	return &task, nil
}

func (r *repository) Insert(ctx context.Context, task *models.Task) (string, error) {
	result, err := r.db.Collection("task").InsertOne(ctx, task)
	if err != nil {
		return "", err
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		return oid.Hex(), nil
	}
	return "", fmt.Errorf("failed to convert InsertedID to ObjectID")
}

func (r *repository) ClearTaskCache(ctx context.Context, id string) error {
	keysToDelete := []string{
		fmt.Sprintf("TaskInfo:tmp:%s", id),
		fmt.Sprintf("TaskInfo:%s", id),
		fmt.Sprintf("TaskInfo:time:%s", id),
	}

	progressKeys, err := r.redisClient.Keys(ctx, fmt.Sprintf("TaskInfo:progress:%s:*", id)).Result()
	if err != nil {
		return fmt.Errorf("failed to get progress keys: %w", err)
	}
	keysToDelete = append(keysToDelete, progressKeys...)

	duplicateKeys, err := r.redisClient.Keys(ctx, fmt.Sprintf("duplicates:%s:*", id)).Result()
	if err != nil {
		return fmt.Errorf("failed to get duplicate keys: %w", err)
	}
	keysToDelete = append(keysToDelete, duplicateKeys...)

	if len(keysToDelete) > 0 {
		if err := r.redisClient.Del(ctx, keysToDelete...).Err(); err != nil {
			return fmt.Errorf("failed to delete redis keys: %w", err)
		}
	}
	return nil
}

func (r *repository) PushTaskInfoList(ctx context.Context, id string, targetList []string) error {
	if len(targetList) == 0 {
		return nil
	}

	values := make([]interface{}, len(targetList))
	for i, v := range targetList {
		values[i] = v
	}

	if err := r.redisClient.LPush(ctx, fmt.Sprintf("TaskInfo:%s", id), values...).Err(); err != nil {
		return fmt.Errorf("failed to lpush TaskInfo list: %w", err)
	}

	return nil
}

func (r *repository) FindTemplateByID(ctx context.Context, id string) (*models.ScanTemplate, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ObjectID: %w", err)
	}

	var result models.ScanTemplate
	err = r.db.Collection("ScanTemplates").FindOne(ctx, bson.M{"_id": objectID}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find template: %w", err)
	}

	return &result, nil
}

func (r *repository) GetCollectionCursor(ctx *gin.Context, collection string, filter bson.M) (*mongo.Cursor, error) {
	return r.db.Collection(collection).Find(ctx.Request.Context(), filter)
}

func (r *repository) RPushNodeTask(ctx context.Context, name string, data any) error {
	key := fmt.Sprintf("NodeTask:%s", name)

	value, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return r.redisClient.RPush(ctx, key, value).Err()
}

func (r *repository) CheckTaskNameExists(ctx context.Context, name string) (bool, error) {
	filter := bson.M{"name": name}
	count, err := r.db.Collection("task").CountDocuments(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check task name: %w", err)
	}
	return count > 0, nil
}

func (r *repository) DeleteMany(ctx *gin.Context, ids []primitive.ObjectID) (int64, error) {
	result, err := r.db.Collection("task").DeleteMany(ctx.Request.Context(), bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

func (r *repository) UpdateOne(ctx context.Context, filter bson.M, update bson.M) error {
	_, err := r.db.Collection("task").UpdateOne(ctx, filter, update)
	return err
}

func (r *repository) GetProgressFromRedis(ctx *gin.Context, key string) (map[string]string, error) {
	result, err := r.redisClient.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get progress from redis: %w", err)
	}
	return result, nil
}

func (r *repository) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.redisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}
	return result > 0, nil
}

func (r *repository) SCard(ctx context.Context, key string) (int64, error) {
	result, err := r.redisClient.SCard(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get set cardinality: %w", err)
	}
	return result, nil
}

func (r *repository) Get(ctx context.Context, key string) (string, error) {
	result, err := r.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redisdriver.Nil {
			return "", nil
		}
		return "", fmt.Errorf("failed to get key value: %w", err)
	}
	return result, nil
}

func (r *repository) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	_, err := r.redisClient.Del(ctx, keys...).Result()
	if err != nil {
		return fmt.Errorf("failed to delete keys: %w", err)
	}
	return nil
}
