package root_domain

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"recon/internal/database/mongodb"
	"recon/internal/models"
)

type Repository interface {
	CountDocuments(ctx context.Context, query bson.M) (int64, error)
	FindWithPagination(ctx context.Context, query bson.M, opts *options.FindOptions) ([]models.RootDomain, error)
}

type repository struct {
	collection *mongo.Collection
}

func NewRepository() Repository {
	return &repository{
		collection: mongodb.DB.Collection("RootDomain"),
	}
}

func (r *repository) CountDocuments(ctx context.Context, query bson.M) (int64, error) {
	return r.collection.CountDocuments(ctx, query)
}

func (r *repository) FindWithPagination(ctx context.Context, query bson.M, opts *options.FindOptions) ([]models.RootDomain, error) {

	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}
	defer cursor.Close(ctx)

	var results []models.RootDomain
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode documents: %w", err)
	}

	return results, nil
}
