package page_monitoring

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"recon/internal/database/mongodb"
	"recon/internal/models"
)

type Repository interface {
	FindHistory(ctx context.Context, filter interface{}, opts *options.FindOptions) ([]*models.PageMonitoring, error)
	FindContent(ctx context.Context, filter interface{}, opts *options.FindOneOptions) (*models.PageMonitoringBody, error)
	FindDiff(ctx context.Context, filter interface{}, opts *options.FindOneOptions) (*models.PageMonitoringBody, error)
	Count(ctx context.Context, filter interface{}) (int64, error)
	Find(ctx context.Context, filter interface{}, opts *options.FindOptions) ([]models.PageMonitoringResult, error)
}

type repository struct {
	pageCollection    *mongo.Collection
	contentCollection *mongo.Collection
}

func NewRepository() Repository {
	return &repository{
		pageCollection:    mongodb.DB.Collection("PageMonitoring"),
		contentCollection: mongodb.DB.Collection("PageMonitoringBody"),
	}
}

func (r *repository) FindHistory(ctx context.Context, filter interface{}, opts *options.FindOptions) ([]*models.PageMonitoring, error) {
	cursor, err := r.pageCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var pages []*models.PageMonitoring
	if err = cursor.All(ctx, &pages); err != nil {
		return nil, err
	}

	return pages, nil
}

func (r *repository) FindContent(ctx context.Context, filter interface{}, opts *options.FindOneOptions) (*models.PageMonitoringBody, error) {
	var content models.PageMonitoringBody
	err := r.contentCollection.FindOne(ctx, filter, opts).Decode(&content)
	if err != nil {
		return nil, err
	}
	return &content, nil
}

func (r *repository) FindDiff(ctx context.Context, filter interface{}, opts *options.FindOneOptions) (*models.PageMonitoringBody, error) {
	var content models.PageMonitoringBody
	err := r.contentCollection.FindOne(ctx, filter, opts).Decode(&content)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("content not found for the provided ID")
		}
		return nil, err
	}

	if len(content.Content) < 2 {
		return nil, fmt.Errorf("invalid content format")
	}

	return &content, nil
}

func (r *repository) Count(ctx context.Context, filter interface{}) (int64, error) {
	return r.pageCollection.CountDocuments(ctx, filter)
}

func (r *repository) Find(ctx context.Context, filter interface{}, opts *options.FindOptions) ([]models.PageMonitoringResult, error) {
	cursor, err := r.pageCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}
	defer cursor.Close(ctx)

	var results []models.PageMonitoringResult
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode documents: %w", err)
	}

	return results, nil
}
