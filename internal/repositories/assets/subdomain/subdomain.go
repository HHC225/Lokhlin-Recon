package subdomain

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"recon/internal/database/mongodb"
	"recon/internal/models"
)

type Repository interface {
	BuildSearchQuery(query map[string]interface{}) (bson.M, error)
	CountDocuments(ctx context.Context, query bson.M) (int64, error)
	FindWithPagination(ctx context.Context, query bson.M, opts *options.FindOptions) ([]models.Subdomain, error)
}

type repository struct {
	collection *mongo.Collection
}

func NewRepository() Repository {
	return &repository{
		collection: mongodb.DB.Collection("subdomain"),
	}
}

func (r *repository) BuildSearchQuery(query map[string]interface{}) (bson.M, error) {
	if len(query) == 0 {
		return bson.M{"_id": bson.M{"$exists": true}}, nil
	}

	searchQuery := bson.M{}
	for key, value := range query {
		searchQuery[key] = value
	}

	return searchQuery, nil
}

func (r *repository) CountDocuments(ctx context.Context, query bson.M) (int64, error) {
	return r.collection.CountDocuments(ctx, query)
}

func (r *repository) FindWithPagination(ctx context.Context, query bson.M, opts *options.FindOptions) ([]models.Subdomain, error) {

	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}
	defer cursor.Close(ctx)

	var results []models.Subdomain
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode documents: %w", err)
	}

	return results, nil
}
