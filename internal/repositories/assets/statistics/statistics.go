package statistics

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo/options"
	"recon/internal/database/mongodb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Repository interface {
	CountDocuments(ctx context.Context, collection string, filter bson.M) (int64, error)
	Aggregate(ctx context.Context, collection string, pipeline []bson.M) (*mongo.Cursor, error)
	Find(ctx context.Context, collection string, filter bson.M, projection bson.M) (*mongo.Cursor, error)
}

type repository struct {
	db *mongo.Database
}

func NewRepository() Repository {
	return &repository{
		db: mongodb.DB,
	}
}

func (r *repository) CountDocuments(ctx context.Context, collection string, filter bson.M) (int64, error) {
	return r.db.Collection(collection).CountDocuments(ctx, filter)
}

func (r *repository) Aggregate(ctx context.Context, collection string, pipeline []bson.M) (*mongo.Cursor, error) {
	cursor, err := r.db.Collection(collection).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	return cursor, nil
}

func (r *repository) Find(ctx context.Context, collection string, filter bson.M, projection bson.M) (*mongo.Cursor, error) {
	cursor, err := r.db.Collection(collection).Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, err
	}
	return cursor, nil
}
