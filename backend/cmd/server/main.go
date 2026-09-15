package main

import (
	"context"
	"errors"
	"fmt"
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
	// 跨实例启动引导锁：旧库缺少表/编号序列时，多个实例同时启动也只让一个执行建表、
	// 建/对齐序列与种子；其他实例不退出，而是等待锁后复用已就绪的对象继续启动。
	bootErr := repository.WithBootstrapLock(db, logger, func() error {
		if err := db.AutoMigrate(
			&model.User{}, &model.Client{}, &model.Case{}, &model.Document{}, &model.Billing{}, &model.AuditLog{},
		); err != nil {
			return fmt.Errorf("auto migrate failed: %w", err)
		}
		// 创建案件/账单业务编号序列并按已有编号向后对齐：幂等、单调只增，不清库、不回退、不改号。
		if err := repository.NewSequenceRepository(db).EnsureSequences(); err != nil {
			return fmt.Errorf("ensure number sequences failed: %w", err)
		}
		if err := service.NewSeedService(db, logger).Seed(); err != nil {
			return fmt.Errorf("seed failed: %w", err)
		}
		return nil
	})
	if bootErr != nil {
		logger.Error("database bootstrap failed", "error", bootErr.Error())
		os.Exit(1)
	}

	userRepo := repository.NewUserRepository(db)
	clientRepo := repository.NewClientRepository(db)
	caseRepo := repository.NewCaseRepository(db)
	documentRepo := repository.NewDocumentRepository(db)
	billingRepo := repository.NewBillingRepository(db)
	txStore := repository.NewTxStore(db)

	userSvc := service.NewUserService(userRepo, logger)
	clientSvc := service.NewClientService(clientRepo, caseRepo, logger)
	caseSvc := service.NewCaseService(caseRepo, clientRepo, userRepo, txStore, logger)
	documentSvc := service.NewDocumentService(documentRepo, caseRepo, logger)
	billingSvc := service.NewBillingService(billingRepo, txStore, logger)

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
