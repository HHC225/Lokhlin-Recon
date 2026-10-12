package plugin

import (
	"context"
	"fmt"
	"strings"

	redisdriver "github.com/redis/go-redis/v9"
	"recon/internal/database/mongodb"
	"recon/internal/database/redis"
	"recon/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository interface {
	FindWithPagination(ctx context.Context, query bson.M, opts *options.FindOptions) ([]models.Plugin, error)
	FindByModule(ctx context.Context, module string) ([]models.Plugin, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Plugin, error)
	FindByHash(ctx context.Context, hash string) (*models.Plugin, error)
	Create(ctx context.Context, plugin *models.Plugin) (string, error)
	Update(ctx context.Context, id primitive.ObjectID, update bson.M) error
	UpdateByHash(ctx context.Context, hash string, update bson.M) error
	UpsertByHash(ctx context.Context, plugin *models.Plugin) (string, error)
	DeleteByHash(ctx context.Context, hashes []string) error
	Count(ctx context.Context, query bson.M) (int64, error)
	GetLogs(ctx context.Context, key string) (string, error)
	CleanLogs(ctx context.Context, key string) error
	CleanAllLogs(ctx context.Context) error
}

type repository struct {
	collection  *mongo.Collection
	redisClient *redisdriver.Client
}

func NewRepository() Repository {
	return &repository{
		collection:  mongodb.DB.Collection("plugins"),
		redisClient: redis.Client,
	}
}

func (r *repository) FindWithPagination(ctx context.Context, query bson.M, opts *options.FindOptions) ([]models.Plugin, error) {
	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}
	defer cursor.Close(ctx)

	var results []models.Plugin
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode documents: %w", err)
	}

	return results, nil
}

func (r *repository) FindByModule(ctx context.Context, module string) ([]models.Plugin, error) {
	query := bson.M{"module": module}

	opts := options.Find().SetProjection(bson.M{
		"source": 0,
	})

	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}
	defer cursor.Close(ctx)

	var results []models.Plugin
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode documents: %w", err)
	}

	return results, nil
}

func (r *repository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Plugin, error) {
	var result models.Plugin
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find document: %w", err)
	}
	return &result, nil
}

func (r *repository) FindByHash(ctx context.Context, hash string) (*models.Plugin, error) {
	var result models.Plugin
	err := r.collection.FindOne(ctx, bson.M{"hash": hash}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find document: %w", err)
	}
	return &result, nil
}

func (r *repository) Create(ctx context.Context, plugin *models.Plugin) (string, error) {
	result, err := r.collection.InsertOne(ctx, plugin)
	if err != nil {
		return "", fmt.Errorf("failed to insert document: %w", err)
	}

	objectID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", fmt.Errorf("inserted ID is not an ObjectID")
	}
	return objectID.Hex(), err
}

func (r *repository) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}
	return nil
}

func (r *repository) UpdateByHash(ctx context.Context, hash string, update bson.M) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"hash": hash}, bson.M{"$set": update})
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}
	return nil
}

func (r *repository) UpsertByHash(ctx context.Context, plugin *models.Plugin) (string, error) {
	opts := options.FindOneAndReplace().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var result models.Plugin
	err := r.collection.FindOneAndReplace(ctx, bson.M{"hash": plugin.Hash}, plugin, opts).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("failed to upsert document: %w", err)
	}

	return result.ID.Hex(), nil
}

func (r *repository) DeleteByHash(ctx context.Context, hashes []string) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"hash": bson.M{"$in": hashes}})
	if err != nil {
		return fmt.Errorf("failed to delete documents: %w", err)
	}
	return nil
}

func (r *repository) Count(ctx context.Context, query bson.M) (int64, error) {
	count, err := r.collection.CountDocuments(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}
	return count, nil
}

func (r *repository) GetLogs(ctx context.Context, key string) (string, error) {

	logs, err := r.redisClient.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get logs from Redis: %w", err)
	}

	logData := strings.Join(logs, "\n")

	return logData, nil
}

func (r *repository) CleanLogs(ctx context.Context, key string) error {

	err := r.redisClient.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to get logs from Redis: %w", err)
	}
	return nil
}

func (r *repository) CleanAllLogs(ctx context.Context) error {

	keys, err := r.redisClient.Keys(ctx, "logs:plugins:*").Result()
	if err != nil {
		return fmt.Errorf("failed to get log keys from Redis: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	err = r.redisClient.Del(ctx, keys...).Err()
	if err != nil {
		return fmt.Errorf("failed to delete logs from Redis: %w", err)
	}
	return nil
}
