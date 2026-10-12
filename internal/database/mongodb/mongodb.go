package mongodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"net/url"
	"strconv"
	"time"

	"recon/internal/logger"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"recon/internal/config"
)

var Client *mongo.Client
var DB *mongo.Database

func init() {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	username := url.QueryEscape(config.GlobalConfig.MongoDB.Username)
	password := url.QueryEscape(config.GlobalConfig.MongoDB.Password)

	uri := config.GlobalConfig.MongoDB.IP + ":" + strconv.Itoa(config.GlobalConfig.MongoDB.Port)
	if username != "" && password != "" {
		uri = "mongodb://" + username + ":" + password + "@" + uri
	}

	clientOptions := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to connect to MongoDB: %v", err))
		return
	}

	if err = client.Ping(ctx, nil); err != nil {
		logger.Error(fmt.Sprintf("failed to ping MongoDB: %v", err))
		return
	}

	Client = client
	DB = client.Database(config.GlobalConfig.MongoDB.Database)

	logger.Info("MongoDB connected successfully")
	return
}

func Close() error {
	if Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := Client.Disconnect(ctx); err != nil {
			return fmt.Errorf("failed to disconnect MongoDB: %v", err)
		}
	}
	return nil
}

func Collection(name string) *mongo.Collection {
	return DB.Collection(name)
}

func defaultContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

func InsertOne(col string, document interface{}) (*mongo.InsertOneResult, error) {
	ctx, cancel := defaultContext()
	defer cancel()

	return Collection(col).InsertOne(ctx, document)
}

func InsertMany(col string, documents []interface{}) (*mongo.InsertManyResult, error) {
	ctx, cancel := defaultContext()
	defer cancel()

	return Collection(col).InsertMany(ctx, documents)
}

func FindOne(col string, filter interface{}, result interface{}) error {
	ctx, cancel := defaultContext()
	defer cancel()

	return Collection(col).FindOne(ctx, filter).Decode(result)
}

func FindMany(col string, filter interface{}, result interface{}, opts ...*options.FindOptions) error {
	ctx, cancel := defaultContext()
	defer cancel()

	cursor, err := Collection(col).Find(ctx, filter, opts...)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	return cursor.All(ctx, result)
}

func FindAll(collectionName string, query, selector, result interface{}) error {
	collection := Collection(collectionName)
	cur, err := collection.Find(context.Background(), query, options.Find().SetProjection(selector))
	if err != nil {
		return err
	}
	defer cur.Close(context.Background())

	return cur.All(context.Background(), result)
}

func UpdateOne(col string, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	ctx, cancel := defaultContext()
	defer cancel()

	return Collection(col).UpdateOne(ctx, filter, update)
}

func UpdateMany(col string, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	ctx, cancel := defaultContext()
	defer cancel()

	return Collection(col).UpdateMany(ctx, filter, update)
}

func DeleteOne(col string, filter interface{}) (*mongo.DeleteResult, error) {
	ctx, cancel := defaultContext()
	defer cancel()

	return Collection(col).DeleteOne(ctx, filter)
}

func DeleteMany(col string, filter interface{}) (*mongo.DeleteResult, error) {
	ctx, cancel := defaultContext()
	defer cancel()

	return Collection(col).DeleteMany(ctx, filter)
}

func Count(col string, filter interface{}) (int64, error) {
	ctx, cancel := defaultContext()
	defer cancel()

	return Collection(col).CountDocuments(ctx, filter)
}

func FindPage(
	col string,
	filter interface{},
	result interface{},
	page int64,
	pageSize int64,
	sort bson.D,
) error {

	ctx, cancel := defaultContext()
	defer cancel()

	if page < 1 {
		page = 1
	}

	opts := options.Find().
		SetSkip((page - 1) * pageSize).
		SetLimit(pageSize).
		SetSort(sort)

	cursor, err := Collection(col).Find(ctx, filter, opts)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	return cursor.All(ctx, result)
}
