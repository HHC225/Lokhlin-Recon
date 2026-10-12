package export

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"recon/internal/config"
	"recon/internal/logger"

	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"recon/internal/models"
	"recon/internal/repositories/export"
)

type Service interface {
	CreateExportTask(ctx context.Context, task *models.ExportTask) error
	ProcessExportTask(ctx context.Context, run models.ExportRun)
	GetExportFields(ctx context.Context, index string) ([]string, error)
	GetExportRecords(ctx context.Context) ([]models.ExportTask, error)
	DeleteExportFiles(ctx context.Context, fileNames []string) error
	IsValidFileName(fileName string) bool
}

type service struct {
	repository export.Repository
}

func NewService() Service {
	return &service{
		repository: export.NewRepository(),
	}
}

func (s *service) CreateExportTask(ctx context.Context, task *models.ExportTask) error {
	return s.repository.CreateExportTask(ctx, task)
}

func (s *service) ProcessExportTask(ctx context.Context, run models.ExportRun) {

	opts := options.Find().
		SetLimit(int64(run.Quantity))

	projection := bson.M{}
	for _, field := range run.Field {
		projection[field] = 1
	}
	if run.Index == "asset" {
		projection["type"] = 1
	}
	opts.SetProjection(projection)

	cursor, err := s.repository.GetCollectionCursor(
		ctx,
		run.Index,
		run.Filter,
		opts,
	)
	if err != nil {
		err := s.repository.UpdateExportTaskStatus(ctx, run.FileName, 2, 0)
		if err != nil {
			logger.Error(fmt.Sprintf("Error updating export task status: %s", err.Error()))
			return
		}
		return
	}
	defer cursor.Close(ctx)

	filePath := filepath.Join(config.GlobalConfig.System.ExeDir, "files", "export", run.FileName)
	var fileSize float64
	if run.FileType == "xlsx" {
		fileSize, err = s.exportToXlsx(cursor, filePath, run.Field, run.Index)
	} else {
		fileSize, err = s.exportToJSON(cursor, filePath)
	}

	if err != nil {
		logger.Error(fmt.Sprintf("Error export: %s", err.Error()))
		err := s.repository.UpdateExportTaskStatus(ctx, run.FileName, 2, 0)
		if err != nil {
			logger.Error(fmt.Sprintf("Error updating export task status: %s", err.Error()))
			return
		}
		return
	}

	err = s.repository.UpdateExportTaskStatus(ctx, run.FileName, 1, fileSize)
	if err != nil {
		logger.Error(fmt.Sprintf("Error updating export task status: %s", err.Error()))
		return
	}
}

func (s *service) GetExportFields(ctx context.Context, index string) ([]string, error) {

	fields, ok := models.Field[index]
	if !ok {
		return []string{}, fmt.Errorf("not found: %s", index)
	}
	return fields, nil
}

func (s *service) GetExportRecords(ctx context.Context) ([]models.ExportTask, error) {

	opts := options.Find().
		SetSort(bson.D{{"create_time", -1}}).
		SetProjection(bson.D{
			{"_id", 0},
			{"id", bson.D{{"$toString", "$_id"}}},
			{"file_name", 1},
			{"end_time", 1},
			{"create_time", 1},
			{"data_type", 1},
			{"state", 1},
			{"file_size", 1},
			{"file_type", 1},
		})

	return s.repository.GetExportRecords(ctx, bson.M{}, opts)
}

func (s *service) DeleteExportFiles(ctx context.Context, fileNames []string) error {

	for _, fileName := range fileNames {
		if s.IsValidFileName(fileName) {
			filePath := filepath.Join(config.GlobalConfig.System.ExeDir, "files", "export", fileName)
			if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
				logger.Error(fmt.Sprintf("Failed to delete file %s: %v", fileName, err))
				continue
			}
		}
	}

	return s.repository.DeleteExportTasks(ctx, fileNames)
}

func (s *service) IsValidFileName(fileName string) bool {
	parts := strings.Split(fileName, ".")
	if len(parts) != 2 {
		return false
	}

	name := parts[0]
	fileType := parts[1]
	if fileType != "json" && fileType != "xlsx" {
		return false
	}
	matched, err := regexp.MatchString(`^[a-z0-9-]+$`, name)
	if err != nil {
		return false
	}
	return matched
}

func getCellName(col, row int) string {
	colName := ""
	for col > 0 {
		col--
		colName = string(rune('A'+col%26)) + colName
		col /= 26
	}
	if colName == "" {
		colName = "A"
	}
	return fmt.Sprintf("%s%d", colName, row)
}

func sanitizeExcelValue(value interface{}) string {
	if value == nil {
		return ""
	}

	str := fmt.Sprintf("%v", value)

	result := make([]rune, 0, len(str))
	for _, r := range str {

		if r >= 0x20 || r == 0x09 || r == 0x0A || r == 0x0D {
			result = append(result, r)
		} else {

		}
	}
	return string(result)
}

func (s *service) exportToXlsx(cursor *mongo.Cursor, filePath string, fields []string, index string) (float64, error) {

	f := excelize.NewFile()
	defer f.Close()
	start := time.Now()

	const flushInterval = 1000

	if index == "asset" {

		httpFields := []string{}
		otherFields := []string{}
		for _, field := range fields {
			if models.IsHTTPAssetField(field) {
				httpFields = append(httpFields, field)
			}
			if models.IsOtherAssetField(field) {
				otherFields = append(otherFields, field)
			}
		}

		httpSheet := "HTTP Data"
		err := f.SetSheetName("Sheet1", httpSheet)
		if err != nil {
			logger.Error(fmt.Sprintf("Error setting sheet name: %s", err.Error()))
			return 0, err
		}

		otherSheet := "Other Data"
		_, err = f.NewSheet(otherSheet)
		if err != nil {
			logger.Error(fmt.Sprintf("Error creating sheet: %s", err.Error()))
			return 0, err
		}

		httpStreamWriter, err := f.NewStreamWriter(httpSheet)
		if err != nil {
			return 0, fmt.Errorf("failed to create stream writer for HTTP sheet: %w", err)
		}

		otherStreamWriter, err := f.NewStreamWriter(otherSheet)
		if err != nil {
			return 0, fmt.Errorf("failed to create stream writer for Other sheet: %w", err)
		}

		httpHeader := make([]interface{}, len(httpFields))
		for i, field := range httpFields {
			httpHeader[i] = field
		}
		if err := httpStreamWriter.SetRow("A1", httpHeader); err != nil {
			return 0, fmt.Errorf("failed to write HTTP header: %w", err)
		}

		otherHeader := make([]interface{}, len(otherFields))
		for i, field := range otherFields {
			otherHeader[i] = field
		}
		if err := otherStreamWriter.SetRow("A1", otherHeader); err != nil {
			return 0, fmt.Errorf("failed to write Other header: %w", err)
		}

		httpRowIdx := 2
		otherRowIdx := 2
		httpCount := 0
		otherCount := 0

		for cursor.Next(context.Background()) {
			var doc bson.M
			if err := cursor.Decode(&doc); err != nil {
				logger.Error(fmt.Sprintf("Failed to decode document: %v, skipping...", err))
				continue
			}

			if doc["type"] == "http" {

				rowData := make([]interface{}, len(httpFields))
				for i, field := range httpFields {
					value := doc[field]
					rowData[i] = sanitizeExcelValue(value)
				}

				cell, err := excelize.CoordinatesToCellName(1, httpRowIdx)
				if err != nil {
					logger.Error(fmt.Sprintf("Failed to convert coordinates for HTTP row %d: %v", httpRowIdx, err))
					httpRowIdx++
					continue
				}

				if err := httpStreamWriter.SetRow(cell, rowData); err != nil {
					logger.Error(fmt.Sprintf("Failed to write HTTP row %d: %v", httpRowIdx, err))

					emptyRow := make([]interface{}, len(httpFields))
					for i := range emptyRow {
						emptyRow[i] = ""
					}
					if retryErr := httpStreamWriter.SetRow(cell, emptyRow); retryErr != nil {
						return 0, fmt.Errorf("excelize stream writer corrupted at HTTP row %d: %w", httpRowIdx, retryErr)
					}
				} else {
					httpCount++
				}
				httpRowIdx++
			} else {

				rowData := make([]interface{}, len(otherFields))
				for i, field := range otherFields {
					value := doc[field]
					rowData[i] = sanitizeExcelValue(value)
				}

				cell, err := excelize.CoordinatesToCellName(1, otherRowIdx)
				if err != nil {
					logger.Error(fmt.Sprintf("Failed to convert coordinates for Other row %d: %v", otherRowIdx, err))
					otherRowIdx++
					continue
				}

				if err := otherStreamWriter.SetRow(cell, rowData); err != nil {
					logger.Error(fmt.Sprintf("Failed to write Other row %d: %v", otherRowIdx, err))

					emptyRow := make([]interface{}, len(otherFields))
					for i := range emptyRow {
						emptyRow[i] = ""
					}
					if retryErr := otherStreamWriter.SetRow(cell, emptyRow); retryErr != nil {
						return 0, fmt.Errorf("excelize stream writer corrupted at Other row %d: %w", otherRowIdx, retryErr)
					}
				} else {
					otherCount++
				}
				otherRowIdx++
			}

		}

		if err := cursor.Err(); err != nil {
			return 0, fmt.Errorf("cursor error: %w", err)
		}

		if err := httpStreamWriter.Flush(); err != nil {
			return 0, fmt.Errorf("failed to flush HTTP stream: %w", err)
		}
		if err := otherStreamWriter.Flush(); err != nil {
			return 0, fmt.Errorf("failed to flush Other stream: %w", err)
		}

		if err := f.SaveAs(filePath); err != nil {
			return 0, fmt.Errorf("failed to save excel file: %w", err)
		}

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			return 0, fmt.Errorf("failed to get file info: %w", err)
		}
		fileSize := float64(fileInfo.Size()) / (1024 * 1024)
		end := time.Now()
		totalRows := httpCount + otherCount
		logger.Info(fmt.Sprintf("Saved %v excel file size: %.2f MB time: %v number: %v (HTTP: %v, Other: %v)", index, fileSize, end.Sub(start), totalRows, httpCount, otherCount))
		return fileSize, nil
	} else {

		sheet := "Sheet1"

		streamWriter, err := f.NewStreamWriter(sheet)
		if err != nil {
			return 0, fmt.Errorf("failed to create stream writer: %w", err)
		}

		header := make([]interface{}, len(fields))
		for i, field := range fields {
			header[i] = field
		}
		if err := streamWriter.SetRow("A1", header); err != nil {
			return 0, fmt.Errorf("failed to write header: %w", err)
		}

		row := 2
		rowCount := 0

		for cursor.Next(context.Background()) {
			var doc bson.M
			if err := cursor.Decode(&doc); err != nil {
				logger.Error(fmt.Sprintf("Failed to decode document: %v, skipping...", err))
				continue
			}

			rowData := make([]interface{}, len(fields))
			for i, field := range fields {
				value := doc[field]
				rowData[i] = sanitizeExcelValue(value)
			}

			cell, err := excelize.CoordinatesToCellName(1, row)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to convert coordinates for row %d: %v", row, err))
				row++
				continue
			}

			if err := streamWriter.SetRow(cell, rowData); err != nil {
				logger.Error(fmt.Sprintf("Failed to write row %d: %v", row, err))

				emptyRow := make([]interface{}, len(fields))
				for i := range emptyRow {
					emptyRow[i] = ""
				}
				if retryErr := streamWriter.SetRow(cell, emptyRow); retryErr != nil {
					return 0, fmt.Errorf("excelize stream writer corrupted at row %d: %w", row, retryErr)
				}
			} else {
				rowCount++
			}
			row++
		}

		if err := cursor.Err(); err != nil {
			return 0, fmt.Errorf("cursor error: %w", err)
		}

		if err := streamWriter.Flush(); err != nil {
			return 0, fmt.Errorf("failed to flush stream: %w", err)
		}

		if err := f.SaveAs(filePath); err != nil {
			return 0, fmt.Errorf("failed to save excel file: %w", err)
		}

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			return 0, fmt.Errorf("failed to get file info: %w", err)
		}
		fileSize := float64(fileInfo.Size()) / (1024 * 1024)
		end := time.Now()
		logger.Info(fmt.Sprintf("Saved %v excel file size: %.2f MB time: %v number: %v", index, fileSize, end.Sub(start), rowCount))
		return fileSize, nil
	}
}

func (s *service) exportToJSON(cursor *mongo.Cursor, filePath string) (float64, error) {
	start := time.Now()
	file, err := os.Create(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetEscapeHTML(false)

	row := 0
	for cursor.Next(context.Background()) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			logger.Error(fmt.Sprintf("Failed to decode document: %v, skipping...", err))
			continue
		}

		if err := encoder.Encode(doc); err != nil {
			logger.Error(fmt.Sprintf("Failed to encode document: %v, skipping...", err))
			continue
		}
		row++
	}

	if err := cursor.Err(); err != nil {
		return 0, fmt.Errorf("cursor error: %w", err)
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return 0, fmt.Errorf("failed to get file info: %w", err)
	}

	fileSize := float64(fileInfo.Size()) / (1024 * 1024)
	end := time.Now()
	logger.Info(fmt.Sprintf("Saved JSON file size: %.2f MB time: %v number: %v", fileSize, end.Sub(start), row))
	return fileSize, nil
}
