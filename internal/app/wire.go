//go:build wireinject
// +build wireinject

// Package app provides dependency injection setup using Wire.
// This file contains Wire provider definitions and injector functions.
// The actual implementation is generated in wire_gen.go by running:
//   wire gen ./internal/app
package app

import (
	"go-clean-arch/internal/adapter/gateway"
	"go-clean-arch/internal/adapter/repository"
	"go-clean-arch/internal/delivery/consumer"
	"go-clean-arch/internal/delivery/http"
	"go-clean-arch/internal/delivery/http/handler"
	"go-clean-arch/internal/delivery/job"
	"go-clean-arch/internal/usecase"
	"go-clean-arch/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/jmoiron/sqlx"
)

// ProvideStripeAPIKey provides the Stripe API key from config
func ProvideStripeAPIKey(cfg *config.Config) string {
	return cfg.Stripe.APIKey
}

// RepositorySet provides all repository dependencies
var RepositorySet = wire.NewSet(
	repository.NewUserRepository,
	repository.NewOrderRepository,
)

// GatewaySet provides all gateway dependencies
var GatewaySet = wire.NewSet(
	ProvideStripeAPIKey,
	gateway.NewStripeGateway,
)

// UseCaseSet provides all use case dependencies
var UseCaseSet = wire.NewSet(
	usecase.NewUserInteractor,
	usecase.NewOrderInteractor,
	wire.Bind(new(usecase.UserUsecase), new(*usecase.UserInteractor)),
	wire.Bind(new(usecase.OrderUsecase), new(*usecase.OrderInteractor)),
)

// HandlerSet provides all HTTP handler dependencies
var HandlerSet = wire.NewSet(
	handler.NewUserHandler,
)

// ConsumerSet provides all message consumer dependencies
var ConsumerSet = wire.NewSet(
	consumer.NewOrderConsumer,
)

// JobSet provides all job dependencies
var JobSet = wire.NewSet(
	job.NewDailyReportJob,
	job.NewScheduler,
)

// InitializeAPIRouter initializes the API router with all dependencies
func InitializeAPIRouter(db *sqlx.DB, cfg *config.Config) *gin.Engine {
	wire.Build(
		RepositorySet,
		GatewaySet,
		UseCaseSet,
		HandlerSet,
		http.Router,
	)
	return nil
}

// InitializeWorker initializes the worker with all dependencies
func InitializeWorker(db *sqlx.DB, cfg *config.Config) *consumer.OrderConsumer {
	wire.Build(
		RepositorySet,
		GatewaySet,
		UseCaseSet,
		ConsumerSet,
	)
	return nil
}

// InitializeCronScheduler initializes the cron scheduler with all dependencies
func InitializeCronScheduler(db *sqlx.DB, cfg *config.Config) *job.Scheduler {
	wire.Build(
		RepositorySet,
		GatewaySet,
		UseCaseSet,
		JobSet,
	)
	return nil
}
