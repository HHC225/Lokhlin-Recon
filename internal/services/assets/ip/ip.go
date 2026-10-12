package ip

import (
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"recon/internal/models"
	iprepo "recon/internal/repositories/assets/ip"
	"recon/internal/utils/helper"
	"recon/internal/utils/random"
)

type Service interface {
	GetIPAssets(ctx *gin.Context, query models.SearchRequest) (*models.IPAssetResponse, error)
}

type service struct {
	repo iprepo.Repository
}

func NewService() Service {
	return &service{
		repo: iprepo.NewRepository(),
	}
}

func (s *service) GetIPAssets(ctx *gin.Context, query models.SearchRequest) (*models.IPAssetResponse, error) {
	if query.PageIndex <= 0 {
		query.PageIndex = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}

	query.Index = "IPAsset"
	searchQuery, err := helper.GetSearchQuery(query)
	if err != nil {
		return nil, err
	}
	filter := bson.M(searchQuery)

	reqCtx := ctx.Request.Context()

	hasNestedField := hasNestedFieldQuery(filter)

	var assets []models.IPAsset
	var total int64

	if hasNestedField {

		pipeline := buildAggregationPipeline(filter, query.PageIndex, query.PageSize)

		countPipeline := buildCountPipeline(filter)
		total, err = s.repo.CountWithAggregation(reqCtx, countPipeline)
		if err != nil {
			return nil, err
		}

		assets, err = s.repo.FindWithAggregation(reqCtx, pipeline)
		if err != nil {
			return nil, err
		}
	} else {

		total, err = s.repo.Count(reqCtx, filter)
		if err != nil {
			return nil, err
		}

		projection := bson.M{
			"ip":    1,
			"ports": 1,
			"time":  1,
		}

		opts := options.Find().
			SetProjection(projection).
			SetSkip(int64((query.PageIndex - 1) * query.PageSize)).
			SetLimit(int64(query.PageSize)).
			SetSort(bson.D{{Key: "time", Value: -1}})

		assets, err = s.repo.Find(reqCtx, filter, opts)
		if err != nil {
			return nil, err
		}
	}

	rows := flattenIPAssets(assets)

	return &models.IPAssetResponse{
		List:  rows,
		Total: total,
	}, nil
}

func hasNestedFieldQuery(filter bson.M) bool {
	nestedFields := map[string]bool{
		"ports.port":                true,
		"ports.server.domain":       true,
		"ports.server.service":      true,
		"ports.server.webServer":    true,
		"ports.server.technologies": true,
	}

	return checkNestedFieldInQuery(filter, nestedFields)
}

func checkNestedFieldInQuery(query interface{}, nestedFields map[string]bool) bool {
	switch v := query.(type) {
	case bson.M:
		for key, value := range v {

			if nestedFields[key] {
				return true
			}

			if strings.HasPrefix(key, "ports.") {
				return true
			}

			if checkNestedFieldInQuery(value, nestedFields) {
				return true
			}
		}
	case bson.A:
		for _, item := range v {
			if checkNestedFieldInQuery(item, nestedFields) {
				return true
			}
		}
	case bson.D:
		for _, elem := range v {
			if nestedFields[elem.Key] {
				return true
			}
			if strings.HasPrefix(elem.Key, "ports.") {
				return true
			}
			if checkNestedFieldInQuery(elem.Value, nestedFields) {
				return true
			}
		}
	case []interface{}:
		for _, item := range v {
			if checkNestedFieldInQuery(item, nestedFields) {
				return true
			}
		}
	case map[string]interface{}:
		for key, value := range v {
			if nestedFields[key] {
				return true
			}
			if strings.HasPrefix(key, "ports.") {
				return true
			}
			if checkNestedFieldInQuery(value, nestedFields) {
				return true
			}
		}
	}
	return false
}

func buildAggregationPipeline(filter bson.M, pageIndex, pageSize int) mongo.Pipeline {

	topLevelFilter := bson.M{}
	nestedFilter := bson.M{}

	topLevelFields := map[string]bool{
		"ip":         true,
		"time":       true,
		"project":    true,
		"taskName":   true,
		"rootDomain": true,
		"_id":        true,
	}

	separateFilters(filter, topLevelFilter, nestedFilter, topLevelFields)

	pipeline := mongo.Pipeline{}

	if len(topLevelFilter) > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: topLevelFilter}})
	}

	pipeline = append(pipeline, bson.D{{Key: "$unwind", Value: bson.M{
		"path":                       "$ports",
		"preserveNullAndEmptyArrays": true,
	}}})

	pipeline = append(pipeline, bson.D{{Key: "$unwind", Value: bson.M{
		"path":                       "$ports.server",
		"preserveNullAndEmptyArrays": true,
	}}})

	if len(nestedFilter) > 0 {

		expandedFilter := expandNestedFilter(nestedFilter)
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: expandedFilter}})
	}

	pipeline = append(pipeline, bson.D{{Key: "$group", Value: bson.M{
		"_id": bson.M{
			"ip":   "$ip",
			"port": "$ports.port",
		},
		"ip":     bson.M{"$first": "$ip"},
		"port":   bson.M{"$first": "$ports.port"},
		"time":   bson.M{"$first": "$time"},
		"origId": bson.M{"$first": "$_id"},
		"server": bson.M{"$push": "$ports.server"},
	}}})

	pipeline = append(pipeline, bson.D{{Key: "$group", Value: bson.M{
		"_id":    "$ip",
		"ip":     bson.M{"$first": "$ip"},
		"time":   bson.M{"$first": "$time"},
		"origId": bson.M{"$first": "$origId"},
		"ports": bson.M{"$push": bson.M{
			"port":   "$port",
			"server": "$server",
		}},
	}}})

	pipeline = append(pipeline, bson.D{{Key: "$addFields", Value: bson.M{
		"_id": "$origId",
	}}})

	pipeline = append(pipeline, bson.D{{Key: "$project", Value: bson.M{
		"origId": 0,
	}}})

	pipeline = append(pipeline, bson.D{{Key: "$sort", Value: bson.D{
		{Key: "time", Value: -1},
	}}})

	skip := (pageIndex - 1) * pageSize
	pipeline = append(pipeline, bson.D{{Key: "$skip", Value: skip}})
	pipeline = append(pipeline, bson.D{{Key: "$limit", Value: pageSize}})

	return pipeline
}

func buildCountPipeline(filter bson.M) mongo.Pipeline {
	topLevelFilter := bson.M{}
	nestedFilter := bson.M{}

	topLevelFields := map[string]bool{
		"ip":         true,
		"time":       true,
		"project":    true,
		"taskName":   true,
		"rootDomain": true,
		"_id":        true,
	}

	separateFilters(filter, topLevelFilter, nestedFilter, topLevelFields)

	pipeline := mongo.Pipeline{}

	if len(topLevelFilter) > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: topLevelFilter}})
	}

	pipeline = append(pipeline, bson.D{{Key: "$unwind", Value: bson.M{
		"path":                       "$ports",
		"preserveNullAndEmptyArrays": true,
	}}})

	pipeline = append(pipeline, bson.D{{Key: "$unwind", Value: bson.M{
		"path":                       "$ports.server",
		"preserveNullAndEmptyArrays": true,
	}}})

	if len(nestedFilter) > 0 {
		expandedFilter := expandNestedFilter(nestedFilter)
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: expandedFilter}})
	}

	pipeline = append(pipeline, bson.D{{Key: "$group", Value: bson.M{
		"_id": "$ip",
	}}})

	return pipeline
}

func separateFilters(filter bson.M, topLevelFilter, nestedFilter bson.M, topLevelFields map[string]bool) {
	for key, value := range filter {
		if topLevelFields[key] || key == "$and" || key == "$or" {

			if key == "$and" || key == "$or" {
				topLevelArray := []interface{}{}
				nestedArray := []interface{}{}

				if arr, ok := value.([]interface{}); ok {
					for _, item := range arr {
						if itemMap, ok := item.(bson.M); ok {
							tempTop := bson.M{}
							tempNested := bson.M{}
							separateFilters(itemMap, tempTop, tempNested, topLevelFields)

							if len(tempTop) > 0 {
								topLevelArray = append(topLevelArray, tempTop)
							}
							if len(tempNested) > 0 {
								nestedArray = append(nestedArray, tempNested)
							}
						} else if itemMap, ok := item.(map[string]interface{}); ok {
							tempTop := bson.M{}
							tempNested := bson.M{}
							separateFilters(bson.M(itemMap), tempTop, tempNested, topLevelFields)

							if len(tempTop) > 0 {
								topLevelArray = append(topLevelArray, tempTop)
							}
							if len(tempNested) > 0 {
								nestedArray = append(nestedArray, tempNested)
							}
						}
					}
				}

				if len(topLevelArray) > 0 {
					topLevelFilter[key] = topLevelArray
				}
				if len(nestedArray) > 0 {
					nestedFilter[key] = nestedArray
				}
			} else {
				topLevelFilter[key] = value
			}
		} else if strings.HasPrefix(key, "ports.") {

			nestedFilter[key] = value
		} else {

			topLevelFilter[key] = value
		}
	}
}

func expandNestedFilter(filter bson.M) bson.M {
	expanded := bson.M{}

	for key, value := range filter {
		if key == "$and" || key == "$or" {

			if arr, ok := value.([]interface{}); ok {
				expandedArr := []interface{}{}
				for _, item := range arr {
					if itemMap, ok := item.(bson.M); ok {
						expandedArr = append(expandedArr, expandNestedFilter(itemMap))
					} else if itemMap, ok := item.(map[string]interface{}); ok {
						expandedArr = append(expandedArr, expandNestedFilter(bson.M(itemMap)))
					} else {
						expandedArr = append(expandedArr, item)
					}
				}
				expanded[key] = expandedArr
			}
		} else {

			expanded[key] = value
		}
	}

	return expanded
}

func flattenIPAssets(assets []models.IPAsset) []models.IPAssetFlat {
	var rows []models.IPAssetFlat

	for _, asset := range assets {
		id := asset.ID.Hex()
		totalRows := 0
		for _, port := range asset.Ports {
			serverCount := len(port.Server)
			if serverCount == 0 {
				serverCount = 1
			}
			totalRows += serverCount
		}
		if totalRows == 0 {
			totalRows = 1
		}

		if len(asset.Ports) == 0 {
			rows = append(rows, models.IPAssetFlat{
				IP:          asset.IP,
				IPRowSpan:   totalRows,
				PortRowSpan: 1,
				Time:        asset.Time,
			})
			continue
		}

		ipRowAssigned := false
		for _, port := range asset.Ports {
			portRowSpan := len(port.Server)
			if portRowSpan == 0 {
				portRowSpan = 1
			}
			portRowAssigned := false

			servers := port.Server
			if len(servers) == 0 {
				servers = []models.PortServer{{}}
			}

			for _, server := range servers {
				row := models.IPAssetFlat{
					ID:        id,
					IP:        asset.IP,
					Port:      port.Port,
					Domain:    server.Domain,
					Service:   server.Service,
					WebServer: server.WebServer,
					DataKey:   random.GenerateRandomString(6),
					Products:  append([]string{}, server.Technologies...),
					Time:      asset.Time,
				}

				if !ipRowAssigned {
					row.IPRowSpan = totalRows
					ipRowAssigned = true
				} else {
					row.IPRowSpan = 0
				}

				if !portRowAssigned {
					row.PortRowSpan = portRowSpan
					portRowAssigned = true
				} else {
					row.PortRowSpan = 0
				}

				rows = append(rows, row)
			}
		}
	}

	return rows
}
