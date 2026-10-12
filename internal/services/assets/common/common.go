package common

import (
	"context"
	"errors"
	"io"
	"strings"

	"recon/internal/services/assets/asset"

	"recon/internal/logger"
	"recon/internal/models"
	"recon/internal/repositories/assets/common"
	commonrepo "recon/internal/repositories/common"
	"recon/internal/utils/helper"
	"recon/internal/utils/random"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

var (
	validCollections = []string{
		"asset", "DirScanResult", "SensitiveResult", "SubdomainTakerResult",
		"UrlScan", "crawler", "subdomain", "vulnerability", "PageMonitoring",
		"app", "RootDomain", "mp", "IPAsset",
	}
)

type Service interface {
	DeleteData(ctx *gin.Context, req *models.DeleteRequest) error
	AddTag(ctx *gin.Context, req *models.TagRequest) error
	DeleteTag(ctx *gin.Context, req *models.TagRequest) error
	UpdateStatus(ctx *gin.Context, req *models.StatusRequest) error
	TotalData(ctx context.Context, req *models.SearchRequest) (int64, error)
}

type service struct {
	repo         common.Repository
	assetService asset.Service
	commonRepo   commonrepo.Repository
}

func NewService() Service {
	return &service{
		repo:         common.NewRepository(),
		assetService: asset.NewService(),
		commonRepo:   commonrepo.NewRepository(),
	}
}

func (s *service) DeleteData(ctx *gin.Context, req *models.DeleteRequest) error {
	if !isValidCollection(req.Index) {
		return ErrInvalidCollection
	}

	var objIDs []primitive.ObjectID
	for _, id := range req.IDs {
		if id != "" && !strings.HasPrefix(id, "http://") && !strings.HasPrefix(id, "https://") && len(id) > 6 {
			if objID, err := primitive.ObjectIDFromHex(id); err == nil {
				objIDs = append(objIDs, objID)
			}
		}
	}
	var filter bson.M
	if len(objIDs) > 0 {
		filter = bson.M{"_id": bson.M{"$in": objIDs}}
	} else {
		return ErrNoValidIDs
	}
	if req.Index == "vulnerability" {
		var vulnerabilities []models.Vulnerability
		err := s.repo.Find(context.Background(), req.Index, filter, -1, []string{"hash"}, &vulnerabilities)
		if err != nil {
			return err
		}
		go func() {
			hashList := []string{}
			for _, vulnerability := range vulnerabilities {
				hashList = append(hashList, vulnerability.Hash)
			}
			vulDetailFilter := bson.M{"hash": bson.M{"$in": hashList}}
			_, err := s.repo.DeleteMany(context.Background(), "vulnerabilityDetail", vulDetailFilter)
			if err != nil {
				logger.Error(err.Error())
				return
			}
		}()
	}

	var bodyHashesToCheck []string
	if req.Index == "asset" {
		var assets []models.Asset
		err := s.repo.Find(context.Background(), req.Index, filter, -1, []string{"bodyhash"}, &assets)
		if err != nil {
			return err
		}

		hashMap := make(map[string]bool)
		for _, asset := range assets {
			if asset.ResponseBodyHash != "" {
				hashMap[asset.ResponseBodyHash] = true
			}
		}
		for hash := range hashMap {
			bodyHashesToCheck = append(bodyHashesToCheck, hash)
		}
	}

	filter = bson.M{"_id": bson.M{"$in": objIDs}}

	count, err := s.repo.DeleteMany(ctx.Request.Context(), req.Index, filter)
	if err != nil {
		return err
	}

	if count == 0 {
		return ErrNoDocumentsDeleted
	}

	if req.Index == "asset" && len(bodyHashesToCheck) > 0 {
		go func() {

			pipeline := mongo.Pipeline{
				{{Key: "$match", Value: bson.M{"bodyhash": bson.M{"$in": bodyHashesToCheck}}}},
				{{Key: "$group", Value: bson.M{"_id": "$bodyhash"}}},
			}

			cursor, err := s.commonRepo.Aggregate(context.Background(), "asset", pipeline)
			if err != nil {
				logger.Error("failed to aggregate bodyhash", zap.Error(err))
				return
			}
			defer cursor.Close(context.Background())

			usedHashes := make(map[string]bool)
			for cursor.Next(context.Background()) {
				var result struct {
					ID string `bson:"_id"`
				}
				if err := cursor.Decode(&result); err != nil {
					logger.Error("failed to decode bodyhash", zap.Error(err))
					continue
				}
				if result.ID != "" {
					usedHashes[result.ID] = true
				}
			}

			unusedHashes := make([]string, 0)
			for _, hash := range bodyHashesToCheck {
				if !usedHashes[hash] {
					unusedHashes = append(unusedHashes, hash)
				}
			}

			if len(unusedHashes) > 0 {
				httpBodyFilter := bson.M{"hash": bson.M{"$in": unusedHashes}}
				_, err := s.repo.DeleteMany(context.Background(), "HttpBody", httpBodyFilter)
				if err != nil {
					logger.Error("failed to delete HttpBody", zap.Error(err))
					return
				}
				logger.Info("deleted unused HttpBody", zap.Int("count", len(unusedHashes)))
			}
		}()
	}

	return nil
}

func (s *service) AddTag(ctx *gin.Context, req *models.TagRequest) error {
	if !isValidCollection(req.Type) || req.ID == "" || req.Tag == "" {
		return ErrInvalidRequest
	}

	objID, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		return ErrInvalidID
	}

	doc, err := s.repo.FindOne(ctx.Request.Context(), req.Type, bson.M{"_id": objID})
	if err != nil {
		return err
	}

	if doc.Tags == nil {
		doc.Tags = []string{req.Tag}
	} else {
		doc.Tags = append(doc.Tags, req.Tag)
	}

	return s.repo.UpdateOne(ctx.Request.Context(), req.Type, bson.M{"_id": objID}, bson.M{"$set": bson.M{"tags": doc.Tags}})
}

func (s *service) DeleteTag(ctx *gin.Context, req *models.TagRequest) error {
	if !isValidCollection(req.Type) || req.ID == "" || req.Tag == "" {
		return ErrInvalidRequest
	}

	objID, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		return ErrInvalidID
	}

	doc, err := s.repo.FindOne(ctx.Request.Context(), req.Type, bson.M{"_id": objID})
	if err != nil {
		return err
	}

	if doc.Tags == nil {
		return nil
	}

	newTags := make([]string, 0)
	for _, tag := range doc.Tags {
		if tag != req.Tag {
			newTags = append(newTags, tag)
		}
	}

	return s.repo.UpdateOne(ctx.Request.Context(), req.Type, bson.M{"_id": objID}, bson.M{"$set": bson.M{"tags": newTags}})
}

func (s *service) UpdateStatus(ctx *gin.Context, req *models.StatusRequest) error {
	validTypes := []string{"SensitiveResult", "vulnerability"}
	if !contains(validTypes, req.Type) || req.ID == "" {
		return ErrInvalidRequest
	}

	objID, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		return ErrInvalidID
	}

	return s.repo.UpdateOne(ctx.Request.Context(), req.Type, bson.M{"_id": objID}, bson.M{"$set": bson.M{"status": req.Status}})
}

func (s *service) TotalData(ctx context.Context, req *models.SearchRequest) (int64, error) {
	if !isValidCollection(req.Index) {
		return 0, ErrInvalidCollection
	}
	var searchQuery map[string]interface{}
	var err error
	if req.Index == "asset" {
		searchQuery, err = s.assetService.BuildSearchQuery(*req)
	} else {
		searchQuery, err = helper.GetSearchQuery(*req)
	}
	if err != nil {
		return 0, err
	}
	if len(searchQuery) == 0 {

		return s.repo.EstimatedDocumentCount(ctx, req.Index)
	}
	filter := bson.M(searchQuery)
	return s.repo.CountDocuments(ctx, req.Index, filter)
}

func isValidCollection(collection string) bool {
	return contains(validCollections, collection)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func getSearchQuery(index string, body io.ReadCloser) (bson.M, error) {

	return bson.M{}, nil
}

var (
	ErrInvalidCollection  = errors.New("invalid collection")
	ErrInvalidRequest     = errors.New("invalid request")
	ErrInvalidID          = errors.New("invalid id")
	ErrNoValidIDs         = errors.New("no valid ids")
	ErrNoDocumentsDeleted = errors.New("no documents deleted")
)

// ---------------------------------------------

// ---------------------------------------------

type DedupService interface {
	AssetDataDedup(ctx context.Context, collection string, filters bson.M, groups []string, subdomain bool) error

	DoAssetDeduplication() error
}

type dedupService struct {
	repo commonrepo.Repository
}

func NewDedupService() DedupService {
	return &dedupService{repo: commonrepo.NewRepository()}
}

func (s *dedupService) DoAssetDeduplication() error {

	ctx := context.Background()
	cfg, err := s.repo.FindOne(ctx, "config", bson.M{"name": "deduplication"}, nil)
	if err != nil {
		return err
	}

	delete(cfg, "_id")
	delete(cfg, "name")
	delete(cfg, "hour")
	delete(cfg, "flag")

	fgk := map[string]struct {
		Filters bson.M
		Groups  []string
	}{
		"DirScanResult":        {Filters: bson.M{}, Groups: []string{"url", "status", "msg"}},
		"SubdomainTakerResult": {Filters: bson.M{}, Groups: []string{"input", "value"}},
		"UrlScan":              {Filters: bson.M{}, Groups: []string{"output"}},
		"crawler":              {Filters: bson.M{}, Groups: []string{"url", "body"}},
		"subdomain":            {Filters: bson.M{}, Groups: []string{"host", "type", "sorted_ip"}},
		"vulnerability":        {Filters: bson.M{}, Groups: []string{"url", "vulnid", "matched"}},
	}

	for key, val := range cfg {
		enabled, ok := val.(bool)
		if !ok || !enabled {
			continue
		}

		switch key {
		case "asset":
			logger.Info("dedup start", zap.String("collection", "asset"))

			if err := s.AssetDataDedup(ctx, "asset", bson.M{"type": bson.M{"$ne": "other"}}, []string{"url", "statuscode", "hashes.body_mmh3"}, false); err != nil {
				logger.Error("dedup failed", zap.String("collection", "asset"), zap.Error(err))
				return err
			}

			if err := s.AssetDataDedup(ctx, "asset", bson.M{"type": "other"}, []string{"host", "ip", "protocol"}, false); err != nil {
				logger.Error("dedup failed", zap.String("collection", "asset"), zap.Error(err))
				return err
			}
			logger.Info("dedup done", zap.String("collection", "asset"))
		case "subdomain":
			conf := fgk["subdomain"]
			logger.Info("dedup start", zap.String("collection", "subdomain"))
			if err := s.AssetDataDedup(ctx, "subdomain", conf.Filters, conf.Groups, true); err != nil {
				logger.Error("dedup failed", zap.String("collection", "subdomain"), zap.Error(err))
				return err
			}
			logger.Info("dedup done", zap.String("collection", "subdomain"))
		default:
			if conf, ok := fgk[key]; ok {
				logger.Info("dedup start", zap.String("collection", key))
				if err := s.AssetDataDedup(ctx, key, conf.Filters, conf.Groups, false); err != nil {
					logger.Error("dedup failed", zap.String("collection", key), zap.Error(err))
					return err
				}
				logger.Info("dedup done", zap.String("collection", key))
			}
		}
	}

	return nil
}

func (s *dedupService) AssetDataDedup(ctx context.Context, collection string, filters bson.M, groups []string, subdomain bool) error {

	processFlag := random.GenerateString(12)

	updatedFlagCount, err := s.repo.UpdateMany(ctx, collection, filters, bson.M{"$set": bson.M{"process_flag": processFlag}})
	if err != nil {
		logger.Error("set process_flag failed", zap.String("collection", collection), zap.Error(err))
		return err
	}
	logger.Info("set process_flag", zap.String("collection", collection), zap.Int64("matched", updatedFlagCount))

	match := bson.M{"process_flag": processFlag}
	for k, v := range filters {
		match[k] = v
	}

	idGroup := bson.M{}
	for _, g := range groups {
		cleaned := strings.ReplaceAll(g, ".", "")
		idGroup[cleaned] = "$" + g
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: match}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "_id", Value: -1}}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: idGroup},
			{Key: "latestId", Value: bson.D{{Key: "$first", Value: "$_id"}}},
		}}},
		bson.D{{Key: "$project", Value: bson.D{{Key: "_id", Value: 0}, {Key: "latestId", Value: 1}}}},
	}

	if subdomain {
		matchStage := bson.D{{Key: "$match", Value: match}}
		addFieldsStage := bson.D{{Key: "$addFields", Value: bson.M{
			"sorted_ip": bson.M{
				"$sortArray": bson.M{"input": "$ip", "sortBy": 1},
			},
		}}}
		sortStage := bson.D{{Key: "$sort", Value: bson.D{{Key: "_id", Value: -1}}}}
		groupStage := bson.D{{Key: "$group", Value: bson.M{"_id": idGroup, "latestId": bson.M{"$first": "$_id"}}}}
		projectStage := bson.D{{Key: "$project", Value: bson.M{"_id": 0, "latestId": 1}}}
		pipeline = mongo.Pipeline{matchStage, addFieldsStage, sortStage, groupStage, projectStage}
	}

	cur, err := s.repo.Aggregate(ctx, collection, pipeline)
	if err != nil {
		logger.Error("aggregate failed", zap.String("collection", collection), zap.Error(err))
		return err
	}
	defer cur.Close(ctx)

	type latest struct {
		LatestID primitive.ObjectID `bson:"latestId"`
	}
	ids := make([]primitive.ObjectID, 0)
	for cur.Next(ctx) {
		var l latest
		if err := cur.Decode(&l); err != nil {
			return err
		}
		ids = append(ids, l.LatestID)
	}
	if err := cur.Err(); err != nil {
		logger.Error("cursor error", zap.String("collection", collection), zap.Error(err))
		return err
	}

	if len(ids) > 0 {
		setLatestCount, err := s.repo.UpdateMany(ctx, collection, bson.M{"_id": bson.M{"$in": ids}}, bson.M{"$set": bson.M{"latest": true}})
		if err != nil {
			logger.Error("set latest failed", zap.String("collection", collection), zap.Error(err))
			return err
		}
		logger.Info("set latest", zap.String("collection", collection), zap.Int("ids", len(ids)), zap.Int64("modified", setLatestCount))
	}

	deletedCount, err := s.repo.DeleteMany(ctx, collection, bson.M{"process_flag": processFlag, "latest": bson.M{"$ne": true}})
	if err != nil {
		logger.Error("delete duplicates failed", zap.String("collection", collection), zap.Error(err))
		return err
	}
	logger.Info("delete duplicates", zap.String("collection", collection), zap.Int64("deleted", deletedCount))

	unsetCount, err := s.repo.UpdateMany(ctx, collection, bson.M{"process_flag": processFlag}, bson.M{"$unset": bson.M{"process_flag": "", "latest": ""}})
	if err != nil {
		logger.Error("unset flags failed", zap.String("collection", collection), zap.Error(err))
		return err
	}
	logger.Info("unset flags", zap.String("collection", collection), zap.Int64("modified", unsetCount))

	return nil
}
