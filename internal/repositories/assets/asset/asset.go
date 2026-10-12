package asset

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"recon/internal/database/mongodb"
	"recon/internal/models"
)

type Repository interface {
	GetAssetByID(ctx context.Context, id string) (*models.Asset, error)
	GetAssetByIDs(ctx context.Context, ids []string) ([]models.Asset, error)
	CreateAsset(ctx context.Context, asset *models.Asset) error
	UpdateAsset(ctx context.Context, id string, asset *models.Asset) error
	DeleteAsset(ctx context.Context, id string) error
	GetScreenshot(ctx context.Context, id string) (string, error)
	GetChangeLog(ctx context.Context, id string) ([]models.AssetChangeLog, error)
	DeduplicateAssets(ctx context.Context, assetType string, filter bson.M, groupFields []string) error
	CountDocuments(ctx context.Context, filter bson.M) (int64, error)
	FindWithOptions(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]models.Asset, error)
	FindBodyByHash(ctx context.Context, hash string) (models.HttpBody, error)
	FindBodyWithOptions(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]models.HttpBody, error)
	GetDistinctHostsByRootDomain(ctx context.Context, rootDomain string, skip, limit int64) ([]string, int64, error)
	GetAssetsByHost(ctx context.Context, host string, skip, limit int64) ([]models.Asset, int64, error)
}

type repository struct {
	collection *mongo.Collection
	bodyColl   *mongo.Collection
}

func NewRepository() Repository {
	return &repository{
		collection: mongodb.DB.Collection("asset"),
		bodyColl:   mongodb.DB.Collection("HttpBody"),
	}
}

func (r *repository) GetAssetByID(ctx context.Context, id string) (*models.Asset, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var asset models.Asset
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&asset)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &asset, nil
}

func (r *repository) GetAssetByIDs(ctx context.Context, ids []string) ([]models.Asset, error) {
	if len(ids) == 0 {
		return []models.Asset{}, nil
	}

	var objectIDs []primitive.ObjectID
	for _, id := range ids {
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, err
		}
		objectIDs = append(objectIDs, objectID)
	}

	filter := bson.M{"_id": bson.M{"$in": objectIDs}}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var assets []models.Asset
	if err := cursor.All(ctx, &assets); err != nil {
		return nil, err
	}

	return assets, nil
}

func (r *repository) CreateAsset(ctx context.Context, asset *models.Asset) error {
	asset.ID = primitive.NewObjectID()

	result, err := r.collection.InsertOne(ctx, asset)
	if err != nil {
		return err
	}

	asset.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *repository) UpdateAsset(ctx context.Context, id string, asset *models.Asset) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{"$set": asset}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

func (r *repository) DeleteAsset(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}

func (r *repository) GetScreenshot(ctx context.Context, id string) (string, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return "", err
	}

	var asset models.Asset
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&asset)
	if err != nil {
		return "", err
	}

	return asset.Screenshot, nil
}

func (r *repository) GetChangeLog(ctx context.Context, id string) ([]models.AssetChangeLog, error) {
	var logs []models.AssetChangeLog

	filter := bson.M{"assetid": id}
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}})

	cursor, err := r.collection.Database().Collection("AssetChangeLog").Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &logs); err != nil {
		return nil, err
	}

	return logs, nil
}

func (r *repository) DeduplicateAssets(ctx context.Context, assetType string, filter bson.M, groupFields []string) error {

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"group": "$" + groupFields[0],
			},
			"count": bson.M{"$sum": 1},
			"docs":  bson.M{"$push": "$$ROOT"},
		}}},
		{{Key: "$match", Value: bson.M{"count": bson.M{"$gt": 1}}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return err
	}

	for _, result := range results {
		docs := result["docs"].([]interface{})

		for i := 1; i < len(docs); i++ {
			doc := docs[i].(bson.M)
			_, err := r.collection.DeleteOne(ctx, bson.M{"_id": doc["_id"]})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *repository) CountDocuments(ctx context.Context, filter bson.M) (int64, error) {
	return r.collection.CountDocuments(ctx, filter)
}

func (r *repository) FindWithOptions(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]models.Asset, error) {
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var assets []models.Asset
	if err := cursor.All(ctx, &assets); err != nil {
		return nil, err
	}

	return assets, nil
}

func (r *repository) FindBodyByHash(ctx context.Context, hash string) (models.HttpBody, error) {
	filter := bson.M{"hash": hash}

	var httpBody models.HttpBody
	err := r.bodyColl.FindOne(ctx, filter).Decode(&httpBody)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return models.HttpBody{}, nil
		}
		return models.HttpBody{}, err
	}
	return httpBody, nil
}

func (r *repository) FindBodyWithOptions(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]models.HttpBody, error) {
	cursor, err := r.bodyColl.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var assets []models.HttpBody
	if err := cursor.All(ctx, &assets); err != nil {
		return nil, err
	}

	return assets, nil
}

func (r *repository) GetDistinctHostsByRootDomain(ctx context.Context, rootDomain string, skip, limit int64) ([]string, int64, error) {

	basePipeline := mongo.Pipeline{

		{{Key: "$match", Value: bson.M{"rootDomain": rootDomain, "host": bson.M{"$ne": ""}}}},

		{{Key: "$sort", Value: bson.D{{Key: "time", Value: -1}}}},

		{{Key: "$group", Value: bson.M{
			"_id":  "$host",
			"host": bson.M{"$first": "$host"},
		}}},

		{{Key: "$sort", Value: bson.D{{Key: "host", Value: 1}}}},
	}

	pipeline := make(mongo.Pipeline, len(basePipeline))
	copy(pipeline, basePipeline)
	pipeline = append(pipeline, primitive.D{
		{Key: "$facet", Value: bson.M{

			"total": mongo.Pipeline{
				{{Key: "$count", Value: "count"}},
			},

			"data": mongo.Pipeline{
				{{Key: "$skip", Value: skip}},
				{{Key: "$limit", Value: limit}},
				{{Key: "$project", Value: bson.M{"host": 1, "_id": 0}}},
			},
		}},
	})

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}

	if len(results) == 0 {
		return []string{}, 0, nil
	}

	facetResult := results[0]

	var total int64
	if totalArr, ok := facetResult["total"].(bson.A); ok && len(totalArr) > 0 {
		if totalDoc, ok := totalArr[0].(bson.M); ok {
			if countVal, ok := totalDoc["count"].(int32); ok {
				total = int64(countVal)
			} else if countVal, ok := totalDoc["count"].(int64); ok {
				total = countVal
			}
		}
	}

	var hosts []string
	if dataArr, ok := facetResult["data"].(bson.A); ok {
		for _, item := range dataArr {
			if doc, ok := item.(bson.M); ok {
				if host, ok := doc["host"].(string); ok && host != "" {
					hosts = append(hosts, host)
				}
			}
		}
	}

	return hosts, total, nil
}

func (r *repository) GetAssetsByHost(ctx context.Context, host string, skip, limit int64) ([]models.Asset, int64, error) {

	filter := bson.M{"host": host}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	projection := bson.M{
		"ip":          1,
		"port":        1,
		"service":     1,
		"statuscode":  1,
		"faviconmmh3": 1,
		"bodyhash":    1,
		"title":       1,
		"url":         1,
	}

	opts := options.Find().
		SetProjection(projection).
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "time", Value: -1}})

	assets, err := r.FindWithOptions(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}

	return assets, total, nil
}
