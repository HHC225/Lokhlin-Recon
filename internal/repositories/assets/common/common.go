package common

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo/options"
	"recon/internal/database/mongodb"
	"recon/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Repository interface {
	DeleteMany(ctx context.Context, collection string, filter bson.M) (int64, error)
	FindOne(ctx context.Context, collection string, filter bson.M) (*models.Document, error)
	Find(ctx context.Context, collection string, filter bson.M, limit int64, fields []string, result interface{}) error
	UpdateOne(ctx context.Context, collection string, filter bson.M, update bson.M) error
	CountDocuments(ctx context.Context, collection string, filter bson.M) (int64, error)
	EstimatedDocumentCount(ctx context.Context, collection string) (int64, error)
	DeleteManyByFilter(ctx context.Context, collection string, filter bson.M) (int64, error)
}

type repository struct {
	db *mongo.Database
}

func NewRepository() Repository {
	return &repository{
		db: mongodb.DB,
	}
}

func (r *repository) Find(ctx context.Context, collection string, filter bson.M, limit int64, fields []string, result interface{}) error {
	findOpts := options.Find()
	if limit > 0 {
		findOpts.SetLimit(limit)
	}
	if len(fields) > 0 {
		projection := bson.M{}
		for _, f := range fields {
			projection[f] = 1
		}
		findOpts.SetProjection(projection)
	}
	cur, err := r.db.Collection(collection).Find(ctx, filter, findOpts)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return cur.All(ctx, result)
}

func (r *repository) DeleteMany(ctx context.Context, collection string, filter bson.M) (int64, error) {
	result, err := r.db.Collection(collection).DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

func (r *repository) FindOne(ctx context.Context, collection string, filter bson.M) (*models.Document, error) {
	var doc models.Document
	err := r.db.Collection(collection).FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *repository) UpdateOne(ctx context.Context, collection string, filter bson.M, update bson.M) error {
	_, err := r.db.Collection(collection).UpdateOne(ctx, filter, update)
	return err
}

func (r *repository) CountDocuments(ctx context.Context, collection string, filter bson.M) (int64, error) {
	return r.db.Collection(collection).CountDocuments(ctx, filter)
}

func (r *repository) EstimatedDocumentCount(ctx context.Context, collection string) (int64, error) {
	return r.db.Collection(collection).EstimatedDocumentCount(ctx)
}

func (r *repository) DeleteManyByFilter(ctx context.Context, collection string, filter bson.M) (int64, error) {
	result, err := r.db.Collection(collection).DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}
