//go:build wireinject
// +build wireinject

// Package app provides dependency injection setup using Wire.
// This file contains Wire provider definitions and injector functions.
// The actual implementation is generated in wire_gen.go by running:
//
//	wire gen ./internal/app
package app

import (
	"go-clean-arch/internal/adapter/gateway"
	"go-clean-arch/internal/adapter/repository"
	"go-clean-arch/internal/delivery/consumer"
	"go-clean-arch/internal/delivery/http"
	"go-clean-arch/internal/delivery/http/handler"
	"go-clean-arch/internal/delivery/job"
	microhandler "go-clean-arch/internal/delivery/micro/handler"
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
	handler.NewOrderHandler,
	handler.NewHealthHandler,
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

// MicroHandlerSet provides all microservice handler dependencies
var MicroHandlerSet = wire.NewSet(
	microhandler.NewUserServiceSimple,
)

// UserRepositorySet provides user repository only (for microservice)
var UserRepositorySet = wire.NewSet(
	repository.NewUserRepository,
)

// UserUseCaseSet provides user use case only (for microservice)
var UserUseCaseSet = wire.NewSet(
	usecase.NewUserInteractor,
	wire.Bind(new(usecase.UserUsecase), new(*usecase.UserInteractor)),
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

// InitializeMicroService initializes the microservice with all dependencies
// Note: Only includes user-related dependencies since microservice only handles users
func InitializeMicroService(db *sqlx.DB, cfg *config.Config) *microhandler.UserServiceSimple {
	wire.Build(
		UserRepositorySet,
		UserUseCaseSet,
		MicroHandlerSet,
	)
	return nil
}
