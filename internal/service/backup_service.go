package service

import (
	"akasha/internal/domain"
	"akasha/internal/repository"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type BackupService struct {
	depRepo    *repository.DependencyRepo
	branchRepo *repository.BranchRepo
	db         *gorm.DB
}

func NewBackupService(depRepo *repository.DependencyRepo, branchRepo *repository.BranchRepo, db *gorm.DB) *BackupService {
	return &BackupService{depRepo: depRepo, branchRepo: branchRepo, db: db}
}

type BackupData struct {
	Version      string              `json:"version"`
	ExportedAt   time.Time           `json:"exportedAt"`
	Branches     []domain.Branch     `json:"branches"`
	Dependencies []domain.Dependency `json:"dependencies"`
}

func (s *BackupService) Export() (*BackupData, error) {
	branches, err := s.branchRepo.FindAll()
	if err != nil {
		return nil, err
	}
	deps, err := s.depRepo.FindAll()
	if err != nil {
		return nil, err
	}
	return &BackupData{
		Version:      "1.0",
		ExportedAt:   time.Now(),
		Branches:     branches,
		Dependencies: deps,
	}, nil
}

func (s *BackupService) Restore(data *BackupData) error {
	if data.Version != "1.0" {
		return &ErrRestore{msg: "unsupported backup version: " + data.Version}
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Exec("DELETE FROM dependencies").Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Exec("DELETE FROM branches").Error; err != nil {
		tx.Rollback()
		return err
	}

	for i := range data.Branches {
		b := data.Branches[i]
		if err := tx.Create(&b).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	for i := range data.Dependencies {
		d := data.Dependencies[i]
		if err := tx.Create(&d).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

type ErrRestore struct {
	msg string
}

func (e *ErrRestore) Error() string {
	return e.msg
}

func (s *BackupService) ValidateJSON(raw []byte) (*BackupData, error) {
	var data BackupData
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, &ErrRestore{msg: "invalid JSON: " + err.Error()}
	}
	if data.Version == "" {
		return nil, &ErrRestore{msg: "missing version field"}
	}
	return &data, nil
}
