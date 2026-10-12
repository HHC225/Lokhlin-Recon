package url

import (
	"context"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"recon/internal/utils/helper"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"recon/internal/models"
	"recon/internal/repositories/assets/url"
)

type Service interface {
	GetURLs(ctx context.Context, query models.SearchRequest) ([]models.URL, error)
	GetTaskTarget(ctx context.Context, query models.SearchRequest) (string, error)
	GetTaskTargetByIDs(ctx context.Context, ids []string) (string, error)
}

type service struct {
	repo url.Repository
}

func NewService() Service {
	return &service{
		repo: url.NewRepository(),
	}
}

func (s *service) GetTaskTargetByIDs(ctx context.Context, ids []string) (string, error) {
	if len(ids) == 0 {
		return "", nil
	}

	var objectIDs []primitive.ObjectID
	for _, id := range ids {
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return "", err
		}
		objectIDs = append(objectIDs, objectID)
	}

	filter := bson.M{"_id": bson.M{"$in": objectIDs}}
	projection := bson.M{
		"output": 1,
	}

	opts := options.Find().
		SetProjection(projection)

	urlResults, err := s.repo.Find(ctx, filter, opts)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	for i, result := range urlResults {
		builder.WriteString(result.Output)
		if i != len(urlResults)-1 {
			builder.WriteString("\n")
		}
	}
	return builder.String(), nil
}

func (s *service) GetURLs(ctx context.Context, query models.SearchRequest) ([]models.URL, error) {

	query.Index = "UrlScan"
	searchQuery, err := helper.GetSearchQuery(query)
	if err != nil {
		return nil, err
	}
	filter := bson.M(searchQuery)

	projection := bson.M{
		"_id":        1,
		"input":      1,
		"source":     1,
		"status":     1,
		"length":     1,
		"outputtype": 1,
		"output":     1,
		"time":       1,
		"tags":       1,
	}

	sortBy := bson.D{{Key: "_id", Value: -1}}
	if val, ok := query.Sort["length"]; ok {
		dir := -1
		if val == "ascending" {
			dir = 1
		}
		sortBy = bson.D{{Key: "length", Value: dir}}
	}

	opts := options.Find().
		SetProjection(projection).
		SetSkip(int64((query.PageIndex - 1) * query.PageSize)).
		SetLimit(int64(query.PageSize)).
		SetSort(sortBy)

	return s.repo.Find(ctx, filter, opts)
}

func (s *service) GetTaskTarget(ctx context.Context, query models.SearchRequest) (string, error) {
	query.Index = "UrlScan"
	searchQuery, err := helper.GetSearchQuery(query)
	if err != nil {
		return "", err
	}
	filter := bson.M(searchQuery)
	projection := bson.M{
		"output": 1,
	}

	opts := options.Find().
		SetProjection(projection)

	if query.PageSize != 0 {
		opts = opts.SetLimit(int64(query.PageSize))
	}

	urlResults, err := s.repo.Find(ctx, filter, opts)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	for i, result := range urlResults {
		builder.WriteString(result.Output)
		if i != len(urlResults)-1 {
			builder.WriteString("\n")
		}
	}
	return builder.String(), nil
}
