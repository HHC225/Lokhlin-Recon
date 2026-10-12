package subdomain

import (
	"context"
	"strings"

	"recon/internal/utils/helper"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"recon/internal/models"
	"recon/internal/repositories/assets/subdomain"
)

type Service interface {
	GetSubdomains(ctx context.Context, query models.SearchRequest) ([]models.Subdomain, error)
	GetTaskTarget(ctx context.Context, query models.SearchRequest) (string, error)
	GetTaskTargetByIDs(ctx context.Context, ids []string) (string, error)
}

type service struct {
	repo subdomain.Repository
}

func NewService() Service {
	return &service{
		repo: subdomain.NewRepository(),
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
		"host": 1,
	}

	opts := options.Find().
		SetProjection(projection)

	subDomains, err := s.repo.FindWithPagination(ctx, filter, opts)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	for i, subDomain := range subDomains {
		builder.WriteString(subDomain.Host)
		if i != len(subDomains)-1 {
			builder.WriteString("\n")
		}
	}
	return builder.String(), nil
}

func (s *service) GetSubdomains(ctx context.Context, query models.SearchRequest) ([]models.Subdomain, error) {

	query.Index = "subdomain"
	searchQuery, err := helper.GetSearchQuery(query)
	if err != nil {
		return nil, err
	}
	filter := bson.M(searchQuery)

	projection := bson.M{
		"_id":   1,
		"host":  1,
		"type":  1,
		"value": 1,
		"ip":    1,
		"time":  1,
		"tags":  1,
	}

	opts := options.Find().
		SetProjection(projection).
		SetSkip(int64((query.PageIndex - 1) * query.PageSize)).
		SetLimit(int64(query.PageSize)).
		SetSort(bson.D{{Key: "time", Value: -1}})

	return s.repo.FindWithPagination(ctx, filter, opts)
}

func (s *service) GetTaskTarget(ctx context.Context, query models.SearchRequest) (string, error) {
	query.Index = "subdomain"
	searchQuery, err := helper.GetSearchQuery(query)
	if err != nil {
		return "", err
	}
	filter := bson.M(searchQuery)
	projection := bson.M{
		"host": 1,
	}

	opts := options.Find().
		SetProjection(projection)

	if query.PageSize != 0 {
		opts = opts.SetLimit(int64(query.PageSize))
	}

	subDomains, err := s.repo.FindWithPagination(ctx, filter, opts)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	for i, rootDomain := range subDomains {
		builder.WriteString(rootDomain.Host)
		if i != len(subDomains)-1 {
			builder.WriteString("\n")
		}
	}
	return builder.String(), nil
}
