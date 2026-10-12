package url

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"recon/internal/database/mongodb"
	"recon/internal/models"
)

type Repository interface {
	Find(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]models.URL, error)
}

type repository struct {
	collection *mongo.Collection
}

func NewRepository() Repository {
	return &repository{
		collection: mongodb.DB.Collection("UrlScan"),
	}
}

func (r *repository) Find(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]models.URL, error) {
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var urls []models.URL
	if err := cursor.All(ctx, &urls); err != nil {
		return nil, err
	}

	return urls, nil
}
