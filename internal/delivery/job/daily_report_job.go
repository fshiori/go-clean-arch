package job

import (
	"go-clean-arch/internal/usecase"
	"log"
	"time"

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
	log.Println("Starting daily report generation...")

	// This is a simplified example
	// In production, you would:
	// 1. Query users and orders for the day
	// 2. Generate statistics
	// 3. Send email or save report

	// Example: Get some statistics
	users, err := j.userInteractor.ListUsers(1, 100)
	if err != nil {
		log.Printf("Error fetching users: %v", err)
		return err
	}

	log.Printf("Daily Report - Total users: %d", len(users))
	log.Printf("Report generated at: %s", time.Now().Format(time.RFC3339))

	// TODO: Implement actual report generation and distribution
	// - Aggregate data
	// - Create report file (PDF, CSV, etc.)
	// - Send email to stakeholders
	// - Store in cloud storage

	log.Println("Daily report completed successfully")
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
	log.Println("Starting cron scheduler...")

	// Schedule daily report at midnight
	_, err := s.cron.AddFunc("0 0 * * *", func() {
		if err := s.dailyReportJob.Run(); err != nil {
			log.Printf("Error running daily report: %v", err)
		}
	})
	if err != nil {
		return err
	}

	// Add more scheduled jobs here
	// Example: Every hour
	// s.cron.AddFunc("0 * * * *", s.someOtherJob.Run)

	s.cron.Start()
	log.Println("Cron scheduler started successfully")

	// Block forever
	select {}
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	log.Println("Stopping cron scheduler...")
	s.cron.Stop()
}
