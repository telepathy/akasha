package repository

import (
	"akasha/internal/domain"
	"time"

	"gorm.io/gorm"
)

type BranchRepo struct {
	db *gorm.DB
}

func NewBranchRepo(db *gorm.DB) *BranchRepo {
	return &BranchRepo{db: db}
}

func (r *BranchRepo) DB() *gorm.DB {
	return r.db
}

func (r *BranchRepo) Create(b *domain.Branch) error {
	return r.db.Create(b).Error
}

func (r *BranchRepo) FindAll() ([]domain.Branch, error) {
	var branches []domain.Branch
	// Include active and archived, exclude deleted
	err := r.db.Where("status IN (?, ?)", "active", "archived").Order("created_at DESC").Find(&branches).Error
	return branches, err
}

func (r *BranchRepo) FindByName(name string) (*domain.Branch, error) {
	var branch domain.Branch
	err := r.db.Where("name = ? AND status != ?", name, "deleted").First(&branch).Error
	if err != nil {
		return nil, err
	}
	return &branch, nil
}

func (r *BranchRepo) Exists(name string) bool {
	var count int64
	r.db.Model(&domain.Branch{}).Where("name = ? AND status != ?", name, "deleted").Count(&count)
	return count > 0
}

func (r *BranchRepo) UpdateStatus(name, status string) error {
	return r.db.Model(&domain.Branch{}).
		Where("name = ?", name).
		Update("status", status).Error
}

func (r *BranchRepo) Delete(name string) error {
	return r.db.Where("name = ?", name).Delete(&domain.Branch{}).Error
}

func (r *BranchRepo) CreateBranch(name, baseBranch string) error {
	tx := r.db.Begin()

	newBranch := &domain.Branch{
		Name:       name,
		BaseBranch: baseBranch,
		Status:     "creating",
		CreatedAt:  time.Now(),
	}
	if err := tx.Create(newBranch).Error; err != nil {
		tx.Rollback()
		return err
	}

	if baseBranch != "" {
		depRepo := NewDependencyRepo(r.db)
		if err := depRepo.CopyToBranch(baseBranch, name); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Model(&domain.Branch{}).
		Where("name = ?", name).
		Update("status", "active").Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}