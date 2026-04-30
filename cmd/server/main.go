package main

import (
	"akasha"
	"akasha/internal/config"
	"akasha/internal/repository"
	"akasha/internal/router"
	"akasha/internal/service"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	db, err := gorm.Open(mysql.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}

	depRepo := repository.NewDependencyRepo(db)
	branchRepo := repository.NewBranchRepo(db)
	depSvc := service.NewDependencyService(depRepo, branchRepo)
	branchSvc := service.NewBranchService(branchRepo, depRepo)
	initSvc := service.NewInitService(db, depRepo, branchRepo)
	backupSvc := service.NewBackupService(depRepo, branchRepo, db)

	if _, err := initSvc.Initialize(); err != nil {
		log.Fatal("failed to initialize database:", err)
	}

	r := router.Setup(depSvc, branchSvc, initSvc, backupSvc, router.RouterConfig{
		Password:     cfg.Admin.Password,
		APIKey:       cfg.APIKey,
		JWTSecret:    cfg.Auth.JWTSecret,
		ExternalHost: cfg.ExternalHost,
		TemplateFS:   akasha.TemplateFS,
		StaticFS:     akasha.StaticFS,
	})

	log.Printf("starting server at %s", cfg.App.Addr())
	if err := r.Run(cfg.App.Addr()); err != nil {
		log.Fatal("failed to start server:", err)
	}
}