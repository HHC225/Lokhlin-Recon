package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"recon/internal/config"
	"recon/internal/database/mongodb"
	"recon/internal/logger"
	"recon/internal/scheduler/jobs"
	"recon/internal/utils/helper"

	"github.com/go-co-op/gocron"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type JobHandler func(params string, nextTime string) error

type Job struct {
	ID          string    `bson:"_id"`
	Name        string    `bson:"name"`
	Schedule    string    `bson:"schedule"`
	Handler     string    `bson:"handler"`
	Params      string    `bson:"params"`
	LastRunTime time.Time `bson:"last_run_time"`
	NextRunTime time.Time `bson:"next_run_time"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
	CycleType   string    `bson:"cycle_type"`
	Interval    struct {
		Days    int `bson:"days"`
		Hours   int `bson:"hours"`
		Minutes int `bson:"minutes"`
	} `bson:"interval"`
}

type Scheduler struct {
	scheduler      *gocron.Scheduler
	db             *mongo.Database
	jobsCollection *mongo.Collection
	logger         *zap.Logger
	jobs           map[string]*gocron.Job
	handlers       map[string]JobHandler
}

var (
	globalScheduler *Scheduler
	once            sync.Once
)

func GetGlobalScheduler() *Scheduler {
	return globalScheduler
}

func InitializeGlobalScheduler() {
	once.Do(func() {
		globalScheduler = NewScheduler()
		globalScheduler.RegisterHandler("scan", func(params string, nextTime string) error {
			return jobs.Scan(params, nextTime)
		})
		globalScheduler.RegisterHandler("page_monitoring", func(params string, nextTime string) error {
			return jobs.CreatePageMonitoringTask()
		})
		globalScheduler.RegisterHandler("server_plugin", func(params string, nextTime string) error {
			return jobs.ServerPluginRunner(params, nextTime)
		})
	})
}

func NewScheduler() *Scheduler {
	loc, err := time.LoadLocation(config.GlobalConfig.System.Timezone)
	if err != nil {

		loc = time.Local
	}
	return &Scheduler{
		scheduler:      gocron.NewScheduler(loc),
		db:             mongodb.DB,
		jobsCollection: mongodb.DB.Collection("scheduled"),
		logger:         logger.Logger,
		jobs:           make(map[string]*gocron.Job),
		handlers:       make(map[string]JobHandler),
	}
}

func (s *Scheduler) RegisterHandler(name string, handler JobHandler) {
	s.handlers[name] = handler
	s.logger.Info("Registered job handler", zap.String("handler_name", name))
}

func (s *Scheduler) Start() {
	s.scheduler.StartAsync()
	s.loadJobs()
}

func (s *Scheduler) Stop() {
	s.scheduler.Stop()
}

func (s *Scheduler) AddJob(job *Job) error {

	if _, exists := s.handlers[job.Handler]; !exists {
		return fmt.Errorf("handler %s not registered", job.Handler)
	}

	if _, exists := s.jobs[job.ID]; exists {

		return nil
	}

	var existingJob Job
	err := s.jobsCollection.FindOne(context.Background(), bson.M{"_id": job.ID}).Decode(&existingJob)
	if err == nil {

		return nil
	} else if err != mongo.ErrNoDocuments {

		return fmt.Errorf("failed to check if job exists: %w", err)
	}

	job.CreatedAt = helper.GetNowTime()
	job.UpdatedAt = helper.GetNowTime()

	_, err = s.jobsCollection.InsertOne(context.Background(), job)
	if err != nil {
		return err
	}

	var schedJob *gocron.Job
	var err2 error

	if job.CycleType == "interval" {

		schedJob, err2 = s.scheduler.Every(job.Interval.Days).Days().
			At(fmt.Sprintf("%02d:%02d", job.Interval.Hours, job.Interval.Minutes)).
			Do(s.executeJob, job)
	} else {

		schedJob, err2 = s.scheduler.Cron(job.Schedule).Do(s.executeJob, job)
	}

	if err2 != nil {
		return err2
	}

	s.jobs[job.ID] = schedJob

	s.logger.Info("Successfully added new job",
		zap.String("job_id", job.ID),
		zap.String("job_name", job.Name),
		zap.String("cycle_type", job.CycleType),
		zap.Time("next_run_time", schedJob.NextRun()),
	)

	return nil
}

func (s *Scheduler) RemoveJob(jobID string) error {

	if job, exists := s.jobs[jobID]; exists {
		s.scheduler.RemoveByReference(job)
		delete(s.jobs, jobID)
	}

	_, err := s.jobsCollection.DeleteOne(context.Background(), bson.M{"_id": jobID})
	if err != nil {
		return err
	}
	s.logger.Info("Successfully remove job")
	return nil
}

func (s *Scheduler) UpdateJob(job *Job) error {

	if _, exists := s.handlers[job.Handler]; !exists {
		return fmt.Errorf("handler %s not registered", job.Handler)
	}

	job.UpdatedAt = helper.GetNowTime()

	_, err := s.jobsCollection.ReplaceOne(context.Background(), bson.M{"_id": job.ID}, job)
	if err != nil {
		return err
	}

	if oldJob, exists := s.jobs[job.ID]; exists {
		s.scheduler.RemoveByReference(oldJob)
	}

	schedJob, err := s.scheduler.Cron(job.Schedule).Do(s.executeJob, job)
	if err != nil {
		return err
	}

	s.jobs[job.ID] = schedJob
	return nil
}

func (s *Scheduler) loadJobs() {
	cursor, err := s.jobsCollection.Find(context.Background(), bson.M{})
	if err != nil {
		s.logger.Error("Failed to load jobs from MongoDB", zap.Error(err))
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var job Job
		if err := cursor.Decode(&job); err != nil {
			s.logger.Error("Failed to decode job", zap.Error(err))
			continue
		}

		if _, exists := s.handlers[job.Handler]; !exists {
			s.logger.Error("Handler not registered for job",
				zap.String("job_id", job.ID),
				zap.String("handler", job.Handler),
			)
			continue
		}

		schedJob, err := s.scheduler.Cron(job.Schedule).Do(s.executeJob, &job)
		if err != nil {
			s.logger.Error("Failed to schedule job", zap.Error(err))
			continue
		}
		s.jobs[job.ID] = schedJob
	}
}

func (s *Scheduler) executeJob(job *Job) {
	s.logger.Info("Executing job",
		zap.String("job_id", job.ID),
		zap.String("job_name", job.Name),
		zap.String("handler", job.Handler),
	)

	handler, exists := s.handlers[job.Handler]
	if !exists {
		s.logger.Error("Handler not found",
			zap.String("job_id", job.ID),
			zap.String("handler", job.Handler),
		)
		return
	}

	job.LastRunTime = helper.GetNowTime()
	job.NextRunTime = s.jobs[job.ID].NextRun()

	if err := handler(job.Params, helper.FormatTime(job.NextRunTime)); err != nil {
		s.logger.Error("Job execution failed",
			zap.String("job_id", job.ID),
			zap.String("job_name", job.Name),
			zap.Error(err),
		)
	}

	_, err := s.jobsCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": job.ID},
		bson.M{
			"$set": bson.M{
				"last_run_time": job.LastRunTime,
				"next_run_time": job.NextRunTime,
			},
		},
	)
	if err != nil {
		s.logger.Error("Failed to update job execution time", zap.Error(err))
	}
}
