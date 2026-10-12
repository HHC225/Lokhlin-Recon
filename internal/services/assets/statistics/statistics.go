package statistics

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"recon/internal/models"
	"recon/internal/repositories/assets/common"
	"recon/internal/repositories/assets/statistics"
	"recon/internal/services/assets/asset"
	"sync"
)

type Service interface {
	GetStatisticsData(ctx *gin.Context) (*models.StatisticsData, error)
	GetPortStatistics(ctx *gin.Context, req *models.SearchRequest) (*models.PortStatisticsResponse, error)
	GetTitleStatistics(ctx *gin.Context, req *models.TitleStatisticsRequest) (*models.TitleStatisticsResponse, error)
	GetTypeStatistics(ctx *gin.Context, req *models.SearchRequest) (*models.TypeStatisticsResponse, error)
	GetIconStatistics(ctx *gin.Context, req *models.SearchRequest) (*models.IconStatisticsResponse, error)
	GetAppStatistics(ctx *gin.Context, req *models.SearchRequest) (*models.AppStatisticsResponse, error)
}

type service struct {
	repo         statistics.Repository
	commonRepo   common.Repository
	assetService asset.Service
}

func NewService() Service {
	return &service{
		repo:         statistics.NewRepository(),
		commonRepo:   common.NewRepository(),
		assetService: asset.NewService(),
	}
}

func (s *service) GetStatisticsData(ctx *gin.Context) (*models.StatisticsData, error) {
	var wg sync.WaitGroup
	var err error

	var assetCount, subdomainCount, sensitiveCount, urlCount, vulnerabilityCount int64

	wg.Add(5)

	go func() {
		defer wg.Done()
		assetCount, err = s.commonRepo.EstimatedDocumentCount(ctx, "asset")
	}()

	go func() {
		defer wg.Done()
		subdomainCount, err = s.commonRepo.EstimatedDocumentCount(ctx, "subdomain")
	}()

	go func() {
		defer wg.Done()
		sensitiveCount, err = s.commonRepo.EstimatedDocumentCount(ctx, "SensitiveResult")
	}()

	go func() {
		defer wg.Done()
		urlCount, err = s.commonRepo.EstimatedDocumentCount(ctx, "UrlScan")
	}()

	go func() {
		defer wg.Done()
		vulnerabilityCount, err = s.commonRepo.EstimatedDocumentCount(ctx, "vulnerability")
	}()

	wg.Wait()

	if err != nil {
		return nil, err
	}

	return &models.StatisticsData{
		AssetCount:         assetCount,
		SubdomainCount:     subdomainCount,
		SensitiveCount:     sensitiveCount,
		UrlCount:           urlCount,
		VulnerabilityCount: vulnerabilityCount,
	}, nil
}

func (s *service) GetPortStatistics(ctx *gin.Context, query *models.SearchRequest) (*models.PortStatisticsResponse, error) {
	query.Index = "asset"
	searchQuery, err := s.assetService.BuildSearchQuery(*query)
	if err != nil {
		return nil, err
	}
	filter := bson.M(searchQuery)

	pipeline := []bson.M{
		{
			"$match": filter,
		},
		{
			"$group": bson.M{
				"_id":    "$port",
				"number": bson.M{"$sum": 1},
			},
		},
		{
			"$match": bson.M{
				"_id": bson.M{"$ne": nil},
			},
		},
		{
			"$sort": bson.M{
				"number": -1,
			},
		},
		{
			"$limit": 200,
		},
	}

	cursor, err := s.repo.Aggregate(ctx, "asset", pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []models.PortItem
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}

	return &models.PortStatisticsResponse{
		Port: items,
	}, nil
}

func (s *service) GetTitleStatistics(ctx *gin.Context, req *models.TitleStatisticsRequest) (*models.TitleStatisticsResponse, error) {
	//req.Filter["type"] = []string{"https", "http"}
	//
	//pipeline := []bson.M{
	//	{
	//		"$match": req.Filter,
	//	},
	//	{
	//		"$facet": bson.M{
	//			"by_title": []bson.M{
	//				{"$group": bson.M{"_id": "$title", "num_tutorial": bson.M{"$sum": 1}}},
	//				{"$match": bson.M{"_id": bson.M{"$ne": ""}}},
	//			},
	//		},
	//	},
	//}
	//
	//result, err := s.repo.Aggregate(ctx, "asset", pipeline)
	//if err != nil {
	//	return nil, err
	//}
	//
	//response := &models.TitleStatisticsResponse{
	//	Title: make([]models.TitleItem, 0),
	//}
	//
	//for _, r := range result {
	//	byTitle := r["by_title"].([]interface{})
	//	for _, t := range byTitle {
	//		title := t.(bson.M)
	//		response.Title = append(response.Title, models.TitleItem{
	//			Value:  title["_id"].(string),
	//			Number: title["num_tutorial"].(int64),
	//		})
	//	}
	//}

	return nil, nil
}

func (s *service) GetTypeStatistics(ctx *gin.Context, query *models.SearchRequest) (*models.TypeStatisticsResponse, error) {
	query.Index = "asset"
	searchQuery, err := s.assetService.BuildSearchQuery(*query)
	if err != nil {
		return nil, err
	}
	filter := bson.M(searchQuery)
	pipeline := []bson.M{
		{
			"$match": filter,
		},
		{
			"$group": bson.M{
				"_id":    "$service",
				"number": bson.M{"$sum": 1},
			},
		},
		{
			"$match": bson.M{
				"_id": bson.M{"$ne": nil},
			},
		},
		{
			"$sort": bson.M{
				"number": -1,
			},
		},
	}

	cursor, err := s.repo.Aggregate(ctx, "asset", pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []models.TypeItem
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}

	return &models.TypeStatisticsResponse{
		Service: items,
	}, nil
}

func (s *service) GetIconStatistics(ctx *gin.Context, query *models.SearchRequest) (*models.IconStatisticsResponse, error) {
	skip := (query.PageIndex - 1) * query.PageSize
	searchQuery, err := s.assetService.BuildSearchQuery(*query)
	if err != nil {
		return nil, err
	}
	filter := bson.M(searchQuery)
	pipeline := []bson.M{
		{"$match": filter},
		{"$project": bson.M{
			"faviconmmh3": 1,
			"iconcontent": 1,
		}},
		{"$group": bson.M{
			"_id":         "$faviconmmh3",
			"number":      bson.M{"$sum": 1},
			"iconcontent": bson.M{"$first": "$iconcontent"},
		}},
		{"$match": bson.M{
			"$and": []bson.M{
				{"_id": bson.M{"$ne": ""}},
				{"_id": bson.M{"$ne": nil}},
			},
		}},
		{"$sort": bson.M{"number": -1}},
		{"$skip": skip},
		{"$limit": query.PageSize},
	}

	cursor, err := s.repo.Aggregate(ctx, "asset", pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []models.IconItem
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}

	return &models.IconStatisticsResponse{
		Icon: items,
	}, nil
}

func (s *service) GetAppStatistics(ctx *gin.Context, query *models.SearchRequest) (*models.AppStatisticsResponse, error) {
	query.Index = "asset"
	searchQuery, err := s.assetService.BuildSearchQuery(*query)
	if err != nil {
		return nil, err
	}
	filter := bson.M(searchQuery)

	pipeline := []bson.M{
		{"$match": filter},
		{"$project": bson.M{
			"technologies": 1,
		}},
		{"$unwind": "$technologies"},
		{"$group": bson.M{
			"_id":    "$technologies",
			"number": bson.M{"$sum": 1},
		}},
		{"$sort": bson.M{"number": -1}},
		{
			"$limit": 200,
		},
	}

	cursor, err := s.repo.Aggregate(ctx, "asset", pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []models.AppItem
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}

	return &models.AppStatisticsResponse{
		Product: items,
	}, nil
}
