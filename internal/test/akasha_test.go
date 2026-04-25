package test

import (
	"akasha/internal/domain"
	"akasha/internal/repository"
	"os"
	"testing"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func getTestDB(t *testing.T) *gorm.DB {
	host := os.Getenv("TEST_MYSQL_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	dsn := "root:root123@tcp(" + host + ":3306)/akasha_test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("skipping test: cannot connect to MySQL at %s: %v", host, err)
		return nil
	}
	return db
}

func SetupTestDB(t *testing.T) *gorm.DB {
	db := getTestDB(t)
	if db == nil {
		t.Skip("MySQL not available")
		return nil
	}
	db.Exec("DROP TABLE IF EXISTS dependencies")
	db.Exec("DROP TABLE IF EXISTS branches")
	db.AutoMigrate(&domain.Dependency{}, &domain.Branch{})
	return db
}

func TestBranchCreate(t *testing.T) {
	db := SetupTestDB(t)
	if db == nil {
		return
	}
	branchRepo := repository.NewBranchRepo(db)

	err := branchRepo.CreateBranch("main", "")
	if err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}

	exists := branchRepo.Exists("main")
	if !exists {
		t.Error("expected branch to exist")
	}
}

func TestDependencyCRUD(t *testing.T) {
	db := SetupTestDB(t)
	if db == nil {
		return
	}
	branchRepo := repository.NewBranchRepo(db)
	depRepo := repository.NewDependencyRepo(db)

	branchRepo.CreateBranch("main", "")

	dep := &domain.Dependency{
		Name:     "spring-core",
		GroupID:  "org.springframework",
		Artifact: "spring-core",
		Version:  "6.2.7",
		Branch:   "main",
	}
	if err := depRepo.Create(dep); err != nil {
		t.Fatalf("failed to create dependency: %v", err)
	}

	found, err := depRepo.FindByNameAndBranch("spring-core", "main")
	if err != nil || found == nil {
		t.Fatalf("failed to find dependency: %v", err)
	}
	if found.Version != "6.2.7" {
		t.Errorf("expected version 6.2.7, got %s", found.Version)
	}
}

func TestVersionUpdateCreatesNewRecord(t *testing.T) {
	db := SetupTestDB(t)
	if db == nil {
		return
	}
	branchRepo := repository.NewBranchRepo(db)
	depRepo := repository.NewDependencyRepo(db)

	branchRepo.CreateBranch("main", "")

	dep1 := &domain.Dependency{
		Name: "spring-core", GroupID: "org.springframework",
		Artifact: "spring-core", Version: "6.2.7", Branch: "main",
		CreatedAt: time.Now(),
	}
	depRepo.Create(dep1)

	dep2 := &domain.Dependency{
		Name: "spring-core", GroupID: "org.springframework",
		Artifact: "spring-core", Version: "6.2.8", Branch: "main",
		CreatedAt: time.Now(),
	}
	depRepo.Create(dep2)

	latest, _ := depRepo.FindByNameAndBranch("spring-core", "main")
	if latest.Version != "6.2.8" {
		t.Errorf("expected latest version 6.2.8, got %s", latest.Version)
	}

	history, _ := depRepo.FindHistory("spring-core", "main")
	if len(history) != 2 {
		t.Errorf("expected 2 history records, got %d", len(history))
	}
}

func TestBranchCopy(t *testing.T) {
	db := SetupTestDB(t)
	if db == nil {
		return
	}
	branchRepo := repository.NewBranchRepo(db)
	depRepo := repository.NewDependencyRepo(db)

	branchRepo.CreateBranch("main", "")
	depRepo.Create(&domain.Dependency{
		Name: "spring-core", GroupID: "org.springframework",
		Artifact: "spring-core", Version: "6.2.7", Branch: "main",
		CreatedAt: time.Now(),
	})

	err := branchRepo.CreateBranch("202603", "main")
	if err != nil {
		t.Fatalf("failed to create branch from base: %v", err)
	}

	deps, _ := depRepo.FindLatestByBranch("202603")
	if len(deps) != 1 {
		t.Errorf("expected 1 dependency in new branch, got %d", len(deps))
	}
	if deps[0].Name != "spring-core" {
		t.Errorf("expected spring-core, got %s", deps[0].Name)
	}
}

func TestDeleteInActiveBranchForbidden(t *testing.T) {
	db := SetupTestDB(t)
	if db == nil {
		return
	}
	branchRepo := repository.NewBranchRepo(db)

	branchRepo.CreateBranch("main", "")

	branch, _ := branchRepo.FindByName("main")
	if branch.Status != "active" {
		t.Errorf("expected active status, got %s", branch.Status)
	}
}