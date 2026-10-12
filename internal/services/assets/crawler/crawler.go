package crawler

import (
	"context"
	"recon/internal/utils/helper"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"recon/internal/models"
	"recon/internal/repositories/assets/crawler"
)

type Service interface {
	GetCrawlers(ctx context.Context, query models.SearchRequest) ([]models.CrawlerResult, error)
}

type service struct {
	repo crawler.Repository
}

func NewService() Service {
	return &service{
		repo: crawler.NewRepository(),
	}
}

func (s *service) GetCrawlers(ctx context.Context, query models.SearchRequest) ([]models.CrawlerResult, error) {

	query.Index = "crawler"
	searchQuery, err := helper.GetSearchQuery(query)
	if err != nil {
		return nil, err
	}
	filter := bson.M(searchQuery)

	projection := bson.M{
		"_id":    1,
		"method": 1,
		"body":   1,
		"url":    1,
		"tags":   1,
		"time":   1,
	}

	opts := options.Find().
		SetProjection(projection).
		SetSkip(int64((query.PageIndex - 1) * query.PageSize)).
		SetLimit(int64(query.PageSize)).
		SetSort(bson.D{{Key: "time", Value: -1}})

	return s.repo.Find(ctx, filter, opts)
}
