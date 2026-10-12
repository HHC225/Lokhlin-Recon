package template

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"recon/internal/database/mongodb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"recon/internal/models"
)

type Repository interface {
	Find(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]models.ScanTemplate, error)
	FindOne(ctx context.Context, filter bson.M, opts *options.FindOneOptions) (*models.ScanTemplate, error)
	Count(ctx context.Context, filter bson.M) (int64, error)
	InsertOne(ctx context.Context, document interface{}) (string, error)
	UpdateOne(ctx context.Context, filter bson.M, update bson.M) error
	DeleteMany(ctx context.Context, filter bson.M) error
}

type repository struct {
	collection *mongo.Collection
}

func NewRepository() Repository {
	return &repository{
		collection: mongodb.DB.Collection("ScanTemplates"),
	}
}

func (r *repository) Find(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]models.ScanTemplate, error) {
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}
	defer cursor.Close(ctx)

	var results []models.ScanTemplate
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode documents: %w", err)
	}

	return results, nil
}

func (r *repository) FindOne(ctx context.Context, filter bson.M, opts *options.FindOneOptions) (*models.ScanTemplate, error) {
	var result models.ScanTemplate
	err := r.collection.FindOne(ctx, filter, opts).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find document: %w", err)
	}

	return &result, nil
}

func (r *repository) Count(ctx context.Context, filter bson.M) (int64, error) {
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}

	return count, nil
}

func (r *repository) InsertOne(ctx context.Context, document interface{}) (string, error) {
	result, err := r.collection.InsertOne(ctx, document)
	if err != nil {
		return "", fmt.Errorf("failed to insert document: %w", err)
	}
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		return oid.Hex(), nil
	}
	return "", nil
}

func (r *repository) UpdateOne(ctx context.Context, filter bson.M, update bson.M) error {
	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	return nil
}

func (r *repository) DeleteMany(ctx context.Context, filter bson.M) error {
	_, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete documents: %w", err)
	}

	return nil
}
