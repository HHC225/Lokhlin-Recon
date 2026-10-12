package mp

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"recon/internal/bootstrap"
	"recon/internal/models"
	"recon/internal/repositories/assets/mp"
	"recon/internal/utils/helper"
)

type Service interface {
	GetMPData(ctx *gin.Context, query models.SearchRequest) (*models.MPResponse, error)
}

type service struct {
	repo mp.Repository
}

func NewService() Service {
	return &service{
		repo: mp.NewRepository(),
	}
}

func (s *service) GetMPData(ctx *gin.Context, query models.SearchRequest) (*models.MPResponse, error) {

	query.Index = "mp"
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
		return &models.MPResponse{
			List:  []models.MPResult{},
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

	return &models.MPResponse{
		List:  results,
		Total: totalCount,
	}, nil
}
