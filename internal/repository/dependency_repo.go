package repository

import (
	"akasha/internal/domain"
	"time"

	"gorm.io/gorm"
)

type DependencyRepo struct {
	db *gorm.DB
}

func NewDependencyRepo(db *gorm.DB) *DependencyRepo {
	return &DependencyRepo{db: db}
}

func (r *DependencyRepo) Create(d *domain.Dependency) error {
	return r.db.Create(d).Error
}

func (r *DependencyRepo) FindLatestByBranch(branch string) ([]domain.Dependency, error) {
	var deps []domain.Dependency
	sub := r.db.Model(&domain.Dependency{}).
		Select("MAX(id) as id").
		Where("branch = ? AND deleted_at IS NULL", branch).
		Group("name")

	err := r.db.Where("id IN (?)", sub).Find(&deps).Error
	return deps, err
}

func (r *DependencyRepo) FindByNameAndBranch(name, branch string) (*domain.Dependency, error) {
	var dep domain.Dependency
	err := r.db.Where("name = ? AND branch = ? AND deleted_at IS NULL", name, branch).
		Order("created_at DESC").
		First(&dep).Error
	if err != nil {
		return nil, err
	}
	return &dep, nil
}

func (r *DependencyRepo) FindHistory(name, branch string) ([]domain.Dependency, error) {
	var deps []domain.Dependency
	err := r.db.Where("name = ? AND branch = ?", name, branch).
		Order("created_at DESC").
		Find(&deps).Error
	return deps, err
}

// FindHistoryBetween 查询某时间段内的变更
func (r *DependencyRepo) FindHistoryBetween(name, branch string, startAt, endAt time.Time) ([]domain.Dependency, error) {
	var deps []domain.Dependency
	err := r.db.Where("name = ? AND branch = ? AND created_at >= ? AND created_at <= ?", name, branch, startAt, endAt).
		Order("created_at DESC").
		Find(&deps).Error
	return deps, err
}

func (r *DependencyRepo) FindAt(name, branch string, at time.Time) (*domain.Dependency, error) {
	var dep domain.Dependency
	err := r.db.Where("name = ? AND branch = ? AND created_at <= ?", name, branch, at).
		Order("created_at DESC").
		First(&dep).Error
	if err != nil {
		return nil, err
	}
	return &dep, nil
}

// FindDepsAt 批量查询某分支在某时间点的所有依赖
func (r *DependencyRepo) FindDepsAt(branch string, at time.Time) ([]domain.Dependency, error) {
	// 先获取该分支所有不重复的依赖名称
	var names []string
	r.db.Model(&domain.Dependency{}).
		Where("branch = ?", branch).
		Distinct("name").
		Pluck("name", &names)

	var deps []domain.Dependency
	for _, name := range names {
		dep, err := r.FindAt(name, branch, at)
		if err != nil {
			continue
		}
		deps = append(deps, *dep)
	}
	return deps, nil
}

func (r *DependencyRepo) CopyToBranch(fromBranch, toBranch string) error {
	var deps []domain.Dependency
	if err := r.db.Where("branch = ? AND deleted_at IS NULL", fromBranch).Find(&deps).Error; err != nil {
		return err
	}
	for i := range deps {
		deps[i].ID = 0
		deps[i].Branch = toBranch
		deps[i].CreatedAt = time.Now()
	}
	return r.db.Create(&deps).Error
}

func (r *DependencyRepo) Delete(name, branch string) error {
	return r.db.Where("name = ? AND branch = ?", name, branch).
		Delete(&domain.Dependency{}).Error
}