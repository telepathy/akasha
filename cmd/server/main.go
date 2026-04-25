package main

import (
	"akasha/internal/config"
	"akasha/internal/domain"
	"akasha/internal/repository"
	"akasha/internal/router"
	"akasha/internal/service"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

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

	if err := db.AutoMigrate(&domain.Dependency{}, &domain.Branch{}); err != nil {
		log.Fatal("failed to migrate:", err)
	}

	depRepo := repository.NewDependencyRepo(db)
	branchRepo := repository.NewBranchRepo(db)
	depSvc := service.NewDependencyService(depRepo, branchRepo)
	branchSvc := service.NewBranchService(branchRepo, depRepo)

	// check if branches table is empty
	var branchCount int64
	db.Model(&domain.Branch{}).Count(&branchCount)
	if branchCount == 0 {
		log.Println("creating main branch...")
		if err := branchRepo.CreateBranch("main", ""); err != nil {
			log.Printf("warning: failed to create main branch: %v", err)
		}
	}

	// check if dependencies table is empty
	var depCount int64
	db.Model(&domain.Dependency{}).Count(&depCount)
	if depCount == 0 {
		log.Println("importing data from dependency.gradle...")
		if err := importFromGradleFile(db); err != nil {
			log.Printf("import warning: %v", err)
		}
	}

	r := router.Setup(depSvc, branchSvc, cfg.Gradle.Password)

	log.Printf("starting server at %s", cfg.App.Addr())
	if err := r.Run(cfg.App.Addr()); err != nil {
		log.Fatal("failed to start server:", err)
	}
}

func importFromGradleFile(db *gorm.DB) error {
	if _, err := os.Stat("dependency.gradle"); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile("dependency.gradle")
	if err != nil {
		return err
	}

	deps := parseDependencyGradle(string(data), "main")
	log.Printf("parsed %d dependencies from gradle file", len(deps))

	for i := range deps {
		if err := db.Create(&deps[i]).Error; err != nil {
			log.Printf("warning: failed to import %s: %v", deps[i].Name, err)
		}
	}

	log.Println("seed import completed")
	return nil
}

func parseDependencyGradle(content, branch string) []domain.Dependency {
	var deps []domain.Dependency
	// Simple pattern: "name" : "group:artifact:version"
	re := regexp.MustCompile(`"([^"]+)"\s*:\s*"([^"]+)"`)

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip comments and empty lines
		if strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*") ||
		   strings.HasPrefix(line, "ext.libraries") || line == "" {
			continue
		}
		// Extract name and maven coord
		matches := re.FindStringSubmatch(line)
		if len(matches) == 3 {
			name := matches[1]
			mavenCoord := matches[2]
			parts := strings.Split(mavenCoord, ":")
			if len(parts) >= 3 {
				dep := domain.Dependency{
					Name:     name,
					Branch:   branch,
					GroupID:  parts[0],
					Artifact: parts[1],
					Version:  parts[2],
					CreatedAt: time.Now(),
				}
				deps = append(deps, dep)
			}
		}
	}
	return deps
}