package export

import (
	"context"
	"fmt"
	"time"

	"recon/internal/utils/helper"

	"recon/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"recon/internal/database/mongodb"
)

type Repository interface {
	CreateExportTask(ctx context.Context, task *models.ExportTask) error

	UpdateExportTaskStatus(ctx context.Context, fileName string, state int, fileSize float64) error

	GetExportTasks(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]map[string]interface{}, error)

	DeleteExportTasks(ctx context.Context, fileNames []string) error

	GetCollectionCursor(ctx context.Context, collection string, query bson.M, opts *options.FindOptions) (*mongo.Cursor, error)

	GetExportRecords(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]models.ExportTask, error)
}

type ExportTask struct {
	FileName   string    `bson:"file_name"`
	CreateTime time.Time `bson:"create_time"`
	Quantity   int       `bson:"quantity"`
	DataType   string    `bson:"data_type"`
	FileType   string    `bson:"file_type"`
	State      int       `bson:"state"`
	EndTime    string    `bson:"end_time,omitempty"`
	FileSize   string    `bson:"file_size,omitempty"`
}

type repository struct {
	db *mongo.Database
}

func NewRepository() Repository {
	return &repository{
		db: mongodb.DB,
	}
}

func (r *repository) CreateExportTask(ctx context.Context, task *models.ExportTask) error {
	_, err := r.db.Collection("export").InsertOne(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to create export task: %w", err)
	}
	return nil
}

func (r *repository) UpdateExportTaskStatus(ctx context.Context, fileName string, state int, fileSize float64) error {
	update := bson.M{
		"$set": bson.M{
			"state":     state,
			"end_time":  helper.GetNowTimeString(),
			"file_size": fmt.Sprintf("%.2f", fileSize),
		},
	}

	_, err := r.db.Collection("export").UpdateOne(
		ctx,
		bson.M{"file_name": fileName},
		update,
	)
	if err != nil {
		return fmt.Errorf("failed to update export task status: %w", err)
	}
	return nil
}

func (r *repository) GetExportTasks(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]map[string]interface{}, error) {
	cursor, err := r.db.Collection("export").Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find export tasks: %w", err)
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode export tasks: %w", err)
	}

	return results, nil
}

func (r *repository) DeleteExportTasks(ctx context.Context, fileNames []string) error {
	_, err := r.db.Collection("export").DeleteMany(
		ctx,
		bson.M{"file_name": bson.M{"$in": fileNames}},
	)
	if err != nil {
		return fmt.Errorf("failed to delete export tasks: %w", err)
	}
	return nil
}

func (r *repository) GetCollectionCursor(ctx context.Context, collection string, query bson.M, opts *options.FindOptions) (*mongo.Cursor, error) {

	if opts == nil {
		opts = options.Find()
	}

	opts.SetBatchSize(1000)

	cursor, err := r.db.Collection(collection).Find(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get cursor: %w", err)
	}

	return cursor, nil
}

func (r *repository) GetExportRecords(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]models.ExportTask, error) {
	cursor, err := r.db.Collection("export").Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find export records: %w", err)
	}
	defer cursor.Close(ctx)

	var results []models.ExportTask
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode export records: %w", err)
	}

	return results, nil
}
