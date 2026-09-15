package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cylawcase/internal/config"
	"cylawcase/internal/handler"
	"cylawcase/internal/model"
	"cylawcase/internal/repository"
	"cylawcase/internal/router"
	"cylawcase/internal/service"
	"cylawcase/internal/util"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	logger := util.NewLogger(slog.LevelInfo)

	db, err := gorm.Open(postgres.Open(cfg.DBDSN()), &gorm.Config{})
	if err != nil {
		logger.Error("connect database failed", "error", err.Error())
		os.Exit(1)
	}
	if err := db.AutoMigrate(
		&model.User{}, &model.Client{}, &model.Case{}, &model.Document{}, &model.Billing{}, &model.AuditLog{},
	); err != nil {
		logger.Error("auto migrate failed", "error", err.Error())
		os.Exit(1)
	}
	if err := service.NewSeedService(db, logger).Seed(); err != nil {
		logger.Error("seed failed", "error", err.Error())
		os.Exit(1)
	}

	userRepo := repository.NewUserRepository(db)
	clientRepo := repository.NewClientRepository(db)
	caseRepo := repository.NewCaseRepository(db)
	documentRepo := repository.NewDocumentRepository(db)
	billingRepo := repository.NewBillingRepository(db)

	userSvc := service.NewUserService(userRepo, logger)
	clientSvc := service.NewClientService(clientRepo, caseRepo, logger)
	caseSvc := service.NewCaseService(caseRepo, clientRepo, userRepo, logger)
	documentSvc := service.NewDocumentService(documentRepo, caseRepo, logger)
	billingSvc := service.NewBillingService(billingRepo, caseRepo, clientRepo, logger)

	userHandler := handler.NewUserHandler(userSvc, logger)
	clientHandler := handler.NewClientHandler(clientSvc, logger)
	caseHandler := handler.NewCaseHandler(caseSvc, logger)
	documentHandler := handler.NewDocumentHandler(documentSvc, logger)
	billingHandler := handler.NewBillingHandler(billingSvc, logger)
	uploadHandler := handler.NewUploadHandler(cfg, logger)
	auditLogHandler := handler.NewAuditLogHandler(db, logger)

	r := router.New(cfg, db, logger, userHandler, clientHandler, caseHandler,
		documentHandler, billingHandler, uploadHandler, auditLogHandler)

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r.Setup(),
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("server starting", "port", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server run failed", "error", err.Error())
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("server shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", "error", err.Error())
	}
	logger.Info("server stopped")
}
