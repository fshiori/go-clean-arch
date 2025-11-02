package job

import (
	"context"
	"go-clean-arch/internal/usecase"
	"go-clean-arch/pkg/logger"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

// DailyReportJob handles daily report generation
type DailyReportJob struct {
	userInteractor  usecase.UserUsecase
	orderInteractor usecase.OrderUsecase
}

// NewDailyReportJob creates a new DailyReportJob
func NewDailyReportJob(
	userInteractor usecase.UserUsecase,
	orderInteractor usecase.OrderUsecase,
) *DailyReportJob {
	return &DailyReportJob{
		userInteractor:  userInteractor,
		orderInteractor: orderInteractor,
	}
}

// Run executes the daily report job
func (j *DailyReportJob) Run() error {
	// Create context with trace ID for this job execution
	ctx := context.Background()
	traceID := uuid.New().String()
	ctx = logger.WithTraceID(ctx, traceID)

	logger.InfoContext(ctx, "Starting daily report generation")

	// This is a simplified example
	// In production, you would:
	// 1. Query users and orders for the day
	// 2. Generate statistics
	// 3. Send email or save report

	// Example: Get some statistics
	users, err := j.userInteractor.ListUsers(1, 100)
	if err != nil {
		logger.ErrorContext(ctx, "Error fetching users for report", "error", err)
		return err
	}

	reportTime := time.Now()
	logger.InfoContext(ctx, "Daily report generated",
		"total_users", len(users),
		"report_time", reportTime.Format(time.RFC3339),
	)

	// TODO: Implement actual report generation and distribution
	// - Aggregate data
	// - Create report file (PDF, CSV, etc.)
	// - Send email to stakeholders
	// - Store in cloud storage

	logger.InfoContext(ctx, "Daily report completed successfully")
	return nil
}

// Scheduler manages all cron jobs
type Scheduler struct {
	cron            *cron.Cron
	dailyReportJob  *DailyReportJob
}

// NewScheduler creates a new job scheduler
func NewScheduler(dailyReportJob *DailyReportJob) *Scheduler {
	return &Scheduler{
		cron:           cron.New(),
		dailyReportJob: dailyReportJob,
	}
}

// Start starts all scheduled jobs
func (s *Scheduler) Start() error {
	logger.Info("Starting cron scheduler...")

	// Schedule daily report at midnight
	_, err := s.cron.AddFunc("0 0 * * *", func() {
		logger.Info("Triggering daily report job")
		if err := s.dailyReportJob.Run(); err != nil {
			logger.Error("Error running daily report", "error", err)
		}
	})
	if err != nil {
		logger.Error("Failed to schedule daily report job", "error", err)
		return err
	}

	// Add more scheduled jobs here
	// Example: Every hour
	// _, err = s.cron.AddFunc("0 * * * *", func() {
	//     logger.Info("Triggering hourly job")
	//     if err := s.someOtherJob.Run(); err != nil {
	//         logger.Error("Error running hourly job", "error", err)
	//     }
	// })

	s.cron.Start()
	logger.Info("Cron scheduler started successfully", "jobs_count", len(s.cron.Entries()))

	// Block forever
	select {}
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	logger.Info("Stopping cron scheduler...")
	s.cron.Stop()
	logger.Info("Cron scheduler stopped")
}
