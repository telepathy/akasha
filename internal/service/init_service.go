package service

import (
	"akasha/internal/domain"
	"akasha/internal/repository"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
)

type InitService struct {
	db         *gorm.DB
	depRepo    *repository.DependencyRepo
	branchRepo *repository.BranchRepo
}

func NewInitService(db *gorm.DB, depRepo *repository.DependencyRepo, branchRepo *repository.BranchRepo) *InitService {
	return &InitService{
		db:         db,
		depRepo:    depRepo,
		branchRepo: branchRepo,
	}
}

type DBStatus struct {
	Initialized      bool     `json:"initialized"`
	Tables           []string `json:"tables"`
	MainBranchExists bool     `json:"mainBranchExists"`
	DependencyCount  int64    `json:"dependencyCount"`
}

func (s *InitService) CheckDBStatus() (*DBStatus, error) {
	status := &DBStatus{
		Tables: []string{},
	}

	tables := []string{"dependencies", "branches"}
	for _, table := range tables {
		if s.tableExists(table) {
			status.Tables = append(status.Tables, table)
		}
	}

	status.Initialized = len(status.Tables) == len(tables)

	if status.Initialized {
		var branchCount int64
		s.db.Model(&domain.Branch{}).Where("name = ?", "main").Count(&branchCount)
		status.MainBranchExists = branchCount > 0

		s.db.Model(&domain.Dependency{}).Count(&status.DependencyCount)
	}

	return status, nil
}

func (s *InitService) Initialize() (*DBStatus, error) {
	if err := s.db.AutoMigrate(&domain.Dependency{}, &domain.Branch{}); err != nil {
		return nil, err
	}

	var branchCount int64
	s.db.Model(&domain.Branch{}).Count(&branchCount)
	if branchCount == 0 {
		log.Println("creating main branch...")
		if err := s.branchRepo.CreateBranch("main", ""); err != nil {
			log.Printf("warning: failed to create main branch: %v", err)
		}
	}

	var depCount int64
	s.db.Model(&domain.Dependency{}).Count(&depCount)
	if depCount == 0 {
		log.Println("importing data from dependency.gradle...")
		if err := s.importFromGradleFile(); err != nil {
			log.Printf("import warning: %v", err)
		}
	}

	return s.CheckDBStatus()
}

func (s *InitService) tableExists(tableName string) bool {
	var count int64
	err := s.db.Raw(
		"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?",
		tableName,
	).Scan(&count).Error
	if err != nil {
		return false
	}
	return count > 0
}

func (s *InitService) importFromGradleFile() error {
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
		if err := s.db.Create(&deps[i]).Error; err != nil {
			log.Printf("warning: failed to import %s: %v", deps[i].Name, err)
		}
	}

	log.Println("seed import completed")
	return nil
}

func parseDependencyGradle(content, branch string) []domain.Dependency {
	var deps []domain.Dependency
	re := regexp.MustCompile(`"([^"]+)"\s*:\s*"([^"]+)"`)

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*") ||
		   strings.HasPrefix(line, "ext.libraries") || line == "" {
			continue
		}
		matches := re.FindStringSubmatch(line)
		if len(matches) == 3 {
			name := matches[1]
			mavenCoord := matches[2]
			parts := strings.Split(mavenCoord, ":")
			if len(parts) >= 3 {
				dep := domain.Dependency{
					Name:      name,
					Branch:    branch,
					GroupID:   parts[0],
					Artifact:  parts[1],
					Version:   parts[2],
					CreatedAt: time.Now(),
				}
				deps = append(deps, dep)
			}
		}
	}
	return deps
}