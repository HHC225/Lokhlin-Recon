package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"recon/internal/bootstrap"
	"recon/internal/models"
	"recon/internal/repositories/assets/app"
	"recon/internal/utils/helper"
)

type Service interface {
	GetAppData(ctx *gin.Context, query models.SearchRequest) (*models.AppResponse, error)
}

type service struct {
	repo app.Repository
}

func NewService() Service {
	return &service{
		repo: app.NewRepository(),
	}
}

func (s *service) GetAppData(ctx *gin.Context, query models.SearchRequest) (*models.AppResponse, error) {

	query.Index = "app"
	searchQuery, err := helper.GetSearchQuery(query)
	if err != nil {
		return nil, fmt.Errorf("failed to build search query: %w", err)
	}
	filter := bson.M(searchQuery)

	totalCount, err := s.repo.CountDocuments(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count documents: %w", err)
	}

	if totalCount == 0 {
		return &models.AppResponse{
			List:  []models.AppResult{},
			Total: 0,
		}, nil
	}

	opts := options.Find().
		SetSkip(int64((query.PageIndex - 1) * query.PageSize)).
		SetLimit(int64(query.PageSize)).
		SetSort(bson.D{{"time", -1}})

	results, err := s.repo.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}
	for i := range results {
		if projectName, exists := bootstrap.ProjectList[results[i].Project]; exists {
			results[i].Project = projectName
		} else {
			results[i].Project = ""
		}
	}
	return &models.AppResponse{
		List:  results,
		Total: totalCount,
	}, nil
}
