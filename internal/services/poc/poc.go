package poc

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"recon/internal/logger"
	"recon/internal/services/node"

	"archive/zip"
	"io"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"recon/internal/models"
	"recon/internal/repositories/poc"
	"recon/internal/utils/helper"
	"recon/internal/utils/random"
)

type Service interface {
	GetPocList(ctx *gin.Context, req *models.PocListRequest) (*models.PocListResponse, error)
	GetPocDetail(ctx *gin.Context, req *models.PocDetailRequest) (*models.Poc, error)
	UpdatePoc(ctx *gin.Context, req *models.PocUpdateRequest) error
	AddPoc(ctx context.Context, req *models.PocAddRequest) (string, error)
	DeletePoc(ctx *gin.Context, req *models.PocDeleteRequest) error
	GetAllPocData(ctx context.Context) ([]models.Poc, error)
	ImportPoc(ctx context.Context, filePath string) (*models.PocImportResponse, error)
	TemplateIdExists(ctx context.Context, templateId string) (bool, error)
}

type service struct {
	repo        poc.Repository
	nodeService node.Service
}

func NewService() Service {
	return &service{
		repo:        poc.NewRepository(),
		nodeService: node.NewService(),
	}
}

func (s *service) GetPocList(ctx *gin.Context, req *models.PocListRequest) (*models.PocListResponse, error) {
	query := bson.M{
		"name": bson.M{
			"$regex":   req.Search,
			"$options": "i",
		},
	}

	if len(req.Filter.Level) > 0 {
		query["level"] = bson.M{"$in": req.Filter.Level}
	}

	opts := options.Find().
		SetSort(bson.D{{"time", -1}}).
		SetSkip(int64((req.PageIndex - 1) * req.PageSize)).
		SetLimit(int64(req.PageSize))

	results, err := s.repo.FindWithPagination(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get poc list: %w", err)
	}

	total, err := s.repo.Count(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to count poc: %w", err)
	}

	return &models.PocListResponse{
		List:  results,
		Total: total,
	}, nil
}

func (s *service) GetPocDetail(ctx *gin.Context, req *models.PocDetailRequest) (*models.Poc, error) {
	id, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid id format: %w", err)
	}

	result, err := s.repo.FindOne(ctx, bson.M{"_id": id})
	if err != nil {
		return nil, fmt.Errorf("failed to get poc detail: %w", err)
	}

	return result, nil
}

func (s *service) UpdatePoc(ctx *gin.Context, req *models.PocUpdateRequest) error {
	id, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		return fmt.Errorf("invalid id format: %w", err)
	}

	var pt models.PocTemplate
	err = yaml.Unmarshal([]byte(req.Content), &pt)
	if err != nil {
		return fmt.Errorf("failed to unmarshal poc template: %w", err)
	}
	ptTags := strings.Split(pt.Info.Tags, ",")
	tags := helper.RemoveArrayDuplicates(ptTags)
	update := bson.M{
		"$set": bson.M{
			"name":    pt.Info.Name,
			"content": req.Content,
			"level":   pt.Info.Severity,
			"tags":    tags,
		},
	}

	_, err = s.repo.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("failed to update poc: %w", err)
	}
	go func() {
		msg := models.Message{
			Name:    "all",
			Type:    "poc",
			Content: fmt.Sprintf(`add:%v`, req.ID),
		}
		err = s.nodeService.RefreshConfig(ctx, msg)
		if err != nil {
			logger.Error("failed to refresh config", zap.Error(err))
		}
	}()

	return nil
}

func (s *service) AddPoc(ctx context.Context, req *models.PocAddRequest) (string, error) {

	var pt models.PocTemplate
	err := yaml.Unmarshal([]byte(req.Content), &pt)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal poc template: %w", err)
	}
	id := pt.ID
	existing, err := s.repo.FindOne(ctx, bson.M{"id": id})
	if err == nil && existing != nil {
		return "", fmt.Errorf("poc already exists")
	}
	ptTags := strings.Split(pt.Info.Tags, ",")
	tags := helper.RemoveArrayDuplicates(ptTags)
	doc := models.Poc{
		Name:       pt.Info.Name,
		Content:    req.Content,
		Level:      pt.Info.Severity,
		TemplateId: id,
		Tags:       tags,
		Time:       helper.GetNowTimeString(),
	}

	res, err := s.repo.InsertOne(ctx, doc)
	if err != nil {
		return "", fmt.Errorf("failed to add poc: %w", err)
	}
	pocId := res.InsertedID.(primitive.ObjectID).Hex()
	go func() {
		msg := models.Message{
			Name:    "all",
			Type:    "poc",
			Content: fmt.Sprintf(`add:%v`, pocId),
		}
		err = s.nodeService.RefreshConfig(ctx, msg)
		if err != nil {
			logger.Error("failed to refresh config", zap.Error(err))
		}
	}()
	return pocId, nil
}

func (s *service) DeletePoc(ctx *gin.Context, req *models.PocDeleteRequest) error {
	var objIDs []primitive.ObjectID
	for _, id := range req.IDs {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return fmt.Errorf("invalid id format: %w", err)
		}
		objIDs = append(objIDs, objID)
	}

	_, err := s.repo.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": objIDs}})
	if err != nil {
		return fmt.Errorf("failed to delete poc: %w", err)
	}
	go func() {
		msg := models.Message{
			Name:    "all",
			Type:    "poc",
			Content: fmt.Sprintf(`delete:%v`, strings.Join(req.IDs, ",")),
		}
		err = s.nodeService.RefreshConfig(ctx, msg)
		if err != nil {
			logger.Error("failed to refresh config", zap.Error(err))
		}
	}()
	return nil
}

func (s *service) GetAllPocData(ctx context.Context) ([]models.Poc, error) {
	result, err := s.repo.GetAllPocData(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all poc data: %w", err)
	}
	return result, nil
}

func (s *service) ImportPoc(ctx context.Context, filePath string) (*models.PocImportResponse, error) {
	logger.Info("Starting POC import")

	fileName := random.GenerateString(11)

	tempDir := os.TempDir()

	unzipPath := filepath.Join(tempDir, fileName)
	extractPath := unzipPath

	if err := os.MkdirAll(extractPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create extract directory: %w", err)
	}

	yamlFiles, err := s.extractZipFile(filePath, extractPath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract zip file: %w", err)
	}

	templateIds, err := s.repo.GetAllTemplateId(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing poc hashes: %w", err)
	}

	successNum := 0
	errorNum := 0
	repeatNum := 0
	pocDataList := []models.Poc{}

	logger.Info(fmt.Sprintf("Total POCs: %d", len(yamlFiles)))

	for _, yamlFile := range yamlFiles {

		content, err := os.ReadFile(yamlFile)
		if err != nil {
			logger.Error(fmt.Sprintf("POC import: failed to read file: %s", yamlFile), zap.Error(err))
			errorNum++
			continue
		}

		var pt models.PocTemplate
		if err := yaml.Unmarshal(content, &pt); err != nil {
			logger.Error(fmt.Sprintf("POC import: failed to parse YAML: %s", yamlFile), zap.Error(err))
			errorNum++
			continue
		}
		if pt.Info.Name == "" || pt.ID == "" {
			errorNum++
			continue
		}

		if helper.StringInSlice(pt.ID, templateIds) {
			repeatNum++
			continue
		}
		templateIds = append(templateIds, pt.ID)

		severity := "unknown"
		if pt.Info.Severity != "" {
			severity = strings.ToLower(pt.Info.Severity)
		}

		tags := []string{}
		if pt.Info.Tags != "" {
			tagList := strings.Split(pt.Info.Tags, ",")
			for _, tag := range tagList {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					tags = append(tags, tag)
				}
			}
		}
		pocData := models.Poc{
			Name:       pt.Info.Name,
			Content:    string(content),
			TemplateId: pt.ID,
			Level:      severity,
			Time:       helper.GetNowTimeString(),
			Tags:       tags,
		}

		pocDataList = append(pocDataList, pocData)
	}

	if len(pocDataList) > 0 {
		result, err := s.repo.InsertMany(ctx, pocDataList)
		if err != nil {
			return nil, fmt.Errorf("failed to insert poc data: %w", err)
		}

		if result.InsertedIDs != nil {
			successNum = len(result.InsertedIDs)

			go func() {
				insertedIDs := make([]string, len(result.InsertedIDs))
				for i, id := range result.InsertedIDs {
					insertedIDs[i] = id.(primitive.ObjectID).Hex()
				}

				msg := models.Message{
					Name:    "all",
					Type:    "poc",
					Content: fmt.Sprintf("add:%s", strings.Join(insertedIDs, ",")),
				}

				if err := s.nodeService.RefreshConfig(ctx, msg); err != nil {
					logger.Error("failed to refresh config", zap.Error(err))
				}
			}()
		}
	}

	s.cleanupTempFiles(filePath, extractPath)

	logger.Info(fmt.Sprintf("POC import completed: %d succeeded, %d duplicate, %d failed", successNum, repeatNum, errorNum))
	logger.Info("POC import finished")

	return &models.PocImportResponse{
		SuccessNum: successNum,
		ErrorNum:   errorNum,
		RepeatNum:  repeatNum,
		Message:    "import success",
	}, nil
}

func (s *service) TemplateIdExists(ctx context.Context, templateId string) (bool, error) {
	return s.repo.TemplateIdExists(ctx, templateId)
}

func (s *service) extractZipFile(zipPath, extractPath string) ([]string, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var yamlFiles []string

	absExtractPath, err := filepath.Abs(extractPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	for _, file := range reader.File {

		cleanName := filepath.Clean(file.Name)
		if strings.HasPrefix(cleanName, "..") || strings.Contains(cleanName, ".."+string(os.PathSeparator)) {
			logger.Error("Path traversal detected in ZIP file", zap.String("path", file.Name))
			continue
		}

		filePath := filepath.Join(absExtractPath, file.Name)

		filePath = filepath.Clean(filePath)

		if !strings.HasPrefix(filePath, absExtractPath) {
			logger.Error("Path traversal detected in ZIP file", zap.String("path", file.Name), zap.String("resolved", filePath))
			continue
		}

		if strings.Contains(filePath, "..") {
			logger.Error("Path traversal detected after resolution", zap.String("path", file.Name), zap.String("resolved", filePath))
			continue
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, 0755); err != nil {
				logger.Error("Failed to create directory", zap.String("path", filePath), zap.Error(err))
				continue
			}
			continue
		}

		parentDir := filepath.Dir(filePath)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			logger.Error("Failed to create parent directory", zap.String("path", parentDir), zap.Error(err))
			continue
		}

		rc, err := file.Open()
		if err != nil {
			logger.Error("Failed to open file in ZIP", zap.String("path", file.Name), zap.Error(err))
			continue
		}

		dst, err := os.Create(filePath)
		if err != nil {
			rc.Close()
			logger.Error("Failed to create file", zap.String("path", filePath), zap.Error(err))
			continue
		}

		_, err = io.Copy(dst, rc)
		rc.Close()
		dst.Close()

		if err != nil {
			logger.Error("Failed to copy file content", zap.String("path", filePath), zap.Error(err))
			continue
		}

		if strings.HasSuffix(strings.ToLower(file.Name), ".yaml") || strings.HasSuffix(strings.ToLower(file.Name), ".yml") {
			yamlFiles = append(yamlFiles, filePath)
		}
	}

	return yamlFiles, nil
}

func (s *service) cleanupTempFiles(zipPath, extractPath string) {

	if err := os.Remove(zipPath); err != nil {
		logger.Error("Failed to delete POC ZIP file", zap.Error(err))
	}

	if err := os.RemoveAll(extractPath); err != nil {
		logger.Error("Failed to delete extracted POC directory", zap.Error(err))
	}
}

func (s *service) isSafeFilePath(filePath string) bool {

	if strings.Contains(filePath, "..") {
		return false
	}

	cleanPath := filepath.Clean(filePath)

	if strings.Contains(cleanPath, "..") {
		return false
	}

	// if filepath.IsAbs(cleanPath) {
	//     return false
	// }

	return true
}
