package fingerprint

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/yaml.v3"
	"recon/internal/models"
	commonRepo "recon/internal/repositories/common"
	repo "recon/internal/repositories/fingerprint"
	"recon/internal/services/node"
)

type BatchAddItem struct {
	ID         string
	Name       string
	Content    string
	UpdateTime string
}

type BatchAddResult struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type Service interface {
	List(ctx context.Context, search string, pageIndex, pageSize int) (list []models.FingerprintRule, total int64, err error)
	Add(ctx context.Context, data models.FingerprintRule) (string, error)
	Update(ctx context.Context, id string, data models.FingerprintRule) error
	Delete(ctx context.Context, ids []string) (int64, error)
	GetVersion(ctx context.Context) (string, error)
	BatchAdd(ctx context.Context, items []BatchAddItem) (results []BatchAddResult, successCount, failedCount int, err error)
}

type service struct {
	repo        repo.Repository
	commonRepo  commonRepo.Repository
	nodeService node.Service
}

func NewService() Service {
	return &service{
		repo:        repo.NewRepository(),
		commonRepo:  commonRepo.NewRepository(),
		nodeService: node.NewService(),
	}
}

func (s *service) List(ctx context.Context, search string, pageIndex, pageSize int) ([]models.FingerprintRule, int64, error) {
	filter := bson.M{"name": bson.M{"$regex": search, "$options": "i"}}

	opts := options.Find().
		SetSkip(int64((pageIndex - 1) * pageSize)).
		SetLimit(int64(pageSize)).
		SetSort(bson.D{{Key: "_id", Value: -1}})

	total, err := s.repo.Count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	list, err := s.repo.List(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *service) Add(ctx context.Context, data models.FingerprintRule) (string, error) {
	if data.Rule == "" {
		return "", errors.New("rule is null")
	}

	// express, err := helper.StringToPostfix(data.Rule)
	// if err != nil || len(express) == 0 {
	// 	return "", errors.New("rule to express error")
	// }
	// data.Express = express
	data.Amount = 0
	id, err := s.repo.Insert(ctx, &data)
	if err != nil {
		return "", err
	}
	_ = s.nodeService.RefreshConfig(ctx, models.Message{Name: "all", Type: "finger"})
	return id, nil
}

func (s *service) Update(ctx context.Context, id string, data models.FingerprintRule) error {
	if data.Rule == "" {
		return errors.New("rule is null")
	}

	// express, err := helper.StringToPostfix(data.Rule)
	// if err != nil || len(express) == 0 {
	// 	return errors.New("rule to express error")
	// }
	update := bson.M{"$set": bson.M{
		"name":            data.Name,
		"rule":            data.Rule,
		"category":        data.Category,
		"parent_category": data.ParentCategory,
		"state":           data.State,
		"company":         data.Company,
	}}
	if err := s.repo.Update(ctx, id, update); err != nil {
		return err
	}
	_ = s.nodeService.RefreshConfig(ctx, models.Message{Name: "all", Type: "finger"})
	return nil
}

func (s *service) Delete(ctx context.Context, ids []string) (int64, error) {
	deleted, err := s.repo.DeleteMany(ctx, ids)
	if err != nil {
		return 0, err
	}
	_ = s.nodeService.RefreshConfig(ctx, models.Message{Name: "all", Type: "finger"})
	return deleted, nil
}

func (s *service) GetVersion(ctx context.Context) (string, error) {
	doc, err := s.commonRepo.FindOne(ctx, "config", bson.M{"name": "FingerVersion"}, bson.M{"_id": 0})
	if err != nil {
		return "", err
	}
	if v, ok := doc["value"].(string); ok {
		return v, nil
	}
	return "", nil
}

func (s *service) BatchAdd(ctx context.Context, items []BatchAddItem) ([]BatchAddResult, int, int, error) {
	results := make([]BatchAddResult, 0, len(items))
	successCount := 0
	failedCount := 0
	updateTimes := make([]string, 0)

	type FingerprintYaml struct {
		Fingerprint models.FingerprintRule `yaml:"fingerprint"`
	}

	for _, item := range items {
		result := BatchAddResult{
			ID:      item.ID,
			Success: false,
		}

		var fp FingerprintYaml
		if err := yaml.Unmarshal([]byte(item.Content), &fp); err != nil {
			result.Error = fmt.Sprintf("Failed to parse YAML: %v", err)
			results = append(results, result)
			failedCount++
			continue
		}

		data := models.FingerprintRule{
			Name:           fp.Fingerprint.Name,
			Rule:           item.Content,
			FingerprintId:  item.ID,
			Category:       fp.Fingerprint.Category,
			ParentCategory: fp.Fingerprint.ParentCategory,
			State:          true,
			Company:        fp.Fingerprint.Company,
			Amount:         0,
		}

		filter := bson.M{"fingerprint_id": item.ID}
		if err := s.commonRepo.Upsert(ctx, "FingerprintRules", filter, data); err != nil {
			result.Error = fmt.Sprintf("Failed to update or insert: %v", err)
			results = append(results, result)
			failedCount++
			continue
		}

		result.Success = true
		results = append(results, result)
		successCount++

		if item.UpdateTime != "" {
			updateTimes = append(updateTimes, item.UpdateTime)
		}
	}

	if successCount > 0 {
		_ = s.nodeService.RefreshConfig(ctx, models.Message{Name: "all", Type: "finger"})
	}

	if len(updateTimes) > 0 {
		latestUpdateTime := updateTimes[0]

		for _, updateTime := range updateTimes {
			if updateTime > latestUpdateTime {
				latestUpdateTime = updateTime
			}
		}

		configFilter := bson.M{"name": "FingerVersion"}
		configDoc := bson.M{
			"name":  "FingerVersion",
			"value": latestUpdateTime,
			"type":  "finger",
		}
		if err := s.commonRepo.Upsert(ctx, "config", configFilter, configDoc); err != nil {

			_ = err
		}
	}

	return results, successCount, failedCount, nil
}
