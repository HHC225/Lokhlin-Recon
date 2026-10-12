package scheduler

import (
	"fmt"

	"recon/internal/models"
	schedulerCore "recon/internal/scheduler"

	schedulerRepo "recon/internal/repositories/task/scheduler"
	"recon/internal/utils/helper"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Service interface {
	GetScheduledData(ctx *gin.Context, search string, pageIndex, pageSize int) ([]models.Task, int64, error)
	CheckScheduledTaskNameExists(ctx *gin.Context, name string) (bool, error)
	CreateScheduledTask(ctx *gin.Context, req *models.Task) error
	UpdateScheduledTask(ctx *gin.Context, req *models.Task) error
	DeleteScheduledTasks(ctx *gin.Context, ids []string) error
	GetScheduledTaskDetail(ctx *gin.Context, id string) (*models.Task, error)
	UpdatePageMonitScheduledTask(ctx *gin.Context, req *models.Task) error
	GetPageMonitData(ctx *gin.Context, search string, pageIndex, pageSize int) ([]models.PageMonitoringTask, int64, error)
	AddPageMonitTask(ctx *gin.Context, url string) error
	DeletePageMonitTasks(ctx *gin.Context, ids []string) error
}

type service struct {
	schedulerRepo schedulerRepo.Repository
}

func NewService() Service {
	return &service{
		schedulerRepo: schedulerRepo.NewRepository(),
	}
}

func (s *service) GetScheduledData(ctx *gin.Context, search string, pageIndex, pageSize int) ([]models.Task, int64, error) {

	query := bson.M{}
	if search != "" {
		query["name"] = bson.M{"$regex": search, "$options": "i"}
	}

	total, err := s.schedulerRepo.Count(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSkip(int64((pageIndex - 1) * pageSize)).
		SetLimit(int64(pageSize))

	tasks, err := s.schedulerRepo.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}

	for i := range tasks {
		tasks[i].Cycle = s.calculateCycle(&tasks[i])
	}

	return tasks, total, nil
}

func (s *service) CheckScheduledTaskNameExists(ctx *gin.Context, name string) (bool, error) {
	query := bson.M{"name": name}
	count, err := s.schedulerRepo.Count(ctx, query)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *service) CreateScheduledTask(ctx *gin.Context, req *models.Task) error {

	if req.Type == "" {
		req.Type = "scan"
	}
	if req.TargetSource == "" {
		req.TargetSource = "general"
	}
	if req.CycleType == "" {
		req.CycleType = "nhours"
	}
	if req.Day == 0 {
		req.Day = 1
	}
	if req.Minute == 0 {
		req.Minute = 0
	}
	if req.Hour == 0 {
		req.Hour = 1
	}
	if req.Week == 0 {
		req.Week = 1
	}
	req.ID = primitive.NewObjectID()
	req.TaskID = req.ID.Hex()

	if err := s.schedulerRepo.Insert(ctx, req); err != nil {
		return err
	}

	if req.ScheduledTasks {
		if err := s.addScheduler(req, "scan"); err != nil {
			return err
		}
	}

	return nil
}

func (s *service) UpdateScheduledTask(ctx *gin.Context, req *models.Task) error {

	existingTask, err := s.GetScheduledTaskDetail(ctx, req.ID.Hex())
	if err != nil {
		return err
	}
	if existingTask == nil {
		return fmt.Errorf("scheduled task not found")
	}

	if req.Name != "" {
		existingTask.Name = req.Name
	}
	if req.Type != "" {
		existingTask.Type = req.Type
	}
	if len(req.Node) > 0 {
		existingTask.Node = req.Node
	}
	existingTask.AllNode = req.AllNode
	if len(req.Project) > 0 {
		existingTask.Project = req.Project
	}
	if req.TargetSource != "" {
		existingTask.TargetSource = req.TargetSource
	}
	if req.Day > 0 {
		existingTask.Day = req.Day
	}
	if req.Minute >= 0 {
		existingTask.Minute = req.Minute
	}
	if req.Hour >= 0 {
		existingTask.Hour = req.Hour
	}
	if req.Search != "" {
		existingTask.Search = req.Search
	}
	if req.CycleType != "" {
		existingTask.CycleType = req.CycleType
	}
	existingTask.ScheduledTasks = req.ScheduledTasks
	if req.Target != "" {
		existingTask.Target = req.Target
	}
	if req.Ignore != "" {
		existingTask.Ignore = req.Ignore
	}
	if req.Template != "" {
		existingTask.Template = req.Template
	}
	if req.Duplicates != "" {
		existingTask.Duplicates = req.Duplicates
	}
	if req.Week > 0 {
		existingTask.Week = req.Week
	}

	filter := bson.M{"_id": req.ID}
	if err := s.schedulerRepo.UpdateOne(ctx, filter, existingTask); err != nil {
		return err
	}

	if req.ScheduledTasks {

		schedulerCore.GetGlobalScheduler().RemoveJob(req.ID.Hex())

		if err := s.addScheduler(existingTask, "scan"); err != nil {
			return err
		}
	} else {

		schedulerCore.GetGlobalScheduler().RemoveJob(req.ID.Hex())
	}

	return nil
}

func (s *service) DeleteScheduledTasks(ctx *gin.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	filter := bson.M{
		"$or": []bson.M{
			{"id": bson.M{"$in": ids}},
		},
	}

	var objectIds []primitive.ObjectID
	for _, id := range ids {
		if id != "page_monitoring" {
			if objID, err := primitive.ObjectIDFromHex(id); err == nil {
				objectIds = append(objectIds, objID)
			}
		}
	}

	if len(objectIds) > 0 {
		filter["$or"] = append(filter["$or"].([]bson.M), bson.M{"_id": bson.M{"$in": objectIds}})
	}

	for _, id := range ids {
		schedulerCore.GetGlobalScheduler().RemoveJob(id)
	}

	if err := s.schedulerRepo.DeleteMany(ctx, filter); err != nil {
		return err
	}

	return nil
}

func (s *service) GetScheduledTaskDetail(ctx *gin.Context, id string) (*models.Task, error) {

	if objID, err := primitive.ObjectIDFromHex(id); err == nil {
		filter := bson.M{"_id": objID}
		task, err := s.schedulerRepo.FindOne(ctx, filter)
		if err == nil && task != nil {
			return task, nil
		}
	}

	filter := bson.M{"id": id}
	task, err := s.schedulerRepo.FindOne(ctx, filter)
	if err == nil && task != nil {
		return task, nil
	}

	return nil, fmt.Errorf("scheduled task not found")
}

func (s *service) addScheduler(task *models.Task, tp string) error {

	job := &schedulerCore.Job{
		ID:      task.ID.Hex(),
		Name:    task.Name,
		Handler: tp,
		Params:  task.ID.Hex(),
	}

	switch task.CycleType {
	case "daily":
		job.CycleType = "cron"
		job.Schedule = fmt.Sprintf("%d %d * * *", task.Minute, task.Hour)
	case "ndays":
		job.CycleType = "interval"
		job.Interval.Days = task.Day
		job.Interval.Hours = task.Hour
		job.Interval.Minutes = task.Minute
	case "nhours":
		job.CycleType = "cron"

		if task.Hour <= 0 && task.Minute <= 0 {
			task.Hour = 1
			task.Minute = 0
		}

		if task.Minute < 0 || task.Minute > 59 {
			task.Minute = 0
		}

		if task.Hour > 0 {

			job.Schedule = fmt.Sprintf("%d 0-23/%d * * *", task.Minute, task.Hour)
		} else {

			job.Schedule = fmt.Sprintf("%d * * * *", task.Minute)
		}
	case "weekly":
		job.CycleType = "cron"
		job.Schedule = fmt.Sprintf("%d %d * * %d", task.Minute, task.Hour, task.Week)
	case "monthly":
		job.CycleType = "cron"
		job.Schedule = fmt.Sprintf("%d %d %d * *", task.Minute, task.Hour, task.Day)
	default:
		return fmt.Errorf("unsupported cycle type: %s", task.CycleType)
	}

	if err := schedulerCore.GetGlobalScheduler().AddJob(job); err != nil {
		return fmt.Errorf("failed to add scheduled task: %w", err)
	}

	return nil
}

func (s *service) calculateCycle(task *models.Task) string {
	week := task.Week
	day := task.Day
	hour := task.Hour
	minute := task.Minute
	cycleType := task.CycleType

	if task.TaskID == "page_monitoring" {
		return fmt.Sprintf("%d hour", hour)
	}

	switch cycleType {
	case "daily":
		return fmt.Sprintf("Every day at %d:%02d", hour, minute)
	case "ndays":
		return fmt.Sprintf("Every %d days at %d:%02d", day, hour, minute)
	case "nhours":
		return fmt.Sprintf("Every %dh %dm", hour, minute)
	case "weekly":
		return fmt.Sprintf("Every week on day %d at %d:%02d", week, hour, minute)
	case "monthly":
		return fmt.Sprintf("Every month on day %d at %d:%02d", day, hour, minute)
	default:
		return "Unknown cycle"
	}
}

func (s *service) UpdatePageMonitScheduledTask(ctx *gin.Context, req *models.Task) error {

	filter := bson.M{"id": "page_monitoring"}
	existingTask, err := s.schedulerRepo.FindOne(ctx, filter)
	if err != nil {
		return err
	}
	if existingTask == nil {
		return fmt.Errorf("scheduled task not found")
	}
	existingTask.CycleType = "nhours"
	existingTask.AllNode = req.AllNode
	existingTask.ScheduledTasks = req.ScheduledTasks
	existingTask.Hour = req.Hour
	existingTask.ScheduledTasks = req.ScheduledTasks
	filter = bson.M{"_id": existingTask.ID}

	if err := s.schedulerRepo.UpdateOne(ctx, filter, existingTask); err != nil {
		return err
	}
	if existingTask.ScheduledTasks {

		schedulerCore.GetGlobalScheduler().RemoveJob(existingTask.ID.Hex())

		if err := s.addScheduler(existingTask, "page_monitoring"); err != nil {
			return err
		}
	} else {

		schedulerCore.GetGlobalScheduler().RemoveJob(existingTask.ID.Hex())
	}

	return nil
}

func (s *service) GetPageMonitData(ctx *gin.Context, search string, pageIndex, pageSize int) ([]models.PageMonitoringTask, int64, error) {

	query := bson.M{}
	if search != "" {
		query["url"] = bson.M{"$regex": search, "$options": "i"}
	}

	total, err := s.schedulerRepo.CountPageMonit(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSkip(int64((pageIndex - 1) * pageSize)).
		SetLimit(int64(pageSize))

	tasks, err := s.schedulerRepo.FindPageMonit(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

func (s *service) AddPageMonitTask(ctx *gin.Context, url string) error {

	task := &models.PageMonitoringTask{
		ID:      primitive.NewObjectID(),
		URL:     url,
		Hash:    []string{},
		MD5:     calculateMD5(url),
		State:   1,
		Project: "",
		Time:    "",
	}

	if err := s.schedulerRepo.InsertPageMonit(ctx, task); err != nil {
		return err
	}

	return nil
}

func (s *service) DeletePageMonitTasks(ctx *gin.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	var objectIds []primitive.ObjectID
	for _, id := range ids {
		if objID, err := primitive.ObjectIDFromHex(id); err == nil {
			objectIds = append(objectIds, objID)
		}
	}

	if len(objectIds) == 0 {
		return fmt.Errorf("invalid ids")
	}

	filter := bson.M{"_id": bson.M{"$in": objectIds}}

	if err := s.schedulerRepo.DeletePageMonit(ctx, filter); err != nil {
		return err
	}

	return nil
}

func calculateMD5(url string) string {
	return helper.CalculateMD5FromContent(url)
}
